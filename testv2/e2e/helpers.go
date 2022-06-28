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
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/Masterminds/semver/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/test/e2e/testutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	k8spath "k8s.io/utils/path"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	labelControlPlaneNode = "node-role.kubernetes.io/control-plane"
	prowImage             = "kubermatic/kubeone-e2e:v0.1.23"
	k1CloneURI            = "ssh://git@github.com/kubermatic/kubeone.git"
)

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

func retryFn(fn func() error) error {
	return retry.OnError(retry.DefaultRetry, trueRetriable, fn)
}

func requiredTemplateFunc(warn string, input interface{}) (interface{}, error) {
	switch val := input.(type) {
	case nil:
		return val, fmt.Errorf(warn)
	case string:
		if val == "" {
			return val, fmt.Errorf(warn)
		}
	}

	return input, nil
}

func makeBin(args ...string) *testutil.Exec {
	return testutil.NewExec("make",
		testutil.WithArgs(args...),
		testutil.WithEnv(os.Environ()),
		testutil.InDir(filepath.Clean("../../")),
		testutil.StdoutDebug,
	)
}

func downloadKubeone(t *testing.T, version string) string {
	binPath := filepath.Join(t.TempDir(), fmt.Sprintf("kubeone-%s", version))
	zipPath := fmt.Sprintf("%s.zip", binPath)

	exists, err := k8spath.Exists(k8spath.CheckSymlinkOnly, binPath)
	if err != nil {
		t.Fatalf("checking if kubeone already downloaded: %v", err)
	}

	if exists {
		return binPath
	}

	const urlTemplate = "https://github.com/kubermatic/kubeone/releases/download/v%s/kubeone_%s_linux_amd64.zip"
	downloadURL := fmt.Sprintf(urlTemplate, version, version)

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		t.Fatalf("building http request to download kubeone: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http request to download kubeone: %v", err)
	}
	defer resp.Body.Close()

	zipBin, err := os.OpenFile(zipPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("open kubeone destination file: %v", err)
	}
	defer zipBin.Close()

	_, err = io.Copy(zipBin, resp.Body)
	if err != nil {
		t.Fatalf("downloading kubeone: %v", err)
	}

	fi, err := zipBin.Stat()
	if err != nil {
		t.Fatalf("file stat: %v", err)
	}

	unzip, err := zip.NewReader(zipBin, fi.Size())
	if err != nil {
		t.Fatalf("opening zip file for reading: %v", err)
	}

	unzipK1Bin, err := unzip.Open("kubeone")
	if err != nil {
		t.Fatalf("opening kubeone file from zip archive: %v", err)
	}
	defer unzipK1Bin.Close()

	k1Bin, err := os.OpenFile(binPath, os.O_CREATE|os.O_WRONLY, 0750)
	if err != nil {
		t.Fatalf("open kubeone destination file: %v", err)
	}
	defer k1Bin.Close()

	_, err = io.Copy(k1Bin, unzipK1Bin)
	if err != nil {
		t.Fatalf("extracting kubeone from zip: %v", err)
	}

	return binPath
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
			"required": requiredTemplateFunc,
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
	if err := os.WriteFile(manifestPath, outBuf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}

	return manifestPath
}

