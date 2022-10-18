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

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/cobra"

	"k8c.io/kubeone/examples"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	"k8c.io/kubeone/pkg/fail"

	"k8s.io/apimachinery/pkg/util/sets"
	kyaml "sigs.k8s.io/yaml"
)

type initProvider struct {
	alternativeName string
	terraformPath   string
	external        bool
	cloudConfig     string
	csiConfig       string
	requiredTFVars  []string
}

var (
	validProviders = map[string]initProvider{
		"aws": {
			terraformPath: "terraform/aws",
			external:      true,
		},
		"azure": {
			terraformPath: "terraform/azure",
			external:      true,
			cloudConfig: heredoc.Doc(`
				{
				    "tenantId": "{{ .Credentials.AZURE_TENANT_ID }}",
				    "subscriptionId": "{{ .Credentials.AZURE_SUBSCRIPTION_ID }}",
				    "aadClientId": "{{ .Credentials.AZURE_CLIENT_ID }}",
				    "aadClientSecret": "{{ .Credentials.AZURE_CLIENT_SECRET }}",
				    "resourceGroup": "<RESOURCE-GROUP>",
				    "location": "<LOCATION>",
				    "subnetName": "<SUBNET>",
				    "routeTableName": "",
				    "securityGroupName": "<SECURITY-GROUP>",
				    "vnetName": "<VPC-NAME>",
				    "primaryAvailabilitySetName": "<AVAILABILITY-SET-NAME>",
				    "useInstanceMetadata": true,
				    "useManagedIdentityExtension": false,
				    "userAssignedIdentityID": ""
				}
			`),
		},
		"digitalocean": {
			terraformPath: "terraform/digitalocean",
			external:      true,
		},
		"equinixmetal": {
			terraformPath: "terraform/equinixmetal",
			external:      true,
			requiredTFVars: []string{
				"project_id = ",
			},
		},
		"gce": {
			terraformPath: "terraform/gce",
			requiredTFVars: []string{
				"project = ",
			},
		},
		"hetzner": {
			terraformPath: "terraform/hetzner",
			external:      true,
		},
		"none": {
			terraformPath: "",
		},
		"nutanix": {
			terraformPath: "terraform/nutanix",
			requiredTFVars: []string{
				"nutanix_cluster_name = ",
				"project_name = ",
				"subnet_name = ",
				"image_name = ",
			},
		},
		"openstack": {
			terraformPath: "terraform/openstack",
			external:      true,
			requiredTFVars: []string{
				"external_network_name = ",
				"image = ",
				"subnet_cidr = ",
			},
			cloudConfig: heredoc.Doc(`
				[Global]
				auth-url=<KEYSTONE-URL>
				username=<USER>
				password=<PASSWORD>
				tenant-id=<TENANT-ID>
				domain-name=DEFAULT
				region=<REGION>

				[LoadBalancer]
				[BlockStorage]
			`),
		},
		"vmware-cloud-director": {
			alternativeName: "vmwareCloudDirector",
			terraformPath:   "terraform/vmware-cloud-director",
			requiredTFVars: []string{
				"vcd_vdc_name = ",
				"vcd_edge_gateway_name = ",
				"catalog_name = ",
				"template_name = ",
			},
		},
		"vsphere": {
			terraformPath: "terraform/vsphere",
			external:      true,
			requiredTFVars: []string{
				"datastore_name = ",
				"network_name = ",
				"template_name = ",
				"resource_pool_name = ",
			},
			cloudConfig: heredoc.Doc(`
				[Global]
				secret-name = "vsphere-ccm-credentials"
				secret-namespace = "kube-system"
				port = "443"
				insecure-flag = "0"

				[VirtualCenter "<VCENTER-ADDRESS>"]

				[Workspace]
				server = "<VCENTER-ADDRESS>"
				datacenter = "<DATACENTER>"
				default-datastore="<DATASTORE>"
				resourcepool-path=""
				folder = ""

				[Disk]
				scsicontrollertype = pvscsi

				[Network]
				public-network = "<VM-NETWORK>"
			`),
			csiConfig: heredoc.Doc(`
				[Global]
				cluster-id = "<CLUSTER-ID>"
				cluster-distribution = "<CLUSTER-DISTRIBUTION>"

				[VirtualCenter "<VCENTER-ADDRESS>"]
				insecure-flag = "false"
				user = "<USERNAME>"
				password = "<PASSWORD>"
				port = "<PORT>"
				datacenters = "<DATACENTER>"
			`),
		},
		"vsphere/flatcar": {
			alternativeName: "vsphere",
			terraformPath:   "terraform/vsphere_flatcar",
			external:        true,
			cloudConfig: heredoc.Doc(`
				[Global]
				secret-name = "vsphere-ccm-credentials"
				secret-namespace = "kube-system"
				port = "443"
				insecure-flag = "0"

				[VirtualCenter "<VCENTER-ADDRESS>"]

				[Workspace]
				server = "<VCENTER-ADDRESS>"
				datacenter = "<DATACENTER>"
				default-datastore="<DATASTORE>"
				resourcepool-path=""
				folder = ""

				[Disk]
				scsicontrollertype = pvscsi

				[Network]
				public-network = "<VM-NETWORK>"
			`),
			csiConfig: heredoc.Doc(`
				[Global]
				cluster-id = "<CLUSTER-ID>"
				cluster-distribution = "<CLUSTER-DISTRIBUTION>"

				[VirtualCenter "<VCENTER-ADDRESS>"]
				insecure-flag = "false"
				user = "<USERNAME>"
				password = "<PASSWORD>"
				port = "<PORT>"
				datacenters = "<DATACENTER>"
			`),
		},
	}
)

