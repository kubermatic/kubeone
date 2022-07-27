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
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
					"subnets_cidr=27",
				},
			},
		},
		"aws_centos": {
			name: "aws_centos",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-aws":     "true",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
					"subnets_cidr=27",
					"os=centos",
					"ssh_username=rocky",
					"bastion_user=rocky",
				},
			},
		},
		"aws_rhel": {
			name: "aws_rhel",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-aws":     "true",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
					"subnets_cidr=27",
					"os=rhel",
					"ssh_username=ec2-user",
					"bastion_user=ec2-user",
					"bastion_type=t3.micro",
				},
			},
		},
		"aws_rockylinux": {
			name: "aws_rockylinux",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-aws":     "true",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
					"subnets_cidr=27",
					"os=rockylinux",
					"ssh_username=rocky",
					"bastion_user=rocky",
				},
			},
		},
		"aws_flatcar": {
			name: "aws_flatcar",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-aws":     "true",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
					"subnets_cidr=27",
					"os=flatcar",
					"ssh_username=core",
					"bastion_user=core",
				},
			},
		},
		"aws_amzn": {
			name: "aws_amzn",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-aws":     "true",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
					"subnets_cidr=27",
					"os=amzn",
					"ssh_username=ec2-user",
					"bastion_user=ec2-user",
				},
			},
		},
		"aws_long_timeout_default": {
			name: "aws_long_timeout_default",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-aws":     "true",
			},
			environ: map[string]string{
				"PROVIDER":     "aws",
				"TEST_TIMEOUT": "120m",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/aws",
				vars: []string{
					"subnets_cidr=27",
				},
			},
		},
		"azure_default": {
			name: "azure_default",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-azure":   "true",
			},
			environ: map[string]string{
				"PROVIDER":     "azure",
				"TEST_TIMEOUT": "120m",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/azure",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
				},
			},
		},
		"digitalocean_default": {
			name: "digitalocean_default",
			labels: map[string]string{
				"preset-goproxy":      "true",
				"preset-digitalocean": "true",
			},
			environ: map[string]string{
				"PROVIDER": "digitalocean",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/digitalocean",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
				},
			},
		},
		"equinixmetal_default": {
			name: "equinixmetal_default",
			labels: map[string]string{
				"preset-goproxy":       "true",
				"preset-equinix-metal": "true",
			},
			environ: map[string]string{
				"PROVIDER": "equinixmetal",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/equinixmetal",
			},
		},
		"hetzner_default": {
			name: "hetzner_default",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-hetzner": "true",
			},
			environ: map[string]string{
				"PROVIDER": "hetzner",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/hetzner",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
				},
			},
		},
		"gce_default": {
			name: "gce_default",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-gce":     "true",
			},
			environ: map[string]string{
				"PROVIDER": "gce",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/gce",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
				},
			},
		},
		"openstack_default": {
			name: "openstack_default",
			labels: map[string]string{
				"preset-goproxy":   "true",
				"preset-openstack": "true",
			},
			environ: map[string]string{
				"PROVIDER": "openstack",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/openstack",
				varFile: "testdata/openstack_vars.tfvars",
			},
		},
		// "vcd_default": {
		// 	name: "vcd_default",
		// 	labels: map[string]string{
		// 		"preset-goproxy": "true",
		// 		"preset-vcd":     "true",
		// 	},
		// 	environ: map[string]string{
		// 		"PROVIDER": "vcd",
		// 	},
		// 	terraform: terraformBin{
		// 		path: "../../examples/terraform/vmware-cloud-director",
		// 	},
		// },
		"vsphere_default": {
			name: "vsphere_default",
			labels: map[string]string{
				"preset-goproxy":        "true",
				"preset-vsphere-legacy": "true",
			},
			environ: map[string]string{
				"PROVIDER": "vsphere",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/vsphere",
				varFile: "testdata/vsphere.tfvars",
				vars: []string{
					"template_name=kubeone-e2e-ubuntu",
					"worker_os=ubuntu",
					"ssh_username=ubuntu",
				},
			},
		},
		"vsphere_flatcar": {
			name: "vsphere_flatcar",
			labels: map[string]string{
				"preset-goproxy":        "true",
				"preset-vsphere-legacy": "true",
			},
			environ: map[string]string{
				"PROVIDER": "vsphere",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/vsphere_flatcar",
				varFile: "testdata/vsphere.tfvars",
				vars: []string{
					"template_name=machine-controller-e2e-flatcar",
				},
			},
		},
		// TODO
		// "vsphere_rhel": {
		// 	name: "vsphere_rhel",
		// 	labels: map[string]string{
		// 		"preset-goproxy": "true",
		// 		"preset-vsphere-legacy": "true",
		// 	},
		// 	environ: map[string]string{
		// 		"PROVIDER": "vsphere",
		// 	},
		// 	terraform: terraformBin{
		// 		path:    "../../examples/terraform/vsphere",
		// 		varFile: "testdata/vsphere.tfvars",
		// 		vars: []string{
		// 			"template_name=machine-controller-e2e-rhel",
		// 			"worker_os=rhel",
		// 			"ssh_username=rhel",
		// 			"disk_size=50",
		// 		},
		// 	},
		// },
	}

	Scenarios = map[string]Scenario{
		// docker
		"install_docker": &scenarioInstall{
			Name:                 "install_docker",
			ManifestTemplatePath: "testdata/docker_simple.yaml",
		},
		"upgrade_docker": &scenarioUpgrade{
			name:                 "upgrade_docker",
			manifestTemplatePath: "testdata/docker_simple.yaml",
		},
		"conformance_docker": &scenarioConformance{
			name:                 "conformance_docker",
			manifestTemplatePath: "testdata/docker_simple.yaml",
		},

		// containerd
		"install_containerd": &scenarioInstall{
			Name:                 "install_containerd",
			ManifestTemplatePath: "testdata/containerd_simple.yaml",
		},
		"upgrade_containerd": &scenarioUpgrade{
			name:                 "upgrade_containerd",
			manifestTemplatePath: "testdata/containerd_simple.yaml",
		},
		"conformance_containerd": &scenarioConformance{
			name:                 "conformance_containerd",
			manifestTemplatePath: "testdata/containerd_simple.yaml",
		},

		// docker external
		"install_docker_external": &scenarioInstall{
			Name:                 "install_docker_external",
			ManifestTemplatePath: "testdata/docker_simple_external.yaml",
		},
		"upgrade_docker_external": &scenarioUpgrade{
			name:                 "upgrade_docker_external",
			manifestTemplatePath: "testdata/docker_simple_external.yaml",
		},
		"conformance_docker_external": &scenarioConformance{
			name:                 "conformance_docker_external",
			manifestTemplatePath: "testdata/docker_simple_external.yaml",
		},

		// external containerd
		"install_containerd_external": &scenarioInstall{
			Name:                 "install_containerd_external",
			ManifestTemplatePath: "testdata/containerd_simple_external.yaml",
		},
		"upgrade_containerd_external": &scenarioUpgrade{
			name:                 "upgrade_containerd_external",
			manifestTemplatePath: "testdata/containerd_simple_external.yaml",
		},
		"conformance_containerd_external": &scenarioConformance{
			name:                 "conformance_containerd_external",
			manifestTemplatePath: "testdata/containerd_simple_external.yaml",
		},

		// Various features
		"calico_containerd": &scenarioInstall{
			Name:                 "calico_containerd",
			ManifestTemplatePath: "testdata/containerd_calico.yaml",
		},
		"calico_docker": &scenarioInstall{
			Name:                 "calico_docker",
			ManifestTemplatePath: "testdata/docker_calico.yaml",
		},
		"weave_containerd": &scenarioInstall{
			Name:                 "weave_containerd",
			ManifestTemplatePath: "testdata/containerd_weave.yaml",
		},
		"weave_docker": &scenarioInstall{
			Name:                 "weave_docker",
			ManifestTemplatePath: "testdata/docker_weave.yaml",
		},
		"cilium_containerd": &scenarioInstall{
			Name:                 "cilium_containerd",
			ManifestTemplatePath: "testdata/containerd_cilium.yaml",
		},
		"cilium_docker": &scenarioInstall{
			Name:                 "cilium_docker",
			ManifestTemplatePath: "testdata/docker_cilium.yaml",
		},
		"install_operating_system_manager": &scenarioInstall{
			Name:                 "install_operating_system_manager",
			ManifestTemplatePath: "testdata/operating_system_manager.yaml",
		},
		"kube_proxy_ipvs": &scenarioInstall{
			Name:                 "kube_proxy_ipvs",
			ManifestTemplatePath: "testdata/kube_proxy_ipvs.yaml",
		},
	}
)

type Infra struct {
	name      string
	environ   map[string]string
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
	Environ   map[string]string
}
