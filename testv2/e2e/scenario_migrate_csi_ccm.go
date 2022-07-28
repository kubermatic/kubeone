package e2e

import (
	"io"
	"testing"
)

type scenarioMigrateCSIAndCCM struct {
	Name                    string
	OldManifestTemplatePath string
	NewManifestTemplatePath string

	versions []string
	infra    Infra
}

func (scenario scenarioMigrateCSIAndCCM) Title() string { return titleize(scenario.Name) }

func (scenario *scenarioMigrateCSIAndCCM) SetInfra(infra Infra) {
	scenario.infra = infra
}

func (scenario *scenarioMigrateCSIAndCCM) SetVersions(versions ...string) {
	scenario.versions = versions
}

func (scenario *scenarioMigrateCSIAndCCM) GenerateTests(wr io.Writer, generatorType GeneratorType, cfg ProwConfig) error {
	install := scenarioInstall{
		Name:     scenario.Name,
		infra:    scenario.infra,
		versions: scenario.versions,
	}

	return install.GenerateTests(wr, generatorType, cfg)
}

func (scenario *scenarioMigrateCSIAndCCM) Run(*testing.T) {

}
