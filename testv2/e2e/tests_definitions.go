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
		"calico_containerd": &install{
			name:                 "calico_containerd",
			manifestTemplatePath: "testdata/containerd_calico.yaml",
		},
		"calico_docker": &install{
			name:                 "calico_docker",
			manifestTemplatePath: "testdata/docker_calico.yaml",
		},
		"weave_containerd": &install{
			name:                 "weave_containerd",
			manifestTemplatePath: "testdata/containerd_weave.yaml",
		},
		"weave_docker": &install{
			name:                 "weave_docker",
			manifestTemplatePath: "testdata/docker_weave.yaml",
		},
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
