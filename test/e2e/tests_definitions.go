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
//go:generate go run ../generator -file ../tests.yml -type yaml -output ./../../.prow/generated.yaml

package e2e

import (
	"context"
	"io"
	"testing"
)

var (
	Infrastructures = map[string]Infra{
		"aws_default": {
			name:   "aws_default",
			osName: "Ubuntu",
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/aws",
				varFile: "testdata/aws_small.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"aws_default_stable": {
			name:   "aws_default_stable",
			osName: "Ubuntu",
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			terraform: terraformBin{
				path:    "../../../kubeone-stable/examples/terraform/aws",
				varFile: "testdata/aws_stable_small.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"aws_ubuntu_previous_lts": {
			name:   "aws_ubuntu_previous_lts",
			osName: "Ubuntu (Previous LTS)",
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/aws",
				varFile: "testdata/aws_small_ubuntu_previous_lts.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"aws_rhel": {
			name:   "aws_rhel",
			osName: "RHEL",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/aws",
				varFile: "testdata/aws_rhel.tfvars",
				vars: []string{
					"worker_volume_size=50",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"aws_rhel_stable": {
			name:   "aws_rhel_stable",
			osName: "RHEL",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			terraform: terraformBin{
				path:    "../../../kubeone-stable/examples/terraform/aws",
				varFile: "testdata/aws_rhel.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"aws_rockylinux": {
			name:   "aws_rockylinux",
			osName: "Rocky Linux",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/aws",
				varFile: "testdata/aws_medium.tfvars",
				vars: []string{
					"os=rockylinux",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"aws_rockylinux_stable": {
			name:   "aws_rockylinux_stable",
			osName: "Rocky Linux",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			terraform: terraformBin{
				path:    "../../../kubeone-stable/examples/terraform/aws",
				varFile: "testdata/aws_stable_medium.tfvars",
				vars: []string{
					"os=rockylinux",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"aws_flatcar": {
			name:   "aws_flatcar",
			osName: "Flatcar",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/aws",
				varFile: "testdata/aws_medium.tfvars",
				vars: []string{
					"disable_auto_update=true",
					"os=flatcar",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"aws_flatcar_stable": {
			name:   "aws_flatcar_stable",
			osName: "Flatcar",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			terraform: terraformBin{
				path:    "../../../kubeone-stable/examples/terraform/aws",
				varFile: "testdata/aws_stable_medium.tfvars",
				vars: []string{
					"disable_auto_update=true",
					"os=flatcar",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"aws_flatcar_cloud_init": {
			name:   "aws_flatcar_cloud_init",
			osName: "Flatcar",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/aws",
				varFile: "testdata/aws_medium.tfvars",
				vars: []string{
					"disable_auto_update=true",
					"os=flatcar",
					"worker_deploy_ssh_key=false",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"aws_flatcar_cloud_init_stable": {
			name:   "aws_flatcar_cloud_init_stable",
			osName: "Flatcar",
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			terraform: terraformBin{
				path:    "../../../kubeone-stable/examples/terraform/aws",
				varFile: "testdata/aws_stable_medium.tfvars",
				vars: []string{
					"disable_auto_update=true",
					"os=flatcar",
					"worker_deploy_ssh_key=false",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"aws_long_timeout_default": {
			name:              "aws_long_timeout_default",
			osName:            "Ubuntu",
			discludeFromIssue: true,
			labels: map[string]string{
				"preset-goproxy":         "true",
				"preset-aws-e2e-kubeone": "true",
			},
			environ: map[string]string{
				"PROVIDER": "aws",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/aws",
				varFile: "testdata/aws_medium.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"azure_default": {
			name:   "azure_default",
			osName: "Ubuntu",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-azure":   "true",
			},
			environ: map[string]string{
				"PROVIDER":     "azure",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/azure",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"azure_default_stable": {
			name:   "azure_default_stable",
			osName: "Ubuntu",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-azure":   "true",
			},
			environ: map[string]string{
				"PROVIDER":     "azure",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path: "../../../kubeone-stable/examples/terraform/azure",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"azure_flatcar": {
			name:   "azure_flatcar",
			osName: "Flatcar",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-azure":   "true",
			},
			environ: map[string]string{
				"PROVIDER":     "azure",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/azure",
				vars: []string{
					"disable_auto_update=true",
					"os=flatcar",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"azure_flatcar_stable": {
			name:   "azure_flatcar_stable",
			osName: "Flatcar",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-azure":   "true",
			},
			environ: map[string]string{
				"PROVIDER":     "azure",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path: "../../../kubeone-stable/examples/terraform/azure",
				vars: []string{
					"disable_auto_update=true",
					"os=flatcar",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"azure_rhel": {
			name:   "azure_rhel",
			osName: "RHEL",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-azure":   "true",
				"preset-rhel":    "true",
			},
			environ: map[string]string{
				"PROVIDER":     "azure",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/azure",
				vars: []string{
					"os=rhel",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"azure_rhel_stable": {
			name:   "azure_rhel_stable",
			osName: "RHEL",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-azure":   "true",
				"preset-rhel":    "true",
			},
			environ: map[string]string{
				"PROVIDER":     "azure",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path: "../../../kubeone-stable/examples/terraform/azure",
				vars: []string{
					"os=rhel",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"azure_rockylinux": {
			name:   "azure_rockylinux",
			osName: "Rocky Linux",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-azure":   "true",
			},
			environ: map[string]string{
				"PROVIDER":     "azure",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path: "../../examples/terraform/azure",
				vars: []string{
					"os=rockylinux",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"azure_rockylinux_stable": {
			name:   "azure_rockylinux_stable",
			osName: "Rocky Linux",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-azure":   "true",
			},
			environ: map[string]string{
				"PROVIDER":     "azure",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path: "../../../kubeone-stable/examples/terraform/azure",
				vars: []string{
					"os=rockylinux",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"digitalocean_default": {
			name:   "digitalocean_default",
			osName: "Ubuntu",
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
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"digitalocean_default_stable": {
			name:   "digitalocean_default_stable",
			osName: "Ubuntu",
			labels: map[string]string{
				"preset-goproxy":      "true",
				"preset-digitalocean": "true",
			},
			environ: map[string]string{
				"PROVIDER": "digitalocean",
			},
			terraform: terraformBin{
				path: "../../../kubeone-stable/examples/terraform/digitalocean",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"digitalocean_rockylinux": {
			name:   "digitalocean_rockylinux",
			osName: "Rocky Linux",
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
					"os=rockylinux",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"digitalocean_rockylinux_stable": {
			name:   "digitalocean_rockylinux_stable",
			osName: "Rocky Linux",
			labels: map[string]string{
				"preset-goproxy":      "true",
				"preset-digitalocean": "true",
			},
			environ: map[string]string{
				"PROVIDER": "digitalocean",
			},
			terraform: terraformBin{
				path: "../../../kubeone-stable/examples/terraform/digitalocean",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
					"control_plane_droplet_image=rockylinux-8-x64",
					"worker_os=rockylinux",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"gce_default": {
			name:   "gce_default",
			osName: "Ubuntu",
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
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"gce_default_stable": {
			name:   "gce_default_stable",
			osName: "Ubuntu",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-gce":     "true",
			},
			environ: map[string]string{
				"PROVIDER": "gce",
			},
			terraform: terraformBin{
				path: "../../../kubeone-stable/examples/terraform/gce",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"hetzner_default": {
			name:   "hetzner_default",
			osName: "Ubuntu",
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
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"hetzner_default_stable": {
			name:   "hetzner_default_stable",
			osName: "Ubuntu",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-hetzner": "true",
			},
			environ: map[string]string{
				"PROVIDER": "hetzner",
			},
			terraform: terraformBin{
				path: "../../../kubeone-stable/examples/terraform/hetzner",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"hetzner_rockylinux": {
			name:   "hetzner_rockylinux",
			osName: "Rocky Linux",
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
					"os=rockylinux",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"hetzner_rockylinux_stable": {
			name:   "hetzner_rockylinux_stable",
			osName: "Rocky Linux",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-hetzner": "true",
			},
			environ: map[string]string{
				"PROVIDER": "hetzner",
			},
			terraform: terraformBin{
				path: "../../../kubeone-stable/examples/terraform/hetzner",
				vars: []string{
					"disable_kubeapi_loadbalancer=true",
					"image=rocky-8",
					"worker_os=rockylinux",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"openstack_default": {
			name:   "openstack_default",
			osName: "Ubuntu",
			labels: map[string]string{
				"preset-goproxy":   "true",
				"preset-openstack": "true",
			},
			environ: map[string]string{
				"PROVIDER": "openstack",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/openstack",
				varFile: "testdata/openstack_ubuntu.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"openstack_default_stable": {
			name:   "openstack_default_stable",
			osName: "Ubuntu",
			labels: map[string]string{
				"preset-goproxy":   "true",
				"preset-openstack": "true",
			},
			environ: map[string]string{
				"PROVIDER":     "openstack",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path:    "../../../kubeone-stable/examples/terraform/openstack",
				varFile: "testdata/openstack_ubuntu.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"openstack_rockylinux": {
			name:   "openstack_rockylinux",
			osName: "Rocky Linux",
			labels: map[string]string{
				"preset-goproxy":   "true",
				"preset-openstack": "true",
			},
			environ: map[string]string{
				"PROVIDER":     "openstack",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/openstack",
				varFile: "testdata/openstack_rockylinux.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"openstack_rockylinux_stable": {
			name:   "openstack_rockylinux_stable",
			osName: "Rocky Linux",
			labels: map[string]string{
				"preset-goproxy":   "true",
				"preset-openstack": "true",
			},
			environ: map[string]string{
				"PROVIDER":     "openstack",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path:    "../../../kubeone-stable/examples/terraform/openstack",
				varFile: "testdata/openstack_rockylinux.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"openstack_rhel": {
			name:   "openstack_rhel",
			osName: "RHEL",
			labels: map[string]string{
				"preset-goproxy":   "true",
				"preset-openstack": "true",
			},
			environ: map[string]string{
				"PROVIDER":     "openstack",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/openstack",
				varFile: "testdata/openstack_rhel.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"openstack_rhel_stable": {
			name:   "openstack_rhel_stable",
			osName: "RHEL",
			labels: map[string]string{
				"preset-goproxy":   "true",
				"preset-openstack": "true",
			},
			environ: map[string]string{
				"PROVIDER":     "openstack",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path:    "../../../kubeone-stable/examples/terraform/openstack",
				varFile: "testdata/openstack_rhel.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"openstack_flatcar": {
			name:   "openstack_flatcar",
			osName: "Flatcar",
			labels: map[string]string{
				"preset-goproxy":   "true",
				"preset-openstack": "true",
			},
			environ: map[string]string{
				"PROVIDER":     "openstack",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/openstack",
				varFile: "testdata/openstack_flatcar.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"openstack_flatcar_stable": {
			name:   "openstack_flatcar_stable",
			osName: "Flatcar",
			labels: map[string]string{
				"preset-goproxy":   "true",
				"preset-openstack": "true",
			},
			environ: map[string]string{
				"PROVIDER":     "openstack",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path:    "../../../kubeone-stable/examples/terraform/openstack",
				varFile: "testdata/openstack_flatcar.tfvars",
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
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
		// 	protokol: protokolBin{
		// 		namespaces: []string{"kube-system"},
		// 		outputDir:  "/logs/artifacts/logs",
		// 	},
		// },
		"vsphere_default": {
			name:   "vsphere_default",
			osName: "Ubuntu",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-vsphere": "true",
			},
			environ: map[string]string{
				"PROVIDER":     "vsphere",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/vsphere",
				varFile: "testdata/vsphere.tfvars",
				vars: []string{
					"worker_os=ubuntu",
					"ssh_username=ubuntu",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"vsphere_default_stable": {
			name:   "vsphere_default_stable",
			osName: "Ubuntu",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-vsphere": "true",
			},
			environ: map[string]string{
				"PROVIDER":     "vsphere",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path:    "../../../kubeone-stable/examples/terraform/vsphere",
				varFile: "testdata/vsphere.tfvars",
				vars: []string{
					"worker_os=ubuntu",
					"ssh_username=ubuntu",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"vsphere_flatcar": {
			name:   "vsphere_flatcar",
			osName: "Flatcar",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-vsphere": "true",
			},
			environ: map[string]string{
				"PROVIDER":     "vsphere",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path:    "../../examples/terraform/vsphere_flatcar",
				varFile: "testdata/vsphere.tfvars",
				vars: []string{
					"disable_auto_update=true",
					"template_name=kubeone-e2e-flatcar",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
			},
		},
		"vsphere_flatcar_stable": {
			name:   "vsphere_flatcar_stable",
			osName: "Flatcar",
			labels: map[string]string{
				"preset-goproxy": "true",
				"preset-vsphere": "true",
			},
			environ: map[string]string{
				"PROVIDER":     "vsphere",
				"TEST_TIMEOUT": "180m",
			},
			terraform: terraformBin{
				path:    "../../../kubeone-stable/examples/terraform/vsphere_flatcar",
				varFile: "testdata/vsphere.tfvars",
				vars: []string{
					"disable_auto_update=true",
					"template_name=kubeone-e2e-flatcar",
				},
			},
			protokol: protokolBin{
				namespaces: []string{"kube-system"},
				outputDir:  "/logs/artifacts/logs",
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
		// containerd
		"install_containerd": &scenarioInstall{
			Name:                 "install_containerd",
			HumanReadableName:    "Cluster Provisioning (In-Tree)",
			ManifestTemplatePath: "testdata/containerd_simple.yaml",
		},

		// external containerd
		"install_containerd_external": &scenarioInstall{
			Name:                 "install_containerd_external",
			HumanReadableName:    "Cluster Provisioning (External)",
			ManifestTemplatePath: "testdata/containerd_simple_external.yaml",
		},
		"upgrade_containerd_external": &scenarioUpgrade{
			Name:                 "upgrade_containerd_external",
			HumanReadableName:    "Cluster Upgrading (External)",
			ManifestTemplatePath: "testdata/containerd_simple_external_v1beta2.yaml",
		},
		"conformance_containerd_external": &scenarioConformance{
			Name:                 "conformance_containerd_external",
			HumanReadableName:    "Cluster Conformance (External)",
			ManifestTemplatePath: "testdata/containerd_simple_external.yaml",
		},

		// Various features
		"cilium_containerd_external": &scenarioInstall{
			Name:                 "cilium_containerd_external",
			HumanReadableName:    "Cluster Provisioning (External, Cilium)",
			ManifestTemplatePath: "testdata/containerd_cilium_external.yaml",
		},
		"upgrade_cilium_containerd_external": &scenarioUpgrade{
			Name:                 "upgrade_cilium_containerd_external",
			HumanReadableName:    "Cluster Upgrading (External, Calico)",
			ManifestTemplatePath: "testdata/containerd_cilium_external_v1beta2.yaml",
		},
		"kube_proxy_ipvs_external": &scenarioInstall{
			Name:                 "kube_proxy_ipvs_external",
			HumanReadableName:    "Cluster Provisioning (External, kube-proxy IPVS)",
			ManifestTemplatePath: "testdata/kube_proxy_ipvs_external.yaml",
		},
		"external_cni_flannel_helm_chart": &scenarioInstall{
			Name:                 "external_cni_flannel_helm_chart",
			HumanReadableName:    "Cluster Provisioning (External, Flannel via Helm)",
			ManifestTemplatePath: "testdata/containerd_flannel_helm_external.yaml",
		},
	}
)

type Infra struct {
	name              string
	osName            string
	discludeFromIssue bool
	environ           map[string]string
	terraform         terraformBin
	protokol          protokolBin
	labels            map[string]string
}

func (i Infra) Provider() string {
	return i.environ["PROVIDER"]
}

func (i Infra) OperatingSystem() string {
	return i.osName
}

func (i Infra) DiscludeFromIssue() bool {
	return i.discludeFromIssue
}

type GeneratorType int

const (
	GeneratorTypeGo     = 1
	GeneratorTypeYAML   = 2
	GeneratorTypeGitHub = 3
)

type Scenario interface {
	GetHumanReadableName() string
	SetInfra(infrastructure Infra)
	SetVersions(versions ...string)
	FetchVersions() error
	GenerateTests(output io.Writer, testType GeneratorType, cfg ProwConfig) error
	Run(context.Context, *testing.T)
}

type ScenarioStable interface {
	SetInitKubeOneVersion(version string)
}

type ProwConfig struct {
	AlwaysRun    bool
	RunIfChanged string
	Optional     bool
	Environ      map[string]string
}
