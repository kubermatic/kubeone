// +build e2e

/*
Copyright 2019 The KubeOne Authors.

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
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
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
			name:              "upgrade k8s 1.13.5 cluster to 1.14.0 on AWS",
			provider:          AWS,
			initialVersion:    "v1.13.5",
			targetVersion:     "v1.14.0",
			initialConfigPath: "../../test/e2e/testdata/config_aws_1.13.5.yaml",
			targetConfigPath:  "../../test/e2e/testdata/config_aws_1.14.0.yaml",
			scenario:          NodeConformance,
		},
		{
			name:              "upgrade k8s 1.13.5 cluster to 1.14.0 on DO",
			provider:          DigitalOcean,
			initialVersion:    "v1.13.5",
			targetVersion:     "v1.14.0",
			initialConfigPath: "../../test/e2e/testdata/config_do_1.13.5.yaml",
			targetConfigPath:  "../../test/e2e/testdata/config_do_1.14.0.yaml",
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
			if testClusterVersion != tc.targetVersion {
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

			client, err := dynclient.New(restConfig, dynclient.Options{})
			if err != nil {
				t.Fatalf("failed to init dynamic client: %s", err)
			}

			t.Log("waiting for nodes to become ready")
			err = waitForNodesReady(client)
			if err != nil {
				t.Fatalf("nodes are not ready: %v", err)
			}

			t.Log("verifying cluster version before upgrade")
			err = verifyVersion(client, metav1.NamespaceSystem, tc.initialVersion)
			if err != nil {
				t.Fatalf("version mismatch before running upgrade: %v", err)
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
			err = waitForNodesReady(client)
			if err != nil {
				t.Fatalf("nodes are not ready: %v", err)
			}

			t.Log("verifying cluster version after upgrade")
			err = verifyVersion(client, metav1.NamespaceSystem, tc.targetVersion)
			if err != nil {
				t.Fatalf("version mismatch after running upgrade: %v", err)
			}

			t.Log("polling nodes to verify are all workers upgraded")
			err = waitForNodesUpgraded(client, tc.targetVersion)
			if err != nil {
				t.Fatalf("nodes are not running the target version: %v", err)
			}

			t.Log("run e2e tests")
			err = clusterVerifier.Verify(tc.scenario)
			if err != nil {
				t.Fatalf("e2e tests failed: %v", err)
			}
		})
	}
}

func waitForNodesReady(client dynclient.Client) error {
	return wait.Poll(5*time.Second, 3*time.Minute, func() (bool, error) {
		nodes := corev1.NodeList{}
		nodeListOpts := dynclient.ListOptions{}
		nodeListOpts.SetLabelSelector(fmt.Sprintf("%s=%s", labelControlPlaneNode, ""))

		err := client.List(context.Background(), &nodeListOpts, &nodes)
		if err != nil {
			return false, errors.Wrap(err, "unable to list nodes")
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

func waitForNodesUpgraded(client dynclient.Client, targetVersion string) error {
	reqVer, err := semver.NewVersion(targetVersion)
	if err != nil {
		return errors.Wrap(err, "desired version is invalid")
	}

	return wait.Poll(5*time.Second, 10*time.Minute, func() (bool, error) {
		nodes := corev1.NodeList{}
		err := client.List(context.Background(), &dynclient.ListOptions{}, &nodes)
		if err != nil {
			return false, errors.Wrap(err, "unable to list nodes")
		}

		// In this case it's safe to check kubelet version because once nodes are replaced
		// there are provisioned from zero with the new version, so we'll not have
		// kubelet and apiserver version mismatch.
		for _, n := range nodes.Items {
			kubeletVer, err := semver.NewVersion(n.Status.NodeInfo.KubeletVersion)
			if err != nil {
				return false, err
			}
			if reqVer.Compare(kubeletVer) != 0 {
				return false, nil
			}
		}
		return true, nil
	})
}

func verifyVersion(client dynclient.Client, namespace string, targetVersion string) error {
	reqVer, err := semver.NewVersion(targetVersion)
	if err != nil {
		return errors.Wrap(err, "desired version is invalid")
	}

	nodes := corev1.NodeList{}
	nodeListOpts := dynclient.ListOptions{}
	_ = nodeListOpts.SetLabelSelector(fmt.Sprintf("%s=%s", labelControlPlaneNode, ""))
	err = client.List(context.Background(), &nodeListOpts, &nodes)
	if err != nil {
		return errors.Wrap(err, "failed to list nodes")
	}

	// Kubelet version check
	for _, n := range nodes.Items {
		kubeletVer, err := semver.NewVersion(n.Status.NodeInfo.KubeletVersion)
		if err != nil {
			return err
		}
		if reqVer.Compare(kubeletVer) != 0 {
			return errors.Errorf("kubelet version mismatch: expected %v, got %v", reqVer.String(), kubeletVer.String())
		}
	}

	apiserverPods := corev1.PodList{}
	podsListOpts := dynclient.ListOptions{Namespace: namespace}
	_ = podsListOpts.SetLabelSelector("component=kube-apiserver")
	err = client.List(context.Background(), &podsListOpts, &apiserverPods)
	if err != nil {
		return errors.Wrap(err, "unable to list apiserver pods")
	}

	for _, p := range apiserverPods.Items {
		apiserverVer, err := parseContainerImageVersion(p.Spec.Containers[0].Image)
		if err != nil {
			return errors.Wrap(err, "unable to parse apiserver version")
		}
		if reqVer.Compare(apiserverVer) != 0 {
			return errors.Errorf("apiserver version mismatch: expected %v, got %v", reqVer.String(), apiserverVer.String())
		}
	}

	return nil
}

func parseContainerImageVersion(image string) (*semver.Version, error) {
	ver := strings.Split(image, ":")
	if len(ver) != 2 {
		return nil, errors.Errorf("invalid container image format: %s", image)
	}
	return semver.NewVersion(ver[1])
}
