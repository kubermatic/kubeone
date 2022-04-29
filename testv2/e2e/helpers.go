package e2e

import (
	"io"
	"strings"
	"testing"
)

var (
	Infrastructures = map[string]Infra{
		"aws": {
			name: "aws",
			terraform: terraformConfig{
				path: "examples/terraform/aws",
			},
			manifestPath: "testv2/e2e/testdata/aws_simple.yaml",
		},
		"aws_centos": {
			name: "aws_centos",
			terraform: terraformConfig{
				path: "examples/terraform/aws",
				vars: []string{
					"os=centos",
					"ssh_username=rocky",
					"bastion_user=rocky",
				},
			},
			manifestPath: "testv2/e2e/testdata/aws_simple.yaml",
		},
	}

	Scenarios = map[string]Scenario{
		"install_docker":           nil,
		"upgrade_docker":           nil,
		"conformance_docker":       nil,
		installContainerd{}.Name(): &installContainerd{},
		upgradeContainerd{}.Name(): &upgradeContainerd{},
		"conformance_containerd":   nil,
		"weave_cni":                nil,
		"calico":                   nil,
	}
)

type Infra struct {
	name         string
	terraform    terraformConfig
	manifestPath string
}

type terraformConfig struct {
	path string
	vars []string
}

type Scenario interface {
	SetInfra(Infra)
	SetVersions(...string)
	SetParams([]map[string]string)
	GenerateTests(io.Writer) error
	Run(*testing.T)
}

func titelize(s string) string {
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
