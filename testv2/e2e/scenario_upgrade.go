package e2e

import (
	"fmt"
	"io"
	"testing"
	"text/template"
)

type upgrade struct {
	name                 string
	manifestTemplatePath string
	versions             []string
	params               []map[string]string
	infra                Infra
}

func (scenario upgrade) Title() string { return titleize(scenario.name) }

func (scenario *upgrade) SetInfra(infra Infra) {
	scenario.infra = infra
}

func (scenario *upgrade) SetVersions(versions ...string) {
	scenario.versions = versions
}

func (scenario *upgrade) SetParams(params []map[string]string) {
	scenario.params = params
}

func (scenario *upgrade) Run(t *testing.T) {
	t.Helper()
}

func (scenario *upgrade) GenerateTests(wr io.Writer) error {
	if len(scenario.params) != len(scenario.versions)-1 {
		return fmt.Errorf("expected %d params with versions to upgrade to", len(scenario.versions)-1)
	}

	var upgradeToVersions []string
	for _, param := range scenario.params {
		for _, v := range param {
			upgradeToVersions = append(upgradeToVersions, v)
		}
	}

	type upgradeFromTo struct {
		From string
		To   string
	}

	upgrades := []upgradeFromTo{}
	for i, upVersion := range upgradeToVersions {
		upgrades = append(upgrades, upgradeFromTo{
			To:   upVersion,
			From: scenario.versions[i],
		})
	}

	type templateData struct {
		Infra         string
		InfraTitle    string
		Scenario      string
		ScenarioTitle string
		FromVersion   string
		ToVersion     string
		VersionTitle  string
	}

	var data []templateData

	for _, up := range upgrades {
		data = append(data, templateData{
			Infra:         scenario.infra.name,
			InfraTitle:    titleize(scenario.infra.name),
			Scenario:      scenario.name,
			ScenarioTitle: scenario.Title(),
			FromVersion:   up.From,
			ToVersion:     up.To,
			VersionTitle:  fmt.Sprintf("From%s_To%s", titleize(up.From), titleize(up.To)),
		})
	}

	tpl, err := template.New("").Parse(upgradeScenarioTemplate)
	if err != nil {
		return err
	}

	return tpl.Execute(wr, data)
}

const upgradeScenarioTemplate = `
{{ range . }}
func Test{{ .InfraTitle }}{{ .ScenarioTitle }}{{ .VersionTitle }}(t *testing.T) {
	infra := Infrastructures["{{ .Infra }}"]
	scenario := Scenarios["{{ .Scenario }}"]
	scenario.SetInfra(infra)
	scenario.SetVersions("{{ .FromVersion }}", "{{ .ToVersion }}")
	scenario.Run(t)
}
{{ end -}}
`
