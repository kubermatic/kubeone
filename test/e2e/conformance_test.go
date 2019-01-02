// +build e2e

package e2e

import (
	"fmt"
	"flag"
	"testing"
)

// testRunIdentifier aka. the build number, a unique identifier for the test run.
var testRunIdentifier = *flag.String("identifier", "", "The unique identifier for this test run")

func TestClusterConformance(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name              string
		provider          Provider
		kubernetesVersion string
		scenario          string
		region            string
	}{
		{
			name:              "scenario 1, verify k8s cluster deployment on AWS",
			provider:          AWS,
			kubernetesVersion: "v1.13.1",
			scenario:          NodeConformance,
			region:            "eu-west-3",
		},
	}

	for _, tc := range testcases {
		// to satisfy scope linter
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testPostfix := RandomString(8)
			testName := fmt.Sprintf("test-%s", testPostfix)
			testPath := fmt.Sprintf("../../_build/%s", testName)

			pr := CreateProvisioner(tc.region, testName, testPath, testRunIdentifier, tc.provider)
			target := NewKubeone(testPath, "../../config.yaml.dist")
			clusterVerifier := NewKubetest(tc.kubernetesVersion, "../../_build", map[string]string{
				"KUBERNETES_CONFORMANCE_TEST": "y",
			})

			t.Log("check prerequisites")
			err := ValidateCommon()
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
			err = target.CreateKubeconfig()
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

func setupTearDown(p Provisioner, k Kubeone) func(t *testing.T) {
	return func(t *testing.T) {
		t.Log("cleanup ....")

		errKubeone := k.Reset()
		errProvisioner := p.Cleanup()

		if errKubeone != nil {
			t.Errorf("%v", errKubeone)
		}
		if errProvisioner != nil {
			t.Errorf("%v", errProvisioner)
		}
	}
}
