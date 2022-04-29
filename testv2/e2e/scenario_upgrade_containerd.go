package e2e

import (
	"io"
	"testing"
)

type upgradeContainerd struct {
	versions []string
	params   []map[string]string
	infra    Infra
}

func (upgradeContainerd) Name() string           { return "upgrade_containerd" }
func (scenario upgradeContainerd) Title() string { return titelize(scenario.Name()) }

func (scenario *upgradeContainerd) SetInfra(infra Infra) {
	scenario.infra = infra
}

func (scenario *upgradeContainerd) SetVersions(versions ...string) {
	scenario.versions = versions
}

func (scenario *upgradeContainerd) SetParams(params []map[string]string) {
	scenario.params = params
}

func (scenario *upgradeContainerd) Run(t *testing.T) {
	t.Helper()
}

func (scenario *upgradeContainerd) GenerateTests(wr io.Writer) error {
	return nil
}
