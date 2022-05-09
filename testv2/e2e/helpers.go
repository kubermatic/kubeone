package e2e

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"text/template"

	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/util/retry"
)

var (
	Infrastructures = map[string]Infra{
		"aws_defaults": {
			name: "aws_defaults",
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"subnets_cidr=27",
				},
			},
		},
		"aws_centos": {
			name: "aws_centos",
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"subnets_cidr=27",
					"os=centos",
					"ssh_username=rocky",
					"bastion_user=rocky",
				},
			},
		},
	}

	Scenarios = map[string]Scenario{
		"install_docker": &install{
			name:                 "install_docker",
			manifestTemplatePath: "testdata/docker_simple.yaml",
		},
		"upgrade_docker": &upgrade{
			name:                 "upgrade_docker",
			manifestTemplatePath: "testdata/containerd_simple.yaml",
		},
		"conformance_docker": nil,
		"install_containerd": &install{
			name:                 "install_containerd",
			manifestTemplatePath: "testdata/containerd_simple.yaml",
		},
		"upgrade_containerd": &upgrade{
			name:                 "upgrade_containerd",
			manifestTemplatePath: "testdata/containerd_simple.yaml",
		},
		"conformance_containerd": nil,
		"weave_cni":              nil,
		"calico":                 nil,
	}
)

type Infra struct {
	name      string
	terraform terraformBin
}

type Scenario interface {
	SetInfra(Infra)
	SetVersions(...string)
	SetParams([]map[string]string)
	GenerateTests(io.Writer) error
	Run(*testing.T)
}

func titleize(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.Title(s)
	return strings.ReplaceAll(s, " ", "")
}

type templateData struct {
	Infra         string
	InfraTitle    string
	Scenario      string
	ScenarioTitle string
	Version       string
	VersionTitle  string
}

func clusterName() string {
	name, found := os.LookupEnv("BUILD_ID")
	if !found {
		name = rand.String(10)
	}

	return fmt.Sprintf("k1-%s", name)
}

func trueRetriable(error) bool {
	return true
}

func retryFn(fn func() error) error {
	return retry.OnError(retry.DefaultRetry, trueRetriable, fn)
}

func requiredTemplateFunc(warn string, input interface{}) (interface{}, error) {
	switch val := input.(type) {
	case nil:
		return val, fmt.Errorf(warn)
	case string:
		if val == "" {
			return val, fmt.Errorf(warn)
		}
	}

	return input, nil
}

type manifestData struct {
	VERSION string
}

func renderManifest(tmpDir, templatePath string, data manifestData) (string, error) {
	var buf bytes.Buffer

	tpl, err := template.New("").Parse(templatePath)
	if err != nil {
		return "", err
	}
	tpl.Funcs(template.FuncMap{
		"required": requiredTemplateFunc,
	})

	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}

	manifest, err := os.CreateTemp(tmpDir, "kubeone-*.yaml")
	if err != nil {
		return "", err
	}
	defer manifest.Close()

	manifestPath := manifest.Name()
	if err := os.WriteFile(manifestPath, buf.Bytes(), 0600); err != nil {
		return "", err
	}

	return manifestPath, nil
}