func waitForNodesReady(t *testing.T, client ctrlruntimeclient.Client, expectedNumberOfNodes int) error {
	waitTimeout := 10 * time.Minute
	t.Logf("waiting maximum %s for %d nodes to be ready", waitTimeout, expectedNumberOfNodes)

	return wait.Poll(5*time.Second, waitTimeout, func() (bool, error) {
		nodes := corev1.NodeList{}

		if err := client.List(context.Background(), &nodes); err != nil {
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
	Name      string            `json:"name"`
	AlwaysRun bool              `json:"always_run"`
	Optional  bool              `json:"optional"`
	Decorate  bool              `json:"decorate"`
	CloneURI  string            `json:"clone_uri"`
	PathAlias string            `json:"path_alias,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Spec      *corev1.PodSpec   `json:"spec"`
}

func newProwJob(prowJobName string, labels map[string]string, testTitle string, settings ProwConfig) ProwJob {
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
		Name:      prowJobName,
		AlwaysRun: settings.AlwaysRun,
		Optional:  settings.Optional,
		Decorate:  true,
		CloneURI:  k1CloneURI,
		Labels:    labels,
		Spec: &corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image:           prowImage,
					ImagePullPolicy: corev1.PullAlways,
					Command: []string{
						"./testv2/go-test-e2e.sh",
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

func pullProwJobName(in ...string) string {
	return fmt.Sprintf("pull-kubeone-e2e-%s", strings.ReplaceAll(strings.Join(in, "-"), "_", "-"))
}

func basicTest(t *testing.T, k1 *kubeoneBin, data manifestData) {
	var (
		kubeoneManifest *kubeoneapi.KubeOneCluster
		err             error
		kubeconfig      []byte
		restConfig      *rest.Config
	)

	fetchKubeoneManifest := func() error {
		kubeoneManifest, err = k1.ClusterManifest()

		return err
	}

	if err = retryFn(fetchKubeoneManifest); err != nil {
		t.Fatalf("failed to get manifest API: %v", err)
	}

	fetchKubeconfig := func() error {
		kubeconfig, err = k1.Kubeconfig()

		return err
	}

	if err = retryFn(fetchKubeconfig); err != nil {
		t.Fatalf("kubeone kubeconfig failed: %v", err)
	}

	initKubeRestConfig := func() error {
		restConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig)

		return err
	}

	if err = retryFn(initKubeRestConfig); err != nil {
		t.Fatalf("unable to build clientset from kubeconfig bytes: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connector := ssh.NewConnector(ctx)
	tun, err := connector.Tunnel(kubeoneManifest.RandomHost())
	if err != nil {
		t.Fatalf("creating SSH tunnel: %v", err)
	}

	restConfig.Dial = tun.TunnelTo

	client, err := ctrlruntimeclient.New(restConfig, ctrlruntimeclient.Options{})
	if err != nil {
		t.Fatalf("failed to init dynamic client: %s", err)
	}

	numberOfNodesToWait := len(kubeoneManifest.ControlPlane.Hosts) + len(kubeoneManifest.StaticWorkers.Hosts)
	for _, worker := range kubeoneManifest.DynamicWorkers {
		if worker.Replicas != nil {
			numberOfNodesToWait += *worker.Replicas
		}
	}

	if err = waitForNodesReady(t, client, numberOfNodesToWait); err != nil {
		t.Fatalf("failed to bring up all nodes up: %v", err)
	}

	if err = verifyVersion(client, metav1.NamespaceSystem, data.VERSION); err != nil {
		t.Fatalf("version mismatch: %v", err)
	}
}

func sonobuoyRun(t *testing.T, k1 *kubeoneBin, mode sonobuoyMode) {
	kubeconfigPath, err := k1.kubeconfigPath(t.TempDir())
	if err != nil {
		t.Fatalf("fetching kubeconfig failed")
	}

	// launch kubeone proxy, to have a HTTPS proxy through the SSH tunnel
	// to open access to the kubeapi behind the bastion host
	proxyCtx, killProxy := context.WithCancel(context.Background())
	proxyURL, waitK1, err := k1.AsyncProxy(proxyCtx)
	if err != nil {
		t.Fatalf("starting kubeone proxy: %v", err)
	}
	defer func() {
		waitErr := waitK1()
		if waitErr != nil {
			t.Logf("wait kubeone proxy: %v", waitErr)
		}
	}()
	defer killProxy()

	t.Logf("kubeone proxy is running on %s", proxyURL)

	// let kubeone proxy start and open the port
	time.Sleep(5 * time.Second)

	sb := sonobuoyBin{
		kubeconfig: kubeconfigPath,
		dir:        t.TempDir(),
		proxyURL:   proxyURL,
	}

	if err = sb.Run(mode); err != nil {
		t.Fatalf("sonobuoy run failed: %v", err)
	}

	if err = sb.Wait(); err != nil {
		t.Fatalf("sonobuoy wait failed: %v", err)
	}

	err = retryFn(sb.Retrieve)
	if err != nil {
		t.Fatalf("sonobuoy retrieve failed: %v", err)
	}

	report, err := sb.Results()
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
		t.Fatalf("some e2e tests failed:\n%s", buf.String())
	}
}