type initOpts struct {
	Provider          oneOfFlag `longflag:"provider"`
	ClusterName       string    `longflag:"cluster-name"`
	KubernetesVersion string    `longflag:"kubernetes-version"`
	Terraform         bool      `longflag:"terraform"`
	Path              string    `longflag:"path"`
}

func initCmd() *cobra.Command {
	opts := &initOpts{
		Provider: oneOfFlag{
			validSet:     sets.StringKeySet(validProviders),
			defaultValue: "none",
		},
	}

	clusterNameFlag := longFlagName(opts, "ClusterName")
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init new kubeone cluster configuration",
		Long: heredoc.Doc(`
			Initialize new KubeOne + terraform configuration for chosen provider.
		`),
		SilenceErrors: true,
		Example:       `kubeone init --provider aws`,
		RunE: func(_ *cobra.Command, args []string) error {
			if opts.KubernetesVersion == "" {
				return fail.Runtime(fmt.Errorf("--kubernetes-version is a required flag"), "flag validation")
			}

			return runInit(opts)
		},
	}

	providerUsageText := fmt.Sprintf("provider to initialize, possible values: %s", strings.Join(opts.Provider.PossibleValues(), ", "))

	cmd.Flags().BoolVar(&opts.Terraform, longFlagName(opts, "Terraform"), true, "generate terraform config")
	cmd.Flags().StringVar(&opts.ClusterName, clusterNameFlag, "", "name of the cluster")
	cmd.Flags().StringVar(&opts.KubernetesVersion, longFlagName(opts, "KubernetesVersion"), defaultKubeVersion, "kubernetes version")
	cmd.Flags().StringVar(&opts.Path, longFlagName(opts, "Path"), ".", "path where to write files")
	cmd.Flags().Var(&opts.Provider, longFlagName(opts, "Provider"), providerUsageText)

	if err := cmd.MarkFlagRequired(clusterNameFlag); err != nil {
		panic(err)
	}

	return cmd
}

func runInit(opts *initOpts) error {
	providerName := opts.Provider.String()
	clusterName := opts.ClusterName

	if opts.Terraform && providerName != "none" {
		clusterName = ""
	}

	ybuf, err := genKubeOneClusterYAML(&genKubeOneClusterYAMLParams{
		providerName:      providerName,
		clusterName:       clusterName,
		kubernetesVersion: opts.KubernetesVersion,
		validProviders:    validProviders,
	})
	if err != nil {
		return fail.Runtime(err, "generating KubeOneCluster")
	}

	// special case to generate JUST yaml and no terraform
	if opts.Path == "-" && !opts.Terraform {
		_, err = fmt.Printf("%s", ybuf)

		return err
	}

	err = os.MkdirAll(opts.Path, 0750)
	if err != nil {
		return err
	}

	k1config, err := os.Create(filepath.Join(opts.Path, "kubeone.yaml"))
	if err != nil {
		return fail.Runtime(err, "creating manifest file")
	}
	defer k1config.Close()

	_, err = io.Copy(k1config, bytes.NewBuffer(ybuf))
	if err != nil {
		return fail.Runtime(err, "writing KubeOneCluster")
	}

	prov := validProviders[opts.Provider.String()]
	if opts.Terraform && prov.terraformPath != "" {
		if err = examples.CopyTo(opts.Path, prov.terraformPath); err != nil {
			return fail.Runtime(err, "copying terraform configuration")
		}

		tfvars, err := os.Create(filepath.Join(opts.Path, "terraform.tfvars"))
		if err != nil {
			return err
		}
		defer tfvars.Close()

		fmt.Fprintf(tfvars, "cluster_name = %q\n", opts.ClusterName)

		for _, param := range prov.requiredTFVars {
			fmt.Fprintf(tfvars, "%s\n", param)
		}
	}

	return nil
}

