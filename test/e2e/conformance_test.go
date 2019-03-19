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
)

func TestClusterConformance(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name              string
		provider          string
		kubernetesVersion string
		scenario          string
		region            string
		configFilePath    string
	}{
		{
			name:              "verify k8s cluster deployment on AWS",
			provider:          AWS,
			kubernetesVersion: "v1.13.3",
			scenario:          NodeConformance,
			configFilePath:    "../../test/e2e/testdata/aws_config.yaml",
		},
		{
			name:              "verify k8s cluster deployment on DO",
			provider:          DigitalOcean,
			kubernetesVersion: "v1.13.3",
			scenario:          NodeConformance,
			configFilePath:    "../../test/e2e/testdata/do_config.yaml",
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
			_, err = target.CreateKubeconfig()
			if err != nil {
				t.Fatalf("creating kubeconfig failed: %v", err)
			}

			t.Log("run e2e tests")
			err = clusterVerifier.Verify(tc.scenario)
			if err != nil {
				t.Fatalf("e2e tests failed: %v", err)
			}
		})
	}
}
