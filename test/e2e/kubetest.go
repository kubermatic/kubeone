package e2e

import (
	"fmt"
	"os"
)

const (
	Pods            = "Pods"
	NodeConformance = `\[NodeConformance\]`
	Conformance     = `\[Conformance\]`
)

const skip = `Alpha|\[(Disruptive|Feature:[^\]]+|Flaky|Serial|Slow)\]`

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
func (p *Kubetest) Verify(scenario string) error {
	k8sVersionPath := fmt.Sprintf("%s/kubernetes-%s/kubernetes/version", p.kubetestDir, p.kubernetesVersion)
	if _, err := os.Stat(k8sVersionPath); os.IsNotExist(err) {
		err = getK8sBinaries(p.kubetestDir, p.kubernetesVersion)
		if err != nil {
			return err
		}
	}

	k8sPath := fmt.Sprintf("%s/kubernetes-%s/kubernetes", p.kubetestDir, p.kubernetesVersion)

	testsArgs := fmt.Sprintf("--test_args=--ginkgo.focus=%s --ginkgo.skip=%s", scenario, skip)

	_, err := executeCommand(k8sPath, "kubetest", []string{"--provider=skeleton", "--test", "--ginkgo-parallel", "--check-version-skew=false", testsArgs}, p.envVars)
	if err != nil {
		return fmt.Errorf("k8s conformnce tests failed: %v", err)
	}

	return nil
}

// getK8sBinaries retrieves kubernetes binaries for version
func getK8sBinaries(kubetestDir string, version string) error {

	k8sPath := fmt.Sprintf("%s/kubernetes-%s", kubetestDir, version)
	err := os.MkdirAll(k8sPath, 0755)
	if err != nil {
		return fmt.Errorf("unable to create directory %s", k8sPath)
	}
	if err != nil {
		return err
	}
	_, err = executeCommand(k8sPath, "kubetest", []string{fmt.Sprintf("--extract=%s", version)})
	if err != nil {
		return fmt.Errorf("getting kubernetes binaries failed: %v", err)
	}
	return nil
}
