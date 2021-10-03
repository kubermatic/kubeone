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
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"

	"k8c.io/kubeone/test/e2e/testutil"
)

const (
	NodeConformance = `\[NodeConformance\]`
	Conformance     = `\[Conformance\]`
	Skip            = `Alpha|\[(Disruptive|Feature:[^\]]+|Flaky|Serial|Slow)\]`
	SkipFlatcar     = `Alpha|\[(Disruptive|Feature:[^\]]+|Flaky|Serial|Slow)\]|should support subPath`
)

// Kubetest configures the Kubetest conformance tester
type Kubetest struct {
	kubetestDir       string
	kubernetesVersion string
	// envVars Kubetest environment variables
	envVars map[string]string
}

// NewKubetest creates and provisions the Kubetest structure
func NewKubetest(k8sVersion, kubetestDir string, envVars map[string]string) *Kubetest {
	return &Kubetest{
		kubetestDir:       filepath.Clean(kubetestDir),
		kubernetesVersion: k8sVersion,
		envVars:           envVars,
	}
}

// Verify verifies the cluster
func (p *Kubetest) Verify(scenario, skip string) error {
	kubetestPath, err := findKubetest(p.kubetestDir, p.kubernetesVersion)
	if err != nil {
		return fmt.Errorf("coudn't find kubetest scenarios: %w", err)
	}

	// Kubetest requires version to have the "v" prefix
	if !strings.HasPrefix(p.kubernetesVersion, "v") {
		p.kubernetesVersion = fmt.Sprintf("v%s", p.kubernetesVersion)
	}

	err = testutil.NewExec("./hack/ginkgo-e2e.sh",
		testutil.WithArgs(
			fmt.Sprintf("--ginkgo.focus=%s", scenario),
			fmt.Sprintf("--ginkgo.skip=%s", skip),
			"--ginkgo.flakeAttempts=2",
		),
		testutil.WithMapEnv(p.envVars),
		testutil.WithEnv(os.Environ()),
		testutil.WithEnvs(
			"GINKGO_PARALLEL=y",
			"GINKGO_NO_COLOR=y",
		),
		testutil.InDir(kubetestPath),
	).Run()
	if err != nil {
		return fmt.Errorf("k8s conformnce tests failed: %w", err)
	}

	return nil
}

// findKubetest tries to locate existing path to kubetest with specified version
// by trying to find "<basedir>/kubernetes-<version>/kubernetes/version" file,
// gradually removing parts of sematic version (e.g. trying versions: [1.16.2,
// 1.16, 1]).
func findKubetest(basedir, version string) (string, error) {
	sver, err := semver.NewVersion(version)
	if err != nil {
		return "", err
	}

	maj := sver.Major()
	min := sver.Minor()
	pat := sver.Patch()

	kubetestVersionsToTry := []string{
		fmt.Sprintf("%d.%d.%d", maj, min, pat),
		fmt.Sprintf("v%d.%d.%d", maj, min, pat),
		fmt.Sprintf("%d.%d", maj, min),
		fmt.Sprintf("v%d.%d", maj, min),
		fmt.Sprintf("%d", maj),
		fmt.Sprintf("v%d", maj),
	}

	for _, kubetestVersion := range kubetestVersionsToTry {
		candidateKubetestDir := fmt.Sprintf("%s/kubernetes-%s", basedir, kubetestVersion)
		fileToCheck := filepath.Join(candidateKubetestDir, "kubernetes/version")

		if _, err := os.Stat(fileToCheck); err == nil {
			return filepath.Clean(filepath.Join(candidateKubetestDir, "kubernetes")), nil
		}
	}

	return "", os.ErrNotExist
}
