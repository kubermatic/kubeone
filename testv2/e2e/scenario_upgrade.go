/*
Copyright 2022 The KubeOne Authors.

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
	"io"
	"testing"
	"text/template"

	"sigs.k8s.io/yaml"
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

func (scenario *upgrade) GenerateTests(wr io.Writer, generatorType GeneratorType) error {
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
		Infra       string
		Scenario    string
		FromVersion string
		ToVersion   string
		TestTitle   string
	}

	var (
		data     []templateData
		prowJobs []ProwJob
	)

	for _, up := range upgrades {
		testTitle := fmt.Sprintf("Test%s%sFrom%s_To%s",
			titleize(scenario.infra.name),
			scenario.Title(),
			titleize(up.From),
			titleize(up.To),
		)

		data = append(data, templateData{
			TestTitle:   testTitle,
			Infra:       scenario.infra.name,
			Scenario:    scenario.name,
			FromVersion: up.From,
			ToVersion:   up.To,
		})

		prowJobName := fmt.Sprintf("pull-%s-%s-from-%s-to-%s",
			scenario.infra.name,
			scenario.name,
			up.From,
			up.To,
		)

		prowJobs = append(prowJobs,
			newProwJob(
				prowJobName,
				scenario.infra.labels,
				testTitle,
			),
		)
	}

	switch generatorType {
	case GeneratorTypeGo:
		tpl, err := template.New("").Parse(upgradeScenarioTemplate)
		if err != nil {
			return err
		}

		return tpl.Execute(wr, data)
	case GeneratorTypeYAML:
		buf, err := yaml.Marshal(prowJobs)
		if err != nil {
			return err
		}

		n, err := wr.Write(buf)
		if err != nil {
			return err
		}

		if n != len(buf) {
			return fmt.Errorf("wrong number of bytes written, expected %d, wrote %d", len(buf), n)
		}

		return nil
	}

	return fmt.Errorf("unknown generator type %d", generatorType)
}

const upgradeScenarioTemplate = `
{{- range . }}
func {{ .TestTitle }}(t *testing.T) {
	infra := Infrastructures["{{ .Infra }}"]
	scenario := Scenarios["{{ .Scenario }}"]
	scenario.SetInfra(infra)
	scenario.SetVersions("{{ .FromVersion }}", "{{ .ToVersion }}")
	scenario.Run(t)
}
{{ end -}}
`
