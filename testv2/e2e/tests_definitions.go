package e2e

import (
	"io"
	"testing"
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
