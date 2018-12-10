package e2e

import (
	"fmt"
	"os"
)

const (
	Pods            TestsScenario = "Pods"
	NodeConformance TestsScenario = `\[NodeConformance\]`
	Conformance     TestsScenario = `\[Conformance\]`
)

const skip = `Alpha|\[(Disruptive|Feature:[^\]]+|Flaky|Serial|Slow)\]`

type TestsScenario string

// Kubetest struct
type Kubetest struct {
	kubetestDir       string
	kubernetesVersion string
	// envVars Kubetest environment variables
	envVars map[string]string
}

func NewKubetest(k8sVersion, kubetestDir string, envVars map[string]string) Kubetest {
	return Kubetest{
		kubetestDir:       kubetestDir,
		kubernetesVersion: k8sVersion,
		envVars:           envVars,
	}

}

// RunTests starts e2e tests
func (p *Kubetest) Verify(scenario TestsScenario) error {

	for k, v := range p.envVars {
		os.Setenv(k, v)
	}

	k8sVersionPath := fmt.Sprintf("%s/kubernetes-%s/kubernetes/version", p.kubetestDir, p.kubernetesVersion)
	if _, err := os.Stat(k8sVersionPath); os.IsNotExist(err) {
		err = getK8sBinaries(p.kubetestDir, p.kubernetesVersion)
		if err != nil {
			return err
		}

	}

	k8sPath := fmt.Sprintf("%s/kubernetes-%s/kubernetes", p.kubetestDir, p.kubernetesVersion)

	testsArgs := fmt.Sprintf("--test_args=--ginkgo.focus=%s --ginkgo.skip=%s", scenario, skip)

	_, stderr, exitCode := executeCommand(k8sPath, "kubetest", []string{"--provider=skeleton", "--test", "--ginkgo-parallel", "--check-version-skew=false", testsArgs})
	if exitCode != 0 {
		return fmt.Errorf("k8s conformnce tests failed: %s", stderr)
	}

	return nil
}

// getK8sBinaries retrieves kubernetes binaries for version
func getK8sBinaries(kubetestDir string, version string) error {

	k8sPath := fmt.Sprintf("%s/kubernetes-%s", kubetestDir, version)
	err := CreateDir(k8sPath)
	if err != nil {
		return err
	}
	_, stderr, exitCode := executeCommand(k8sPath, "kubetest", []string{fmt.Sprintf("--extract=%s", version)})
	if exitCode != 0 {
		return fmt.Errorf("getting kubernetes binaries failed: %s", stderr)
	}
	return nil
}
