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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestClusterConformance(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name                  string
		provider              string
		kubernetesVersion     string
		scenario              string
		configFilePath        string
		expectedNumberOfNodes int
	}{
		{
			name:                  "verify k8s 1.13.5 cluster deployment on AWS",
			provider:              AWS,
			kubernetesVersion:     "v1.13.5",
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_aws_1.13.5.yaml",
			expectedNumberOfNodes: 6, // 3 control planes + 3 workers
		},
		{
			name:                  "verify k8s 1.14.0 cluster deployment on AWS",
			provider:              AWS,
			kubernetesVersion:     "v1.14.0",
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_aws_1.14.0.yaml",
			expectedNumberOfNodes: 6, // 3 control planes + 3 workers
		},
		{
			name:                  "verify k8s 1.13.5 cluster deployment on DO",
			provider:              DigitalOcean,
			kubernetesVersion:     "v1.13.5",
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_do_1.13.5.yaml",
			expectedNumberOfNodes: 6, // 3 control planes + 3 workers
		},
		{
			name:                  "verify k8s 1.14.0 cluster deployment on DO",
			provider:              DigitalOcean,
			kubernetesVersion:     "v1.14.0",
			scenario:              NodeConformance,
			configFilePath:        "../../test/e2e/testdata/config_do_1.14.0.yaml",
			expectedNumberOfNodes: 6, // 3 control planes + 3 workers
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
			if testClusterVersion != tc.kubernetesVersion {
				t.SkipNow()
			}
			testPath := fmt.Sprintf("../../_build/%s", testRunIdentifier)

			pr, err := CreateProvisioner(testPath, testRunIdentifier, tc.provider)
			if err != nil {
				t.Fatal(err)
			}
			target := NewKubeone(testPath, tc.configFilePath)
			clusterVerifier := NewKubetest(tc.kubernetesVersion, "../../_build", map[string]string{
				"KUBERNETES_CONFORMANCE_TEST": "y",
			})

			t.Log("check prerequisites")
			err = ValidateCommon()
			if err != nil {
				t.Fatalf("%v", err)
			}

			teardown := setupTearDown(pr, target)
			defer teardown(t)

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
			err = waitForNodesReady(client, tc.expectedNumberOfNodes)
			if err != nil {
				t.Fatalf("nodes are not ready: %v", err)
			}

			t.Log("verifying cluster version")
			err = verifyVersion(client, metav1.NamespaceSystem, tc.kubernetesVersion)
			if err != nil {
				t.Fatalf("version mismatch: %v", err)
			}

			t.Log("run e2e tests")
			err = clusterVerifier.Verify(tc.scenario)
			if err != nil {
				t.Fatalf("e2e tests failed: %v", err)
			}
		})
	}
}
