package e2e

import (
	"io"
	"testing"
)

var (
	Infrastructures = map[string]Infra{
		"aws_defaults": {
			name: "aws_defaults",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-aws":     "true",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"subnets_cidr=27",
				},
			},
		},
		"aws_centos": {
			name: "aws_centos",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-aws":     "true",
			},
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
		"aws_rhel": {
			name: "aws_rhel",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-aws":     "true",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"subnets_cidr=27",
					"os=rhel",
					"ssh_username=ec2-user",
					"bastion_user=ec2-user",
				},
			},
		},
		"aws_flatcar": {
			name: "aws_flatcar",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-aws":     "true",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"subnets_cidr=27",
					"os=flatcar",
					"ssh_username=core",
					"bastion_user=core",
				},
			},
		},
		"aws_amzn": {
			name: "aws_amzn",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-aws":     "true",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"subnets_cidr=27",
					"os=amzn",
					"ssh_username=ec2-user",
					"bastion_user=ec2-user",
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
	labels    map[string]string
}

type GeneratorType int

const (
	GeneratorTypeGo   = 1
	GeneratorTypeYAML = 2
)

type Scenario interface {
	SetInfra(infrastructure Infra)
	SetVersions(version ...string)
	SetParams(params []map[string]string)
	GenerateTests(output io.Writer, testType GeneratorType) error
	Run(*testing.T)
}
