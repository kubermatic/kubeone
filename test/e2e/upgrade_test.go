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
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/test/e2e/provisioner"
	"github.com/kubermatic/kubeone/test/e2e/testutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	labelControlPlaneNode = "node-role.kubernetes.io/master"
	delayUpgrade          = 2 * time.Minute
)

func TestClusterUpgrade(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name                  string
		provider              string
		providerExternal      bool
		initialVersion        string
		targetVersion         string
		initialConfigPath     string
		targetConfigPath      string
		expectedNumberOfNodes int
	}{
		{
			name:                  "upgrade k8s 1.13.5 cluster to 1.14.1 on AWS",
			provider:              provisioner.AWS,
			providerExternal:      false,
			initialVersion:        "1.13.5",
			targetVersion:         "1.14.1",
			initialConfigPath:     "../../test/e2e/testdata/config_aws_1.13.5.yaml",
			targetConfigPath:      "../../test/e2e/testdata/config_aws_1.14.1.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 workers
		},
		{
			name:                  "upgrade k8s 1.13.5 cluster to 1.14.1 on DO",
			provider:              provisioner.DigitalOcean,
			providerExternal:      true,
			initialVersion:        "1.13.5",
			targetVersion:         "1.14.1",
			initialConfigPath:     "../../test/e2e/testdata/config_do_1.13.5.yaml",
			targetConfigPath:      "../../test/e2e/testdata/config_do_1.14.1.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 workers
		},
		{
			name:                  "upgrade k8s 1.13.5 cluster to 1.14.1 on Hetzner",
			provider:              provisioner.Hetzner,
			providerExternal:      true,
			initialVersion:        "1.13.5",
			targetVersion:         "1.14.1",
			initialConfigPath:     "../../test/e2e/testdata/config_hetzner_1.13.5.yaml",
			targetConfigPath:      "../../test/e2e/testdata/config_hetzner_1.14.1.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 workers
		},
	}

	for _, tc := range testcases {
		// to satisfy scope linter
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Only run selected test suite.
			// Test options are controlled using flags.
			if len(testRunIdentifier) == 0 {
				t.Fatalf("-identifier must be set")
			}
			if testProvider != tc.provider {
				t.SkipNow()
			}
			if testClusterVersion != tc.targetVersion {
				t.SkipNow()
			}

			// Create provisioner
			testPath := fmt.Sprintf("../../_build/%s", testRunIdentifier)
			pr, err := provisioner.CreateProvisioner(testPath, testRunIdentifier, tc.provider)
			if err != nil {
				t.Fatalf("failed to create provisioner: %v", err)
			}

			// Create KubeOne target
			target := NewKubeone(testPath, tc.initialConfigPath)

			// Ensure terraform, kubetest and all needed prerequisites are in place before running test
			t.Log("Validating prerequisites…")
			err = testutil.ValidateCommon()
			if err != nil {
				t.Fatalf("unable to validate prerequisites: %v", err)
			}

			// Create configuration manifest
			t.Log("Creating KubeOneCluster manifest…")
			err = target.CreateConfig(tc.initialVersion, tc.provider, tc.providerExternal)
			if err != nil {
				t.Fatalf("failed to create KubeOneCluster manifest: %v", err)
			}

			// Ensure cleanup at the end
			teardown := setupTearDown(pr, target)
			defer teardown(t)

			// Create infrastructure
			t.Log("Provisioning infrastructure using Terraform…")
			tf, err := pr.Provision()
			if err != nil {
				t.Fatalf("failed to provision the infrastructure: %v", err)
			}

			// Run 'kubeone install'
			t.Log("Running 'kubeone install'…")
			err = target.Install(tf)
			if err != nil {
				t.Fatalf("failed to install cluster ('kubeone install'): %v", err)
			}

			// Run 'kubeone kubeconfig'
			t.Log("Downloading kubeconfig…")
			kubeconfig, err := target.Kubeconfig()
			if err != nil {
				t.Fatalf("failed to download kubeconfig failed ('kubeone kubeconfig'): %v", err)
			}

			// Build clientset
			t.Log("Building Kubernetes clientset…")
			restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
			if err != nil {
				t.Errorf("unable to build config from kubeconfig bytes: %v", err)
			}
			client, err := dynclient.New(restConfig, dynclient.Options{})
			if err != nil {
				t.Fatalf("failed to init dynamic client: %s", err)
			}

			// Ensure nodes are ready and version is matching desired
			t.Log("Waiting for all nodes to become ready…")
			err = waitForNodesReady(client, tc.expectedNumberOfNodes)
			if err != nil {
				t.Fatalf("nodes are not ready: %v", err)
			}
			t.Log("Verifying cluster version before running upgrade…")
			err = verifyVersion(client, metav1.NamespaceSystem, tc.initialVersion)
			if err != nil {
				t.Fatalf("version mismatch before running upgrade: %v", err)
			}

			// Delay running upgrade to leave some time for all components to become ready
			t.Logf("Waiting %s for nodes to settle down…", delayUpgrade.String())
			time.Sleep(delayUpgrade)

			// Create a new KubeOne provisioner pointing to the new configuration file
			target = NewKubeone(testPath, tc.targetConfigPath)

			// Create new configuration manifest
			t.Log("Creating KubeOneCluster manifest…")
			err = target.CreateConfig(tc.targetVersion, tc.provider, tc.providerExternal)
			if err != nil {
				t.Fatalf("failed to create KubeOneCluster manifest: %v", err)
			}

			// Run 'kubeone install'
			t.Log("Running 'kubeone upgrade'…")
			err = target.Upgrade()
			if err != nil {
				t.Fatalf("failed to upgrade the cluster ('kubeone upgrade'): %v", err)
			}

			// Ensure nodes are ready and version is matching desired
			t.Log("Waiting for all nodes to become ready…")
			err = waitForNodesReady(client, tc.expectedNumberOfNodes)
			if err != nil {
				t.Fatalf("nodes are not ready: %v", err)
			}
			t.Log("Verifying cluster version after running upgrade…")
			err = verifyVersion(client, metav1.NamespaceSystem, tc.targetVersion)
			if err != nil {
				t.Fatalf("version mismatch before running upgrade: %v", err)
			}
			t.Log("Polling nodes to verify are all workers upgraded…")
			err = waitForNodesUpgraded(client, tc.targetVersion)
			if err != nil {
				t.Fatalf("nodes are not running the target version: %v", err)
			}
		})
	}
}

func waitForNodesUpgraded(client dynclient.Client, targetVersion string) error {
	reqVer, err := semver.NewVersion(targetVersion)
	if err != nil {
		return errors.Wrap(err, "desired version is invalid")
	}

	return wait.Poll(5*time.Second, 20*time.Minute, func() (bool, error) {
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
