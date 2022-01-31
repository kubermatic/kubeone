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

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"

	kubeonev1beta1 "k8c.io/kubeone/pkg/apis/kubeone/v1beta1"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	"k8c.io/kubeone/test/e2e/provisioner"
	"k8c.io/kubeone/test/e2e/testutil"

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

func TestClusterUpgrade(t *testing.T) { //nolint:gocyclo
	testcases := []struct {
		name                  string
		provider              string
		providerExternal      bool
		initialConfigPath     string
		targetConfigPath      string
		expectedNumberOfNodes int
	}{
		{
			name:                  "upgrade k8s cluster on AWS",
			provider:              provisioner.AWS,
			providerExternal:      false,
			initialConfigPath:     "../../test/e2e/testdata/config_aws_initial.yaml",
			targetConfigPath:      "../../test/e2e/testdata/config_aws_target.yaml",
			expectedNumberOfNodes: 6, // 3 control planes + 3 workers
		},
		{
			name:                  "upgrade k8s cluster on DO",
			provider:              provisioner.DigitalOcean,
			providerExternal:      true,
			initialConfigPath:     "../../test/e2e/testdata/config_do_initial.yaml",
			targetConfigPath:      "../../test/e2e/testdata/config_do_target.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 3 workers
		},
		{
			name:                  "upgrade k8s cluster on Hetzner",
			provider:              provisioner.Hetzner,
			providerExternal:      true,
			initialConfigPath:     "../../test/e2e/testdata/config_hetzner_initial.yaml",
			targetConfigPath:      "../../test/e2e/testdata/config_hetzner_target.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 3 workers
		},
		{
			name:                  "upgrade k8s cluster on GCE",
			provider:              provisioner.GCE,
			providerExternal:      false,
			initialConfigPath:     "../../test/e2e/testdata/config_gce_initial.yaml",
			targetConfigPath:      "../../test/e2e/testdata/config_gce_target.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 3 workers
		},
		{
			name:                  "upgrade k8s cluster on Equinix Metal",
			provider:              provisioner.EquinixMetal,
			providerExternal:      true,
			initialConfigPath:     "../../test/e2e/testdata/config_packet_initial.yaml",
			targetConfigPath:      "../../test/e2e/testdata/config_packet_target.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 3 workers
		},
		{
			name:                  "upgrade k8s cluster on OpenStack",
			provider:              provisioner.OpenStack,
			providerExternal:      true,
			initialConfigPath:     "../../test/e2e/testdata/config_openstack_initial.yaml",
			targetConfigPath:      "../../test/e2e/testdata/config_openstack_target.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 3 workers
		},
	}

	for _, tc := range testcases {
		// to satisfy scope linter
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			checkEnv(t)

			// Only run selected test suite.
			// Test options are controlled using flags.
			if len(testRunIdentifier) == 0 {
				t.Fatalf("-identifier must be set")
			}

			if len(testInitialVersion) == 0 {
				t.Fatal("-initial-version must be set")
			}

			if len(testTargetVersion) == 0 {
				t.Fatal("-target-version must be set")
			}

			if testConfigAPIVersion != kubeonev1beta1.SchemeGroupVersion.Version &&
				testConfigAPIVersion != kubeonev1beta2.SchemeGroupVersion.Version {
				t.Fatal("-config-api-version must be v1beta1 or v1beta2")
			}

			if testProvider != tc.provider {
				t.SkipNow()
			}

			t.Logf("Running upgrade tests from Kubernetes v%s to v%s...", testInitialVersion, testTargetVersion)

			// Create provisioner
			testPath := fmt.Sprintf("../../_build/%s", testRunIdentifier)

			pr, err := provisioner.CreateProvisioner(testPath, testRunIdentifier, tc.provider)
			if err != nil {
				t.Fatalf("failed to create provisioner: %v", err)
			}

			// Create KubeOne target
			target := NewKubeone(testPath, tc.initialConfigPath)

			// Ensure terraform, kubetest and all needed prerequisites are in place before running test
			t.Log("Validating prerequisites...")

			err = testutil.ValidateCommon()
			if err != nil {
				t.Fatalf("unable to validate prerequisites: %v", err)
			}

			// Create configuration manifest
			t.Log("Creating KubeOneCluster manifest...")
			var (
				clusterNetworkPod     string
				clusterNetworkService string
			)

			if tc.provider == provisioner.OpenStack {
				clusterNetworkPod = clusterNetworkPodCIDR
				clusterNetworkService = clusterNetworkServiceCIDR
			}

			switch testConfigAPIVersion {
			case kubeonev1beta1.SchemeGroupVersion.Version:
				err = target.CreateV1Beta1Config(testInitialVersion,
					tc.provider,
					tc.providerExternal,
					clusterNetworkPod,
					clusterNetworkService,
					testCredentialsFile,
					testContainerRuntime.ContainerRuntimeConfig(),
				)
				if err != nil {
					t.Fatalf("failed to create KubeOneCluster manifest: %v", err)
				}
			case kubeonev1beta2.SchemeGroupVersion.Version:
				err = target.CreateV1Beta2Config(testInitialVersion,
					tc.provider,
					tc.providerExternal,
					clusterNetworkPod,
					clusterNetworkService,
					testCredentialsFile,
					testContainerRuntime.ContainerRuntimeConfig(),
				)
				if err != nil {
					t.Fatalf("failed to create KubeOneCluster manifest: %v", err)
				}
			}

			// Ensure cleanup at the end
			teardown := setupTearDown(pr, target)
			defer teardown(t)

			// Create infrastructure
			t.Log("Provisioning infrastructure using Terraform...")
			args := []string{}

			if tc.provider == provisioner.GCE {
				args = append(args, "-var", "control_plane_target_pool_members_count=1")
			}

			tf, err := pr.Provision(args...)
			if err != nil {
				t.Fatalf("failed to provision the infrastructure: %v", err)
			}

			// Run 'kubeone install'
			t.Log("Running 'kubeone install'...")
			var installFlags []string

			if tc.provider == provisioner.OpenStack {
				installFlags = append(installFlags, "-c", "/tmp/credentials.yaml")
			}

			err = target.Install(tf, installFlags)
			if err != nil {
				t.Fatalf("failed to install cluster ('kubeone install'): %v", err)
			}

			// Run 'kubeone kubeconfig'
			t.Log("Downloading kubeconfig...")

			kubeconfig, err := target.Kubeconfig()
			if err != nil {
				t.Fatalf("failed to download kubeconfig failed ('kubeone kubeconfig'): %v", err)
			}

			// Run Terraform again for GCE to add nodes to the load balancer
			if tc.provider == provisioner.GCE {
				t.Log("Adding other control plane nodes to the load balancer...")

				_, err = pr.Provision()
				if err != nil {
					t.Fatalf("failed to provision the infrastructure: %v", err)
				}
			}

			// Build clientset
			t.Log("Building Kubernetes clientset...")

			restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
			if err != nil {
				t.Errorf("unable to build config from kubeconfig bytes: %v", err)
			}

			client, err := dynclient.New(restConfig, dynclient.Options{})
			if err != nil {
				t.Fatalf("failed to init dynamic client: %s", err)
			}

			// Ensure nodes are ready and version is matching desired
			t.Log("Waiting for all nodes to become ready...")

			err = waitForNodesReady(t, client, tc.expectedNumberOfNodes)
			if err != nil {
				t.Fatalf("nodes are not ready: %v", err)
			}
			t.Log("Verifying cluster version before running upgrade...")

			err = verifyVersion(client, metav1.NamespaceSystem, testInitialVersion)
			if err != nil {
				t.Fatalf("version mismatch before running upgrade: %v", err)
			}

			// Delay running upgrade to leave some time for all components to become ready
			t.Logf("Waiting %s for nodes to settle down...", delayUpgrade.String())
			time.Sleep(delayUpgrade)

			// Create a new KubeOne provisioner pointing to the new configuration file
			target = NewKubeone(testPath, tc.targetConfigPath)

			// Create new configuration manifest
			t.Log("Creating KubeOneCluster manifest...")
			if tc.provider == provisioner.OpenStack {
				clusterNetworkPod = "192.168.0.0/16"
				clusterNetworkService = "172.16.0.0/12"
			}

			switch testConfigAPIVersion {
			case "v1beta1":
				err = target.CreateV1Beta1Config(testTargetVersion,
					tc.provider,
					tc.providerExternal,
					clusterNetworkPod,
					clusterNetworkService,
					testCredentialsFile,
					testContainerRuntime.ContainerRuntimeConfig(),
				)
				if err != nil {
					t.Fatalf("failed to create KubeOneCluster manifest: %v", err)
				}
			case "v1beta2":
				err = target.CreateV1Beta2Config(testTargetVersion,
					tc.provider,
					tc.providerExternal,
					clusterNetworkPod,
					clusterNetworkService,
					testCredentialsFile,
					testContainerRuntime.ContainerRuntimeConfig(),
				)
				if err != nil {
					t.Fatalf("failed to create KubeOneCluster manifest: %v", err)
				}
			}

			// Run 'kubeone upgrade'
			t.Log("Running 'kubeone upgrade'...")
			var upgradeFlags []string

			if tc.provider == provisioner.OpenStack {
				upgradeFlags = append(upgradeFlags, "-c", "/tmp/credentials.yaml")
			}

			err = target.Upgrade(upgradeFlags)
			if err != nil {
				t.Fatalf("failed to upgrade the cluster ('kubeone upgrade'): %v", err)
			}

			// Ensure nodes are ready and version is matching desired
			t.Log("Waiting for all nodes to become ready...")

			err = waitForNodesReady(t, client, tc.expectedNumberOfNodes)
			if err != nil {
				t.Fatalf("nodes are not ready: %v", err)
			}

			t.Log("Verifying cluster version after running upgrade...")

			err = verifyVersion(client, metav1.NamespaceSystem, testTargetVersion)
			if err != nil {
				t.Fatalf("version mismatch before running upgrade: %v", err)
			}

			t.Log("Polling nodes to verify are all workers upgraded...")

			err = waitForNodesUpgraded(client, testTargetVersion)
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

		if err := client.List(context.Background(), &nodes); err != nil {
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
