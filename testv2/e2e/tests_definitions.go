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

//go:generate go run ../generator -file ../tests.yml -type go -output ./tests_test.go
//go:generate go run ../generator -file ../tests.yml -type yaml -output ./prow.yaml

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
		"install_docker": &scenarioInstall{
			name:                 "install_docker",
			manifestTemplatePath: "testdata/docker_simple.yaml",
		},
		"upgrade_docker": &scenarioUpgrade{
			name:                 "upgrade_docker",
			manifestTemplatePath: "testdata/containerd_simple.yaml",
		},
		"conformance_docker": &scenarioConformance{
			name:                 "conformance_docker",
			manifestTemplatePath: "testdata/containerd_simple.yaml",
		},
		"install_containerd": &scenarioInstall{
			name:                 "install_containerd",
			manifestTemplatePath: "testdata/containerd_simple.yaml",
		},
		"upgrade_containerd": &scenarioUpgrade{
			name:                 "upgrade_containerd",
			manifestTemplatePath: "testdata/containerd_simple.yaml",
		},
		"conformance_containerd": &scenarioConformance{
			name:                 "conformance_containerd",
			manifestTemplatePath: "testdata/containerd_simple.yaml",
		},
		"calico_containerd": &scenarioInstall{
			name:                 "calico_containerd",
			manifestTemplatePath: "testdata/containerd_calico.yaml",
		},
		"calico_docker": &scenarioInstall{
			name:                 "calico_docker",
			manifestTemplatePath: "testdata/docker_calico.yaml",
		},
		"weave_containerd": &scenarioInstall{
			name:                 "weave_containerd",
			manifestTemplatePath: "testdata/containerd_weave.yaml",
		},
		"weave_docker": &scenarioInstall{
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
	SetVersions(versions ...string)
	GenerateTests(output io.Writer, testType GeneratorType, cfg ProwConfig) error
	Run(*testing.T)
}

type ProwConfig struct {
	AlwaysRun bool
	Optional  bool
}
