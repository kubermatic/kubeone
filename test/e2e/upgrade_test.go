// +build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	corev1types "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	labelControlPlaneNode = "node-role.kubernetes.io/master"
)

func TestClusterUpgrade(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name              string
		provider          string
		initialVersion    string
		targetVersion     string
		initialConfigPath string
		targetConfigPath  string
		scenario          string
	}{
		{
			name:              "verify k8s cluster deployment on AWS",
			provider:          AWS,
			initialVersion:    "v1.13.1",
			targetVersion:     "v1.13.3",
			initialConfigPath: "../../test/e2e/testdata/upgrades_aws_1.13.1.yaml",
			targetConfigPath:  "../../test/e2e/testdata/upgrades_aws_1.13.3.yaml",
			scenario:          NodeConformance,
		},
		{
			name:              "verify k8s cluster deployment on AWS",
			provider:          DigitalOcean,
			initialVersion:    "v1.13.1",
			targetVersion:     "v1.13.3",
			initialConfigPath: "../../test/e2e/testdata/upgrades_do_1.13.1.yaml",
			targetConfigPath:  "../../test/e2e/testdata/upgrades_do_1.13.3.yaml",
			scenario:          NodeConformance,
		},
	}

	for _, tc := range testcases {
		// to satisfy scope linter
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if len(testRunIdentifier) == 0 {
				t.Fatalf("-identifier must be set")
			}
			if testProvider != tc.provider {
				t.SkipNow()
			}
			testPath := fmt.Sprintf("../../_build/%s", testRunIdentifier)

			pr, err := CreateProvisioner(testPath, testRunIdentifier, tc.provider)
			if err != nil {
				t.Fatal(err)
			}
			target := NewKubeone(testPath, tc.initialConfigPath)

			teardown := setupTearDown(pr, target)
			defer teardown(t)

			t.Log("check prerequisites")
			err = ValidateCommon()
			if err != nil {
				t.Fatalf("%v", err)
			}

			t.Log("start provisioning")
			tf, err := pr.Provision()
			if err != nil {
				t.Fatalf("provisioning failed: %v", err)
			}

			t.Log("start cluster deployment")
			err = target.Install(tf)
			if err != nil {
				t.Fatalf("k8s cluster deployment failed: %v", err)
			}

			t.Log("create kubeconfig")
			kubeconfig, err := target.CreateKubeconfig()
			if err != nil {
				t.Fatalf("creating kubeconfig failed: %v", err)
			}

			t.Log("build kubernetes clientset")
			restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
			if err != nil {
				t.Errorf("unable to build config from kubeconfig bytes: %v", err)
			}
			clientset, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				t.Errorf("unable to build kubernetes clientset: %v", err)
			}

			t.Log("waiting for nodes to become ready")
			err = waitForNodesReady(clientset.CoreV1().Nodes())
			if err != nil {
				t.Fatalf("nodes are not ready: %v", err)
			}

			// Create a new KubeOne provisioner pointing to the new configuration file
			target = NewKubeone(testPath, tc.targetConfigPath)
			clusterVerifier := NewKubetest(tc.targetVersion, "../../_build", map[string]string{
				"KUBERNETES_CONFORMANCE_TEST": "y",
			})

			t.Log("start cluster upgrade")
			err = target.Upgrade()
			if err != nil {
				t.Fatalf("k8s cluster upgrade failed: %v", err)
			}

			t.Log("waiting for nodes to become ready")
			err = waitForNodesReady(clientset.CoreV1().Nodes())
			if err != nil {
				t.Fatalf("nodes are not ready: %v", err)
			}

			t.Log("run e2e tests")
			err = clusterVerifier.Verify(tc.scenario)
			if err != nil {
				t.Fatalf("e2e tests failed: %v", err)
			}
		})
	}
}

func waitForNodesReady(nodeClient corev1types.NodeInterface) error {
	return wait.Poll(5*time.Second, 3*time.Minute, func() (bool, error) {
		nodes, err := nodeClient.List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", labelControlPlaneNode, ""),
		})
		if err != nil {
			return false, errors.Wrap(err, "unable to list nodes")
		}
		for _, n := range nodes.Items {
			for _, c := range n.Status.Conditions {
				if c.Type == corev1.NodeReady && c.Status != corev1.ConditionTrue {
					return false, errors.Errorf("node %s is not running", n.ObjectMeta.Name)
				}
			}
		}
		return true, nil
	})
}
