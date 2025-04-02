/*
Copyright 2022 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"testing"
	"text/template"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-logr/logr"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"k8c.io/kubeone/test/testexec"
	clusterv1alpha1 "k8c.io/machine-controller/pkg/apis/cluster/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	labelControlPlaneNode = "node-role.kubernetes.io/control-plane"
	prowjobImage          = "quay.io/kubermatic/build:go-1.24-node-20-0"
	k1CloneURI            = "ssh://git@github.com/kubermatic/kubeone.git"
)

var sonobuoyBackoff = wait.Backoff{
	Steps:    10,
	Duration: 10 * time.Second,
	Factor:   1.0,
	Jitter:   0.1,
}

func init() {
	if err := clusterv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	ctrlruntimelog.SetLogger(logr.Discard())
}

func mustGetwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return cwd
}

func mustAbsolutePath(file string) string {
	if !strings.HasPrefix(file, "/") {
		// find the absolute path to the file
		file = filepath.Join(mustGetwd(), file)
	}

	return filepath.Clean(file)
}

func getKubeoneDistPath() string {
	const distPath = "../../dist/kubeone"

	return mustAbsolutePath(distPath)
}

func titleize(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, ".", "_")
	s = cases.Title(language.English).String(s)

	return strings.ReplaceAll(s, " ", "")
}

func trueRetriable(error) bool {
	return true
}

func retryFnWithBackoff(backoff wait.Backoff, fn func() error) error {
	return retry.OnError(backoff, trueRetriable, fn)
}

func retryFn(fn func() error) error {
	return retryFnWithBackoff(retry.DefaultRetry, fn)
}

func requiredTemplateFunc(warn string, input interface{}) (interface{}, error) {
	switch val := input.(type) {
	case nil:
		return val, errors.New(warn)
	case string:
		if val == "" {
			return val, errors.New(warn)
		}
	}

	return input, nil
}

func makeBin(args ...string) *testexec.Exec {
	return makeBinWithPath(filepath.Clean("../../"), args...)
}

func makeBinWithPath(path string, args ...string) *testexec.Exec {
	return testexec.NewExec("make",
		testexec.WithArgs(args...),
		testexec.WithEnv(os.Environ()),
		testexec.InDir(path),
		testexec.StdoutDebug,
	)
}

func kubeoneStableProwExtraRefs(baseRef string) []ProwRef {
	return []ProwRef{
		{
			Org:       "kubermatic",
			Repo:      "kubeone",
			BaseRef:   baseRef,
			PathAlias: "k8c.io/kubeone-stable",
		},
	}
}

type kubeoneBinOpts func(*kubeoneBin)

func withKubeoneBin(bin string) kubeoneBinOpts {
	return func(kb *kubeoneBin) {
		kb.bin = bin
	}
}

func withKubeoneVerbose(kb *kubeoneBin) {
	kb.verbose = true
}

func withKubeoneCredentials(credentialsPath string) kubeoneBinOpts {
	return func(kb *kubeoneBin) {
		kb.credentialsPath = credentialsPath
	}
}

func newKubeoneBin(terraformPath, manifestPath string, opts ...kubeoneBinOpts) *kubeoneBin {
	k1 := &kubeoneBin{
		bin:          getKubeoneDistPath(),
		dir:          terraformPath,
		tfjsonPath:   ".",
		manifestPath: manifestPath,
	}

	for _, mod := range opts {
		mod(k1)
	}

	return k1
}

type manifestData struct {
	VERSION string
	INFRA   string
}

func renderManifest(t *testing.T, templatePath string, data manifestData) string {
	var (
		outBuf bytes.Buffer
		tmpDir = t.TempDir()
	)

	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatal(err)
	}

	tpl, err := template.New("").
		Funcs(template.FuncMap{
			"required":  requiredTemplateFunc,
			"hasPrefix": strings.HasPrefix,
		}).
		Parse(string(templateContent))
	if err != nil {
		t.Fatal(err)
	}

	if err = tpl.Execute(&outBuf, data); err != nil {
		t.Fatal(err)
	}

	manifest, err := os.CreateTemp(tmpDir, "kubeone-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer manifest.Close()

	manifestPath := manifest.Name()
	if err := os.WriteFile(manifestPath, outBuf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	return manifestPath
}

func waitForNodesReady(ctx context.Context, t *testing.T, client ctrlruntimeclient.Client, expectedNumberOfNodes int) error {
	waitTimeout := 20 * time.Minute
	t.Logf("waiting maximum %s for %d nodes to be ready", waitTimeout, expectedNumberOfNodes)

	return wait.PollUntilContextTimeout(ctx, 5*time.Second, waitTimeout, false, func(ctx context.Context) (bool, error) {
		if err := ctx.Err(); err != nil {
			return false, fmt.Errorf("wait for nodes ready: %w", err)
		}

		nodes := corev1.NodeList{}

		if err := client.List(ctx, &nodes); err != nil {
			t.Logf("error: %v", err)

			return false, nil
		}

		if len(nodes.Items) != expectedNumberOfNodes {
			return false, nil
		}

		for _, n := range nodes.Items {
			for _, c := range n.Status.Conditions {
				if c.Type == corev1.NodeReady && c.Status != corev1.ConditionTrue {
					return false, nil
				}
			}
		}

		return true, nil
	})
}

func verifyVersion(client ctrlruntimeclient.Client, namespace string, targetVersion string) error {
	reqVer, err := semver.NewVersion(targetVersion)
	if err != nil {
		return fmt.Errorf("desired version is invalid: %w", err)
	}

	nodes := corev1.NodeList{}
	nodeListOpts := ctrlruntimeclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			labelControlPlaneNode: "",
		}),
	}

	if err = client.List(context.Background(), &nodes, &nodeListOpts); err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	// Kubelet version check
	for _, n := range nodes.Items {
		kubeletVer, errSemver := semver.NewVersion(n.Status.NodeInfo.KubeletVersion)
		if errSemver != nil {
			return errSemver
		}
		if reqVer.Compare(kubeletVer) != 0 {
			return fmt.Errorf("kubelet version mismatch: expected %v, got %v", reqVer.String(), kubeletVer.String())
		}
	}

	apiserverPods := corev1.PodList{}
	podsListOpts := ctrlruntimeclient.ListOptions{
		Namespace: namespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"component": "kube-apiserver",
		}),
	}

	if err = client.List(context.Background(), &apiserverPods, &podsListOpts); err != nil {
		return fmt.Errorf("unable to list apiserver pods: %w", err)
	}

	for _, p := range apiserverPods.Items {
		apiserverVer, err := parseContainerImageVersion(p.Spec.Containers[0].Image)
		if err != nil {
			return fmt.Errorf("unable to parse apiserver version: %w", err)
		}

		if reqVer.Compare(apiserverVer) != 0 {
			return fmt.Errorf("apiserver version mismatch: expected %v, got %v", reqVer.String(), apiserverVer.String())
		}
	}

	return nil
}

func parseContainerImageVersion(image string) (*semver.Version, error) {
	ver := strings.Split(image, ":")
	if len(ver) != 2 {
		return nil, fmt.Errorf("invalid container image format: %s", image)
	}

	return semver.NewVersion(ver[1])
}

type ProwJob struct {
	Name             string                `json:"name"`
	AlwaysRun        bool                  `json:"always_run"`
	RunIfChanged     string                `json:"run_if_changed,omitempty"`
	Optional         bool                  `json:"optional"`
	Decorate         bool                  `json:"decorate"`
	DecorationConfig *ProwDecorationConfig `json:"decoration_config,omitempty"`
	CloneURI         string                `json:"clone_uri"`
	PathAlias        string                `json:"path_alias,omitempty"`
	Labels           map[string]string     `json:"labels,omitempty"`
	ExtraRefs        []ProwRef             `json:"extra_refs,omitempty"`
	Spec             *corev1.PodSpec       `json:"spec"`
}

type ProwRef struct {
	Org       string `json:"org"`
	Repo      string `json:"repo"`
	BaseRef   string `json:"base_ref,omitempty"`
	PathAlias string `json:"path_alias,omitempty"`
}

type ProwDecorationConfig struct {
	Timeout string `json:"timeout,omitempty"`
}

func newProwJob(prowJobName string, labels map[string]string, testTitle string, settings ProwConfig, extraRefs []ProwRef) ProwJob {
	var env []corev1.EnvVar

	for k, v := range settings.Environ {
		env = append(env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	sort.Slice(env, func(i, j int) bool {
		return env[i].Name < env[j].Name
	})

	return ProwJob{
		Name:         prowJobName,
		AlwaysRun:    settings.AlwaysRun,
		RunIfChanged: settings.RunIfChanged,
		Optional:     settings.Optional,
		Decorate:     true,
		DecorationConfig: &ProwDecorationConfig{
			Timeout: "210m",
		},
		CloneURI:  k1CloneURI,
		Labels:    labels,
		ExtraRefs: extraRefs,
		PathAlias: "k8c.io/kubeone",
		Spec: &corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image:           prowjobImage,
					ImagePullPolicy: corev1.PullAlways,
					Command: []string{
						"./test/go-test-e2e.sh",
						testTitle,
					},
					Env: env,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("1"),
						},
					},
				},
			},
		},
	}
}

func PullProwJobName(in ...string) string {
	return fmt.Sprintf("pull-kubeone-e2e-%s", strings.ReplaceAll(strings.Join(in, "-"), "_", "-"))
}

func dynamicClientRetriable(t *testing.T, k1 *kubeoneBin) ctrlruntimeclient.Client {
	var (
		client ctrlruntimeclient.Client
		err    error
	)

	err = retryFn(func() error {
		client, err = k1.DynamicClient()

		return err
	})
	if err != nil {
		t.Fatalf("initializing dynamic client: %s", err)
	}

	return client
}

func latestUpstreamVersion(majorMinorVersion string) (string, error) {
	majorMinorSemver := semver.MustParse(majorMinorVersion)

	const urlTemplate = "https://dl.k8s.io/release/stable-%d.%d.txt"
	downloadURL := fmt.Sprintf(urlTemplate, majorMinorSemver.Major(), majorMinorSemver.Minor())

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("building http request to download kubeone: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request to download kubeone: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading version marker http response: %w", err)
	}

	return strings.TrimSpace(string(body)), nil
}

func labelNodesSkipEviction(t *testing.T, client ctrlruntimeclient.Client) {
	ctx := context.Background()

	var nodeList corev1.NodeList

	err := retryFn(func() error {
		return client.List(ctx, &nodeList, ctrlruntimeclient.HasLabels{"machine-controller/owned-by"})
	})
	if err != nil {
		t.Fatalf("listing nodes: %v", err)
	}

	for _, node := range nodeList.Items {
		nodeOld := node.DeepCopy()
		nodeNew := node

		if nodeNew.Annotations == nil {
			nodeNew.Annotations = map[string]string{}
		}
		nodeNew.Annotations["kubermatic.io/skip-eviction"] = "true"

		err = retryFn(func() error {
			return client.Patch(context.Background(), &nodeNew, ctrlruntimeclient.MergeFrom(nodeOld))
		})
		if err != nil {
			t.Fatalf("patching node %q to skip eviction: %v", node.Name, err)
		}
	}
}

func waitMachinesHasNodes(t *testing.T, k1 *kubeoneBin, client ctrlruntimeclient.Client) {
	ctx := context.Background()

	kubeoneManifest, err := k1.ClusterManifest()
	if err != nil {
		t.Fatalf("rendering cluster manifest: %v", err)
	}

	numberOfMachinesToWait := 0
	for _, worker := range kubeoneManifest.DynamicWorkers {
		if worker.Replicas != nil {
			numberOfMachinesToWait += *worker.Replicas
		}
	}

	waitErr := wait.PollUntilContextTimeout(ctx, 15*time.Second, 30*time.Minute, false, func(ctx context.Context) (bool, error) {
		var (
			machineList              clusterv1alpha1.MachineList
			someMachinesLacksTheNode bool
		)

		err := retryFn(func() error {
			return client.List(ctx, &machineList, ctrlruntimeclient.InNamespace(metav1.NamespaceSystem))
		})

		if len(machineList.Items) != numberOfMachinesToWait {
			t.Logf("Found %d Machines, but expected %d...", len(machineList.Items), numberOfMachinesToWait)

			return false, nil
		}

		t.Logf("checking %d machines for node reference", len(machineList.Items))
		for _, machine := range machineList.Items {
			if machine.DeletionTimestamp != nil {
				t.Logf("machine %q is being deleted", machine.Name)
				someMachinesLacksTheNode = true
			}
			if machine.Status.NodeRef == nil {
				t.Logf("machine %q still has no nodeRef", machine.Name)
				someMachinesLacksTheNode = true
			}
		}

		return !someMachinesLacksTheNode, err
	})
	if waitErr != nil {
		t.Fatalf("waiting for machines to get nodes references: %v", waitErr)
	}
}

func waitKubeOneNodesReady(ctx context.Context, t *testing.T, k1 *kubeoneBin) {
	client := dynamicClientRetriable(t, k1)

	kubeoneManifest, err := k1.ClusterManifest()
	if err != nil {
		t.Fatalf("rendering cluster manifest: %v", err)
	}

	numberOfNodesToWait := len(kubeoneManifest.ControlPlane.Hosts) + len(kubeoneManifest.StaticWorkers.Hosts)
	for _, worker := range kubeoneManifest.DynamicWorkers {
		if worker.Replicas != nil {
			numberOfNodesToWait += *worker.Replicas
		}
	}

	if err = waitForNodesReady(ctx, t, client, numberOfNodesToWait); err != nil {
		t.Fatalf("waiting %d nodes to be Ready: %v", numberOfNodesToWait, err)
	}

	if err = verifyVersion(client, metav1.NamespaceSystem, kubeoneManifest.Versions.Kubernetes); err != nil {
		t.Fatalf("version mismatch: %v", err)
	}
}

func sonobuoyRun(ctx context.Context, t *testing.T, k1 *kubeoneBin, mode sonobuoyMode, proxyURL string) {
	sonobuoyRunWithRunCount(ctx, t, k1, mode, t.TempDir(), 0, proxyURL)
}

func sonobuoyRunWithRunCount(ctx context.Context, t *testing.T, k1 *kubeoneBin, mode sonobuoyMode, testDir string, runCount int, proxyURL string) {
	kubeconfigPath, err := k1.kubeconfigPath(t.TempDir())
	if err != nil {
		t.Fatalf("fetching kubeconfig failed")
	}

	sb := sonobuoyBin{
		kubeconfig: kubeconfigPath,
		dir:        testDir,
		proxyURL:   proxyURL,
	}

	rerunFailed := runCount > 0

	if err = retryFnWithBackoff(sonobuoyBackoff, func() error { return sb.Run(ctx, mode, rerunFailed) }); err != nil {
		t.Fatalf("sonobuoy run failed: %v", err)
	}

	if err = retryFnWithBackoff(sonobuoyBackoff, func() error { return sb.Wait(ctx) }); err != nil {
		t.Fatalf("sonobuoy wait failed: %v", err)
	}

	err = retryFnWithBackoff(sonobuoyBackoff, func() error { return sb.Retrieve(ctx) })
	if err != nil {
		t.Fatalf("sonobuoy retrieve failed: %v", err)
	}

	var report []sonobuoyReport
	err = retryFnWithBackoff(sonobuoyBackoff, func() error {
		report, err = sb.Results(ctx)

		return err
	})
	if err != nil {
		t.Fatalf("sonobuoy results failed: %v", err)
	}

	if len(report) > 0 {
		var buf strings.Builder

		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		if err = enc.Encode(report); err != nil {
			t.Errorf("failed to json encode sonobuoy report: %v", err)
		}

		if runCount < sonobuoyRunRetries {
			t.Logf("some e2e tests failed:\n%s", buf.String())

			t.Logf("deleting previous sonobuoy run...")
			if err = retryFnWithBackoff(sonobuoyBackoff, func() error { return sb.Delete(ctx) }); err != nil {
				t.Fatalf("sonobuoy delete failed: %v", err)
			}

			t.Logf("restarting failed e2e tests (try %d/%d)...", runCount+1, sonobuoyRunRetries)
			sonobuoyRunWithRunCount(ctx, t, k1, mode, testDir, runCount+1, proxyURL)
		} else {
			t.Fatalf("some e2e tests failed:\n%s", buf.String())
		}
	}
}

func NewSignalContext(logger func(format string, args ...any)) context.Context {
	// We use context.WithTimeout because we want to cancel the context
	// before test timeouts. If we allow the test to timeout, the main test
	// process will terminate, but other subprocesses that we call will
	// remain running. For example, we run `sonobuoy wait`-- if it eventually
	// gets stuck, once the test timeouts, the main test process will terminate,
	// but `sonobuoy wait` will continue running until the ProwJob doesn't
	// timeout. This also means, because the main process has been terminated,
	// there will be NO cleanup, so we'll leak resources.
	testTimeout := 120 * time.Minute
	if timeout, ok := os.LookupEnv("TEST_TIMEOUT"); ok {
		if parsedDuration, err := time.ParseDuration(timeout); err == nil {
			testTimeout = parsedDuration
		} else {
			logger("WARNING: failed to parse %q TEST_TIMEOUT from env: %v", timeout, err)
			logger("WARNING: defaulting TEST_TIMEOUT to %s", testTimeout)
		}
	}

	// We allow 5 minutes for tests to clean up after themselves.
	testTimeout -= 5 * time.Minute
	ctx, _ := context.WithTimeout(context.Background(), testTimeout) //nolint:govet
	sCtx, _ := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	return sCtx
}
