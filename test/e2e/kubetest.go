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
	"os"
)

const (
	NodeConformance = `\[NodeConformance\]`
	Conformance     = `\[Conformance\]`
)

const skip = `Alpha|\[(Disruptive|Feature:[^\]]+|Flaky|Serial|Slow)\]`

// Kubetest configures the Kubetest conformance tester
type Kubetest struct {
	kubetestDir       string
	kubernetesVersion string
	// envVars Kubetest environment variables
	envVars map[string]string
}

// NewKubetest creates and provisions the Kubetest structure
func NewKubetest(k8sVersion, kubetestDir string, envVars map[string]string) Kubetest {
	return Kubetest{
		kubetestDir:       kubetestDir,
		kubernetesVersion: k8sVersion,
		envVars:           envVars,
	}

}

// Verify verifies the cluster
func (p *Kubetest) Verify(scenario string) error {
	k8sVersionPath := fmt.Sprintf("%s/kubernetes-%s/kubernetes/version", p.kubetestDir, p.kubernetesVersion)
	if _, err := os.Stat(k8sVersionPath); os.IsNotExist(err) {
		err = getK8sBinaries(p.kubetestDir, p.kubernetesVersion)
		if err != nil {
			return err
		}
	}

	k8sPath := fmt.Sprintf("%s/kubernetes-%s/kubernetes", p.kubetestDir, p.kubernetesVersion)

	testsArgs := fmt.Sprintf("--test_args=--ginkgo.focus=%s --ginkgo.skip=%s -ginkgo.noColor=true -ginkgo.flakeAttempts=2", scenario, skip)

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
	_, err = executeCommand(k8sPath, "kubetest", []string{fmt.Sprintf("--extract=%s", version)}, nil)
	if err != nil {
		return fmt.Errorf("getting kubernetes binaries failed: %v", err)
	}
	return nil
}
