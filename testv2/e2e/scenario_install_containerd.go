package e2e

import (
	"io"
	"testing"
	"text/template"
)

type installContainerd struct {
	versions []string
	params   []map[string]string
	infra    Infra
}

func (installContainerd) Name() string           { return "install_containerd" }
func (scenario installContainerd) Title() string { return titelize(scenario.Name()) }

func (scenario *installContainerd) SetInfra(infra Infra) {
	scenario.infra = infra
}

func (scenario *installContainerd) SetVersions(versions ...string) {
	scenario.versions = versions
}

func (scenario *installContainerd) SetParams(params []map[string]string) {
	scenario.params = params
}

func (scenario *installContainerd) Run(t *testing.T) {
	t.Helper()
}

func (ic *installContainerd) GenerateTests(wr io.Writer) error {
	var data []templateData

	for _, version := range ic.versions {
		data = append(data, templateData{
			Infra:         ic.infra.name,
			InfraTitle:    titelize(ic.infra.name),
			Scenario:      ic.Name(),
			ScenarioTitle: ic.Title(),
			Version:       version,
			VersionTitle:  titelize(version),
		})
	}

	tpl, err := template.New("").Parse(installContainerdTemplate)
	if err != nil {
		return err
	}

	if err := tpl.Execute(wr, data); err != nil {
		return err
	}

	return nil
}

const installContainerdTemplate = `
{{ range . }}
func Test{{.InfraTitle}}{{.ScenarioTitle}}{{.VersionTitle}}(t *testing.T) {
	infra := Infrastructures["{{.Infra}}"]
	scenario := Scenarios["{{.Scenario}}"]
	scenario.SetInfra(infra)
	scenario.SetVersions("{{.Version}}")
	scenario.Run(t)
}
{{ end -}}
`