var (
	manifestTemplateSource = heredoc.Doc(`
		apiVersion: {{ .APIVersion }}
		kind: {{ .Kind }}
		{{- with .Name}}
		name: {{ . }}
		{{- end }}

		cloudProvider:
		  {{ .CloudProvider.Name }}: {}
		{{- with .CloudProvider.External }}
		  external: true
		{{ end -}}
		{{- with .CloudProvider.CloudConfig }}
		  cloudConfig: |
		{{ . | indent 4 -}}
		{{ end -}}
		{{- with .CloudProvider.CSIConfig }}
		  csiConfig: |
		{{ . | indent 4 -}}
		{{ end }}
		containerRuntime:
		  containerd: {}

		versions:
		  kubernetes: {{ .Versions.Kubernetes }}
		{{ with .MachineController }}
		machineController:
		  deploy: false
		{{ end -}}

		{{- with .OperatingSystemManager }}
		operatingSystemManager:
		  deploy: false
		{{ end }}

		{{- with .Addons }}
		addons:
		  enable: true
		  addons:
		{{- range .Addons }}
		    - name: {{ .Name }}
		{{ end }}
		{{- end -}}
	`)

	manifestTemplate = template.Must(
		template.New("manifest").Funcs(sprig.TxtFuncMap()).
			Parse(manifestTemplateSource),
	)
)

type genKubeOneClusterYAMLParams struct {
	providerName      string
	clusterName       string
	kubernetesVersion string
	validProviders    map[string]initProvider
}

func genKubeOneClusterYAML(params *genKubeOneClusterYAMLParams) ([]byte, error) {
	prov := validProviders[params.providerName]

	cluster := kubeonev1beta2.KubeOneCluster{
		TypeMeta: kubeonev1beta2.NewKubeOneCluster().TypeMeta,
		Name:     params.clusterName,
		CloudProvider: kubeonev1beta2.CloudProviderSpec{
			External:    prov.external,
			CloudConfig: prov.cloudConfig,
			CSIConfig:   prov.csiConfig,
		},
		ContainerRuntime: kubeonev1beta2.ContainerRuntimeConfig{
			Containerd: &kubeonev1beta2.ContainerRuntimeContainerd{},
		},
		Versions: kubeonev1beta2.VersionConfig{
			Kubernetes: params.kubernetesVersion,
		},
		Addons: &kubeonev1beta2.Addons{
			Enable: true,
			Addons: []kubeonev1beta2.Addon{
				{
					Name: "default-storage-class",
				},
			},
		},
	}

	providerName := prov.alternativeName
	if providerName == "" {
		providerName = params.providerName
	}

	err := kubeonev1beta2.SetCloudProvider(&cluster.CloudProvider, providerName)
	if err != nil {
		return nil, err
	}

	if cluster.CloudProvider.None != nil {
		cluster.Addons = nil
		cluster.MachineController = &kubeonev1beta2.MachineControllerConfig{
			Deploy: false,
		}
		cluster.OperatingSystemManager = &kubeonev1beta2.OperatingSystemManagerConfig{
			Deploy: false,
		}
	}

	var buf bytes.Buffer
	err = manifestTemplate.Execute(&buf, &cluster)
	if err != nil {
		return nil, fail.Runtime(err, "generating kubeone manifest")
	}

	dummy := kubeonev1beta2.NewKubeOneCluster()
	if err = kyaml.UnmarshalStrict(buf.Bytes(), &dummy); err != nil {
		return nil, fail.Runtime(err, "kubeone manifest testing marshal/unmarshal")
	}

	return buf.Bytes(), err
}
