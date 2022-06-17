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

type scenarioInstall struct {
	Name                 string
	ManifestTemplatePath string

	versions    []string
	infra       Infra
	kubeonePath string
}

func (scenario scenarioInstall) KubeonePath() string {
	if scenario.kubeonePath != "" {
		return scenario.kubeonePath
	}

	return getKubeoneDistPath()
}

func (scenario scenarioInstall) Title() string { return titleize(scenario.Name) }

func (scenario *scenarioInstall) SetInfra(infra Infra) {
	scenario.infra = infra
}

func (scenario *scenarioInstall) SetVersions(versions ...string) {
	scenario.versions = versions
}

func (scenario *scenarioInstall) Run(t *testing.T) {
	if err := makeBin("build").Run(); err != nil {
		t.Fatalf("building kubeone: %v", err)
	}

	scenario.install(t)
	scenario.test(t)
}

func (scenario *scenarioInstall) install(t *testing.T) {
	if len(scenario.versions) != 1 {
		t.Fatalf("only 1 version is expected to be set, got %v", scenario.versions)
	}

	if err := scenario.infra.terraform.Init(); err != nil {
		t.Fatalf("terraform init failed: %v", err)
	}

	t.Cleanup(func() {
		if err := retryFn(func() error {
			return scenario.infra.terraform.Destroy()
		}); err != nil {
			t.Fatalf("terraform destroy failed: %v", err)
		}
	})

	if err := retryFn(scenario.infra.terraform.Apply); err != nil {
		t.Fatalf("terraform apply failed: %v", err)
	}

	k1 := scenario.kubeone(t)

	t.Cleanup(func() {
		if err := retryFn(func() error {
			return k1.Reset()
		}); err != nil {
			t.Fatalf("terraform destroy failed: %v", err)
		}
	})

	if err := k1.Apply(); err != nil {
		t.Fatalf("kubeone apply failed: %v", err)
	}
}

func (scenario *scenarioInstall) kubeone(t *testing.T) *kubeoneBin {
	var k1Opts = []kubeoneBinOpts{
		withKubeoneBin(scenario.KubeonePath()),
	}

	if *kubeoneVerboseFlag {
		k1Opts = append(k1Opts, withKubeoneVerbose)
	}

	if *credentialsFlag != "" {
		k1Opts = append(k1Opts, withKubeoneCredentials(*credentialsFlag))
	}

	return newKubeoneBin(
		scenario.infra.terraform.path,
		renderManifest(t,
			scenario.ManifestTemplatePath,
			manifestData{
				VERSION: scenario.versions[0],
			},
		),
		k1Opts...,
	)
}

func (scenario *scenarioInstall) test(t *testing.T) {
	var (
		data = manifestData{VERSION: scenario.versions[0]}
		k1   = scenario.kubeone(t)
	)

	basicTest(t, k1, data)
	sonobuoyRun(t, k1, sonobuoyConformanceLite)
}

func (scenario *scenarioInstall) GenerateTests(wr io.Writer, generatorType GeneratorType, cfg ProwConfig) error {
	type templateData struct {
		TestTitle string
		Infra     string
		Scenario  string
		Version   string
	}

	var (
		data     []templateData
		prowJobs []ProwJob
	)

	version := scenario.versions[0]
	testTitle := fmt.Sprintf("Test%s%s%s",
		titleize(scenario.infra.name),
		scenario.Title(),
		titleize(version),
	)

	data = append(data, templateData{
		TestTitle: testTitle,
		Infra:     scenario.infra.name,
		Scenario:  scenario.Name,
		Version:   version,
	})

	cfg.Environ = scenario.infra.environ

	prowJobs = append(prowJobs,
		newProwJob(
			pullProwJobName(scenario.infra.name, scenario.Name, version),
			scenario.infra.labels,
			testTitle,
			cfg,
		),
	)

	switch generatorType {
	case GeneratorTypeGo:
		tpl, err := template.New("").Parse(installScenarioTemplate)
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

const installScenarioTemplate = `
{{- range . }}
func {{.TestTitle}}(t *testing.T) {
	infra := Infrastructures["{{.Infra}}"]
	scenario := Scenarios["{{.Scenario}}"]
	scenario.SetInfra(infra)
	scenario.SetVersions("{{.Version}}")
	scenario.Run(t)
}
{{ end -}}
`
