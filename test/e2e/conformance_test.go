package e2e

import (
	"fmt"
	"testing"
)

func TestClusterConformance(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		Name              string
		Provider          Provider
		KubernetesVersion string
		Scenario          TestsScenario
		Region            string
	}{
		{
			Name:              "scenario 1, verify k8s cluster deployment on AWS",
			Provider:          AWS,
			KubernetesVersion: "v1.12.3",
			Scenario:          NodeConformance,
			Region:            "eu-west-3",
		},
	}

	for _, tc := range testcases {
		// to satisfy scope linter
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			testPostfix := RandomString(8)
			testName := fmt.Sprintf("test-%s", testPostfix)
			testPath := fmt.Sprintf("../../_build/%s", testName)

			pr := CreateProvisioner(tc.Region, testName, testPath, tc.Provider)
			target := NewKubeone(testPath, "../../config.yaml.dist")
			clusterVerifier := NewKubetest(tc.KubernetesVersion, "../../_build", map[string]string{
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
			err = clusterVerifier.Verify(tc.Scenario)
			if err != nil {
				t.Fatalf("e2e tests failed: %v", err)
			}
		})
	}
}

func setupTearDown(p Provisioner, k Kubeone) func(t *testing.T) {
	return func(t *testing.T) {
		t.Log("clenaup ....")

		err := k.Reset()
		if err != nil {
			t.Logf("%v", err)
		}
		err = p.Cleanup()
		if err != nil {
			t.Logf("%v", err)
		}
	}
}
