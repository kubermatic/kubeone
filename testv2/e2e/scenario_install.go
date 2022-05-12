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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type install struct {
	name                 string
	manifestTemplatePath string
	versions             []string
	params               []map[string]string
	infra                Infra
	hook                 func()
}

func (scenario install) Title() string { return titleize(scenario.name) }

func (scenario *install) SetInfra(infra Infra) {
	scenario.infra = infra
}

func (scenario *install) SetVersions(versions ...string) {
	scenario.versions = versions
}

func (scenario *install) SetParams(params []map[string]string) {
	scenario.params = params
}

func (scenario *install) Run(t *testing.T) {
	t.Helper()

	if len(scenario.versions) != 1 {
		t.Fatalf("only 1 version is expected to be set, got %v", scenario.versions)
	}

	clusterName := clusterName()
	tmpDir := t.TempDir()
	data := manifestData{
		VERSION: scenario.versions[0],
	}

	manifestPath, err := renderManifest(tmpDir, scenario.manifestTemplatePath, data)
	if err != nil {
		t.Fatalf("failed to render kubeone manifest: %v", err)
	}

	if err := scenario.infra.terraform.init(clusterName); err != nil {
		t.Fatalf("terraform init failed: %v", err)
	}

	if err := retryFn(scenario.infra.terraform.apply); err != nil {
		t.Fatalf("terraform apply failed: %v", err)
	}

	defer func() {
		if err := retryFn(func() error {
			return scenario.infra.terraform.destroy()
		}); err != nil {
			t.Fatalf("terraform destroy failed: %v", err)
		}
	}()

	k1 := kubeoneBin{
		bin:          "kubeone",
		dir:          scenario.infra.terraform.path,
		tfjsonPath:   ".",
		manifestPath: manifestPath,
	}

	if err := k1.Apply(); err != nil {
		t.Fatalf("kubeone apply failed: %v", err)
	}

	defer func() {
		if err := retryFn(func() error {
			return k1.Reset()
		}); err != nil {
			t.Fatalf("terraform destroy failed: %v", err)
		}
	}()

	kubeoneManifest, err := k1.Manifest()
	if err != nil {
		t.Fatalf("failed to get manifest API")
	}

	numberOfNodesToWait := len(kubeoneManifest.ControlPlane.Hosts) + len(kubeoneManifest.StaticWorkers.Hosts)
	for _, worker := range kubeoneManifest.DynamicWorkers {
		if worker.Replicas != nil {
			numberOfNodesToWait += *worker.Replicas
		}
	}

	var kubeconfig []byte
	fetchKubeconfig := func() error {
		kubeconfig, err = k1.Kubeconfig()
		if err != nil {
			return err
		}

		return nil
	}

	if err := retryFn(fetchKubeconfig); err != nil {
		t.Fatalf("kubeone kubeconfig failed: %v", err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		t.Fatalf("unable to build clientset from kubeconfig bytes: %v", err)
	}

	client, err := ctrlruntimeclient.New(restConfig, ctrlruntimeclient.Options{})
	if err != nil {
		t.Fatalf("failed to init dynamic client: %s", err)
	}

	if err = waitForNodesReady(t, client, numberOfNodesToWait); err != nil {
		t.Fatalf("failed to bring up all nodes up: %v", err)
	}

	if err = verifyVersion(client, metav1.NamespaceSystem, scenario.versions[0]); err != nil {
		t.Fatalf("version mismatch: %v", err)
	}

	if scenario.hook != nil {
		scenario.hook()
	}
}

func (scenario *install) GenerateTests(wr io.Writer, generatorType GeneratorType) error {
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

	for _, version := range scenario.versions {
		testTitle := fmt.Sprintf("Test%s%s%s",
			titleize(scenario.infra.name),
			scenario.Title(),
			titleize(version),
		)

		data = append(data, templateData{
			TestTitle: testTitle,
			Infra:     scenario.infra.name,
			Scenario:  scenario.name,
			Version:   version,
		})

		prowJobName := fmt.Sprintf("pull-%s-%s-%s",
			scenario.infra.name,
			scenario.name,
			version,
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
