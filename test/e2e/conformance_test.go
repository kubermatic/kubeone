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
	"fmt"
	"testing"

	"github.com/kubermatic/kubeone/test/e2e/provisioner"
	"github.com/kubermatic/kubeone/test/e2e/testutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestClusterConformance(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name                  string
		provider              string
		providerExternal      bool
		kubernetesVersion     string
		scenario              string
		configFilePath        string
		expectedNumberOfNodes int
	}{
		{
			name:                  "verify k8s 1.13.5 cluster deployment on AWS",
			provider:              provisioner.AWS,
			providerExternal:      false,
			kubernetesVersion:     "1.13.5",
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_aws_1.13.5.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 worker
		},
		{
			name:                  "verify k8s 1.14.1 cluster deployment on AWS",
			provider:              provisioner.AWS,
			providerExternal:      false,
			kubernetesVersion:     "1.14.1",
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_aws_1.14.1.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 worker
		},
		{
			name:                  "verify k8s 1.13.5 cluster deployment on DO",
			provider:              provisioner.DigitalOcean,
			providerExternal:      true,
			kubernetesVersion:     "1.13.5",
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_do_1.13.5.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 worker
		},
		{
			name:                  "verify k8s 1.14.1 cluster deployment on DO",
			provider:              provisioner.DigitalOcean,
			providerExternal:      true,
			kubernetesVersion:     "1.14.1",
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_do_1.14.1.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 worker
		},
		{
			name:                  "verify k8s 1.13.5 cluster deployment on Hetzner",
			provider:              provisioner.Hetzner,
			providerExternal:      true,
			kubernetesVersion:     "1.13.5",
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_hetzner_1.13.5.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 worker
		},
		{
			name:                  "verify k8s 1.14.1 cluster deployment on Hetzner",
			provider:              provisioner.Hetzner,
			providerExternal:      true,
			kubernetesVersion:     "1.14.1",
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_hetzner_1.14.1.yaml",
			expectedNumberOfNodes: 4, // 3 control planes + 1 worker
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
			if testClusterVersion != tc.kubernetesVersion {
				t.SkipNow()
			}

			// Create provisioner
			testPath := fmt.Sprintf("../../_build/%s", testRunIdentifier)
			pr, err := provisioner.CreateProvisioner(testPath, testRunIdentifier, tc.provider)
			if err != nil {
				t.Fatalf("failed to create provisioner: %v", err)
			}

			// Create KubeOne target and prepare kubetest
			target := NewKubeone(testPath, tc.configFilePath)
			clusterVerifier := NewKubetest(tc.kubernetesVersion, "../../_build", map[string]string{
				"KUBERNETES_CONFORMANCE_TEST": "y",
			})

			// Ensure terraform, kubetest and all needed prerequisites are in place before running test
			t.Log("Validating prerequisites…")
			err = testutil.ValidateCommon()
			if err != nil {
				t.Fatalf("unable to validate prerequisites: %v", err)
			}

			// Create configuration manifest
			t.Log("Creating KubeOneCluster manifest…")
			err = target.CreateConfig(tc.kubernetesVersion, tc.provider, tc.providerExternal)
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
				t.Errorf("unable to build clientset from kubeconfig bytes: %v", err)
			}
			client, err := dynclient.New(restConfig, dynclient.Options{})
			if err != nil {
				t.Fatalf("failed to init dynamic client: %s", err)
			}

			// Ensure nodes are ready and version is matching desired
			t.Log("Waiting for all nodes to become ready…")
			err = waitForNodesReady(client, tc.expectedNumberOfNodes)
			if err != nil {
				t.Fatalf("failed to bring up all nodes up: %v", err)
			}
			t.Log("Verifying cluster version…")
			err = verifyVersion(client, metav1.NamespaceSystem, tc.kubernetesVersion)
			if err != nil {
				t.Fatalf("version mismatch: %v", err)
			}

			// Run NodeConformance tests
			t.Log("Running conformance tests (this can take up to 30 minutes)…")
			err = clusterVerifier.Verify(tc.scenario)
			if err != nil {
				t.Fatalf("e2e tests failed: %v", err)
			}
		})
	}
}
