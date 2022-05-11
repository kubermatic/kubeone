package e2e

import (
	"io"
	"testing"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	cntr "sigs.k8s.io/controller-runtime/pkg/client"
)

type install struct {
	name                 string
	manifestTemplatePath string
	versions             []string
	params               []map[string]string
	infra                Infra
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

	k1 := kubeoneBin{
		bin:          "kubeone",
		dir:          scenario.infra.terraform.path,
		tfjsonPath:   ".",
		manifestPath: manifestPath,
	}

	if err := k1.Apply(); err != nil {
		t.Fatalf("kubeone apply failed: %v", err)
	}

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

	client, err := cntr.New(restConfig, cntr.Options{})
	if err != nil {
		t.Fatalf("failed to init dynamic client: %s", err)
	}

	if err = waitForNodesReady(t, client, numberOfNodesToWait); err != nil {
		t.Fatalf("failed to bring up all nodes up: %v", err)
	}

	if err = verifyVersion(client, metav1.NamespaceSystem, scenario.versions[0]); err != nil {
		t.Fatalf("version mismatch: %v", err)
	}

	if err := retryFn(func() error {
		return k1.Reset()
	}); err != nil {
		t.Fatalf("terraform destroy failed: %v", err)
	}

	if err := retryFn(func() error {
		return scenario.infra.terraform.destroy()
	}); err != nil {
		t.Fatalf("terraform destroy failed: %v", err)
	}
}

func (scenario *install) GenerateTests(wr io.Writer) error {
	type templateData struct {
		Infra         string
		InfraTitle    string
		Scenario      string
		ScenarioTitle string
		Version       string
		VersionTitle  string
	}

	var data []templateData

	for _, version := range scenario.versions {
		data = append(data, templateData{
			Infra:         scenario.infra.name,
			InfraTitle:    titleize(scenario.infra.name),
			Scenario:      scenario.name,
			ScenarioTitle: scenario.Title(),
			Version:       version,
			VersionTitle:  titleize(version),
		})
	}

	tpl, err := template.New("").Parse(installScenarioTemplate)
	if err != nil {
		return err
	}

	return tpl.Execute(wr, data)
}

const installScenarioTemplate = `
{{- range . }}
func Test{{.InfraTitle}}{{.ScenarioTitle}}{{.VersionTitle}}(t *testing.T) {
	infra := Infrastructures["{{.Infra}}"]
	scenario := Scenarios["{{.Scenario}}"]
	scenario.SetInfra(infra)
	scenario.SetVersions("{{.Version}}")
	scenario.Run(t)
}
{{ end -}}
`
