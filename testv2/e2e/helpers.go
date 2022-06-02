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
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/Masterminds/semver/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	labelControlPlaneNode = "node-role.kubernetes.io/control-plane"
	prowImage             = "kubermatic/kubeone-e2e:v0.1.22"
	k1CloneURI            = "ssh://git@github.com/kubermatic/kubeone.git"
)

func titleize(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, ".", "_")
	s = cases.Title(language.English).String(s)

	return strings.ReplaceAll(s, " ", "")
}

func clusterName() string {
	name, found := os.LookupEnv("BUILD_ID")
	if !found {
		name = rand.String(10)
	}

	return fmt.Sprintf("k1-%s", name)
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

func newKubeoneBin(terraformPath, manifestPath string) *kubeoneBin {
	return &kubeoneBin{
		bin:          "kubeone",
		dir:          terraformPath,
		tfjsonPath:   ".",
		manifestPath: manifestPath,
	}
}

type manifestData struct {
	VERSION string
}

func renderManifest(t *testing.T, templatePath string, data manifestData) string {
	t.Helper()

	tmpDir := t.TempDir()

	var buf bytes.Buffer

	tpl, err := template.New("").Parse(templatePath)
	if err != nil {
		t.Fatal(err)
	}
	tpl.Funcs(template.FuncMap{
		"required": requiredTemplateFunc,
	})

	if err = tpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}

	manifest, err := os.CreateTemp(tmpDir, "kubeone-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer manifest.Close()

	manifestPath := manifest.Name()
	if err := os.WriteFile(manifestPath, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}

	return manifestPath
}

func waitForNodesReady(t *testing.T, client ctrlruntimeclient.Client, expectedNumberOfNodes int) error {
	t.Helper()

	return wait.Poll(5*time.Second, 10*time.Minute, func() (bool, error) {
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
						"go", "test", "-v",
						"./testv2/e2e/...",
						"-tags", "e2e",
						"-run", fmt.Sprintf("^%s$", testTitle),
					},
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
	t.Helper()

	kubeoneManifest, err := k1.Manifest()
	if err != nil {
		t.Fatalf("failed to get manifest API")
	}

	numberOfNodesToWait := len(kubeoneManifest.ControlPlane.Hosts) + len(kubeoneManifest.StaticWorkers.Hosts)
	for _, worker := range kubeoneManifest.DynamicWorkers {
		if worker.Replicas != nil {
			numberOfNodesToWait += *worker.Replicas
		}
	}

	var kubeconfig []byte
	fetchKubeconfig := func() error {
		kubeconfig, err = k1.Kubeconfig()
		if err != nil {
			return err
		}

		return nil
	}

	if err = retryFn(fetchKubeconfig); err != nil {
		t.Fatalf("kubeone kubeconfig failed: %v", err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		t.Fatalf("unable to build clientset from kubeconfig bytes: %v", err)
	}

	client, err := ctrlruntimeclient.New(restConfig, ctrlruntimeclient.Options{})
	if err != nil {
		t.Fatalf("failed to init dynamic client: %s", err)
	}

	if err = waitForNodesReady(t, client, numberOfNodesToWait); err != nil {
		t.Fatalf("failed to bring up all nodes up: %v", err)
	}

	if err = verifyVersion(client, metav1.NamespaceSystem, data.VERSION); err != nil {
		t.Fatalf("version mismatch: %v", err)
	}

}

func sonobuoyRun(t *testing.T, k1 *kubeoneBin, mode sonobuoyMode) {
	kubeconfigPath, err := k1.KubeconfigPath(t.TempDir())
	if err != nil {
		t.Fatalf("fetching kubeconfig failed")
	}

	sb := sonobuoyBin{
		kubeconfig: kubeconfigPath,
	}

	if err := sb.Run(mode); err != nil {
		t.Fatalf("sonobuoy run failed: %v", err)
	}

	if err := sb.Wait(); err != nil {
		t.Fatalf("sonobuoy wait failed: %v", err)
	}

	if err := sb.Retrieve(); err != nil {
		t.Fatalf("sonobuoy retrieve failed: %v", err)
	}
}
