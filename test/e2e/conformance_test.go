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
	"fmt"
	"testing"

	kubeonev1beta1 "k8c.io/kubeone/pkg/apis/kubeone/v1beta1"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	"k8c.io/kubeone/test/e2e/provisioner"
	"k8c.io/kubeone/test/e2e/testutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	clusterNetworkPodCIDR     = "192.168.0.0/16"
	clusterNetworkServiceCIDR = "172.16.0.0/12"
)

func TestClusterConformance(t *testing.T) { //nolint:gocyclo
	testcases := []struct {
		name                  string
		provider              string
		providerExternal      bool
		scenario              string
		configFilePath        string
		configAPIVersion      string
		expectedNumberOfNodes int
	}{
		{
			name:                  "verify k8s cluster deployment on AWS",
			provider:              provisioner.AWS,
			providerExternal:      false,
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_aws.yaml",
			expectedNumberOfNodes: 6, // 3 control planes + 3 workers
		},
		{
			name:                  "verify k8s cluster deployment on DO",
			provider:              provisioner.DigitalOcean,
			providerExternal:      true,
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_do.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 worker
		},
		{
			name:                  "verify k8s cluster deployment on Hetzner",
			provider:              provisioner.Hetzner,
			providerExternal:      true,
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_hetzner.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 worker
		},
		{
			name:                  "verify k8s cluster deployment on GCE",
			provider:              provisioner.GCE,
			providerExternal:      false,
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_gce.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 worker
		},
		{
			name:                  "verify k8s cluster deployment on Equinix Metal",
			provider:              provisioner.EquinixMetal,
			providerExternal:      true,
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_equinix_metal.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 worker
		},
		{
			name:                  "verify k8s cluster deployment on OpenStack",
			provider:              provisioner.OpenStack,
			providerExternal:      true,
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_os.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 worker
		},
	}

	for _, tc := range testcases {
		// to satisfy scope linter
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			checkEnv(t)

			// Only run selected test suite.
			// Test options are controlled using flags.
			if testProvider != tc.provider {
				t.SkipNow()
			}

			if len(testRunIdentifier) == 0 {
				t.Fatal("-identifier must be set")
			}

			if len(testTargetVersion) == 0 {
				t.Fatal("-target-version must be set")
			}

			if testConfigAPIVersion != kubeonev1beta1.SchemeGroupVersion.Version &&
				testConfigAPIVersion != kubeonev1beta2.SchemeGroupVersion.Version {
				t.Fatal("-config-api-version must be v1beta1 or v1beta2")
			}

			if err := ValidateOperatingSystem(testOSControlPlane); err != nil {
				t.Fatal(err)
			}

			if err := ValidateOperatingSystem(testOSWorkers); err != nil {
				t.Fatal(err)
			}

			osControlPlane := OperatingSystem(testOSControlPlane)
			osWorkers := OperatingSystem(testOSWorkers)

			t.Logf("Running conformance tests for Kubernetes v%s...", testTargetVersion)

			// Create provisioner
			testPath := fmt.Sprintf("../../_build/%s", testRunIdentifier)

			pr, err := provisioner.CreateProvisioner(testPath, testRunIdentifier, tc.provider)
			if err != nil {
				t.Fatalf("failed to create provisioner: %v", err)
			}

			// Create KubeOne target and prepare kubetest
			target := NewKubeone(testPath, tc.configFilePath)

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
			case kubeonev1beta2.SchemeGroupVersion.Version:
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

			// Ensure cleanup at the end
			teardown := setupTearDown(pr, target)
			defer teardown(t)

			// Create infrastructure
			t.Log("Provisioning infrastructure using Terraform...")
			args := []string{}

			if osControlPlane != OperatingSystemDefault {
				tfFlags, errFlags := ControlPlaneImageFlags(tc.provider, osControlPlane)
				if errFlags != nil {
					t.Fatalf("failed to discover control plane os image: %v", errFlags)
				}

				args = append(args, tfFlags...)
			}

			if osWorkers != OperatingSystemDefault {
				switch {
				case osWorkers == OperatingSystemCentOS7:
					args = append(args, "-var", fmt.Sprintf("worker_os=%s", "centos"))
					args = append(args, "-var", fmt.Sprintf("ami=%s", AWSCentOS7AMI))
				default:
					args = append(args, "-var", fmt.Sprintf("worker_os=%s", osWorkers))
				}
			}

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
				args = []string{}

				if osControlPlane != OperatingSystemDefault {
					tfFlags, errFlags := ControlPlaneImageFlags(tc.provider, osControlPlane)
					if errFlags != nil {
						t.Fatalf("failed to discover control plane os image: %v", errFlags)
					}
					args = append(args, tfFlags...)
				}

				if osWorkers != OperatingSystemDefault {
					switch {
					case osWorkers == OperatingSystemCentOS7:
						args = append(args, "-var", fmt.Sprintf("worker_os=%s", "centos"))
						args = append(args, "-var", fmt.Sprintf("ami=%s", AWSCentOS7AMI))
					default:
						args = append(args, "-var", fmt.Sprintf("worker_os=%s", osWorkers))
					}
				}

				_, err = pr.Provision(args...)
				if err != nil {
					t.Fatalf("failed to provision the infrastructure: %v", err)
				}
			}

			// Build clientset
			t.Log("Building Kubernetes clientset...")

			restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
			if err != nil {
				t.Errorf("unable to build clientset from kubeconfig bytes: %v", err)
			}

			client, err := dynclient.New(restConfig, dynclient.Options{})
			if err != nil {
				t.Fatalf("failed to init dynamic client: %s", err)
			}

			// Ensure nodes are ready and version is matching desired
			t.Log("Waiting for all nodes to become ready...")
			if err = waitForNodesReady(t, client, tc.expectedNumberOfNodes); err != nil {
				t.Fatalf("failed to bring up all nodes up: %v", err)
			}

			t.Log("Verifying cluster version...")
			if err = verifyVersion(client, metav1.NamespaceSystem, testTargetVersion); err != nil {
				t.Fatalf("version mismatch: %v", err)
			}

			clusterVerifier := NewKubetest(testTargetVersion, "../../_build", map[string]string{
				"KUBERNETES_CONFORMANCE_TEST": "y",
			})

			// Run NodeConformance tests
			t.Log("Running conformance tests (this can take up to 30 minutes)...")
			skipTests := Skip
			if osControlPlane == OperatingSystemFlatcar || osWorkers == OperatingSystemFlatcar {
				skipTests = SkipFlatcar
			}
			if err = clusterVerifier.Verify(tc.scenario, skipTests); err != nil {
				t.Fatalf("e2e tests failed: %v", err)
			}
		})
	}
}
