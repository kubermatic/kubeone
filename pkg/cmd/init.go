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

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	yamlv2 "gopkg.in/yaml.v2"

	"k8c.io/kubeone/examples"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/yamled"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

type initProvider struct {
	terraformPath string
	inTree        bool
	cloudConfig   string
	csiConfig     string
}

var (
	validProviders = map[string]initProvider{
		"aws": {
			terraformPath: "terraform/aws",
		},
		"azure": {
			terraformPath: "terraform/azure",
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
		},
		"equinixmetal": {
			terraformPath: "terraform/equinixmetal",
		},
		"gce": {
			terraformPath: "terraform/gce",
		},
		"hetzner": {
			terraformPath: "terraform/hetzner",
		},
		"none": {
			terraformPath: "",
			inTree:        true,
		},
		"openstack": {
			terraformPath: "terraform/openstack",
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
			terraformPath: "terraform/vmware-cloud-director",
		},
		"vsphere": {
			terraformPath: "terraform/vsphere",
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
			terraformPath: "terraform/vsphere_flatcar",
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
	Provider  oneOfFlag `longflag:"provider"`
	Name      string    `longflag:"name"`
	Version   string    `longflag:"version"`
	Terraform bool      `longflag:"terraform"`
	Path      string    `longflag:"path"`
}

func initCmd() *cobra.Command {
	opts := &initOpts{
		Provider: oneOfFlag{
			validSet:     sets.StringKeySet(validProviders),
			defaultValue: "none",
		},
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "init new kubeone cluster configuration",
		Long: heredoc.Doc(`
			Initialize new KubeOne + terraform configuration for chosen provider.
		`),
		SilenceErrors: true,
		Example:       `kubeone init --provider aws`,
		RunE: func(_ *cobra.Command, args []string) error {
			return runInit(opts)
		},
	}

	providerUsageText := fmt.Sprintf("provider to initialize, possible values: %s", strings.Join(opts.Provider.PossibleValues(), ", "))

	cmd.Flags().BoolVar(&opts.Terraform, longFlagName(opts, "Terraform"), false, "generate terraform config")
	cmd.Flags().StringVar(&opts.Name, longFlagName(opts, "Name"), "example", "name of the cluster")
	cmd.Flags().StringVar(&opts.Version, longFlagName(opts, "version"), "v1.24.5", "kubernetes version")
	cmd.Flags().StringVar(&opts.Path, longFlagName(opts, "Path"), ".", "path where to write files")
	cmd.Flags().Var(&opts.Provider, longFlagName(opts, "Provider"), providerUsageText)

	return cmd
}

func runInit(opts *initOpts) error {
	err := os.MkdirAll(opts.Path, 0750)
	if err != nil {
		return err
	}

	k1config, err := os.OpenFile(filepath.Join(opts.Path, "kubeone.yaml"), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fail.Runtime(err, "creating manifest file")
	}
	defer k1config.Close()

	ybuf, err := genKubeOneClusterYAML(opts)
	if err != nil {
		return fail.Runtime(err, "generating KubeOneCluster")
	}

	_, err = io.Copy(k1config, bytes.NewBuffer(ybuf))
	if err != nil {
		return fail.Runtime(err, "writing KubeOneCluster")
	}

	if opts.Terraform {
		prov := validProviders[opts.Provider.String()]
		if err = examples.CopyTo(opts.Path, prov.terraformPath); err != nil {
			return fail.Runtime(err, "copying terraform configuration")
		}
	}

	return nil
}

func genKubeOneClusterYAML(opts *initOpts) ([]byte, error) {
	providerName := opts.Provider.String()
	prov := validProviders[providerName]

	cluster := kubeonev1beta2.KubeOneCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeOneCluster",
			APIVersion: kubeonev1beta2.SchemeGroupVersion.Identifier(),
		},
		Name: opts.Name,
		CloudProvider: kubeonev1beta2.CloudProviderSpec{
			External:    !prov.inTree,
			CloudConfig: prov.cloudConfig,
			CSIConfig:   prov.csiConfig,
		},
		ContainerRuntime: kubeonev1beta2.ContainerRuntimeConfig{
			Containerd: &kubeonev1beta2.ContainerRuntimeContainerd{},
		},
		Versions: kubeonev1beta2.VersionConfig{
			Kubernetes: opts.Version,
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

	switch strings.Split(providerName, "/")[0] {
	case "aws":
		cluster.CloudProvider.AWS = &kubeonev1beta2.AWSSpec{}
	case "azure":
		cluster.CloudProvider.Azure = &kubeonev1beta2.AzureSpec{}
	case "digitalocean":
		cluster.CloudProvider.DigitalOcean = &kubeonev1beta2.DigitalOceanSpec{}
	case "equinixmetal":
		cluster.CloudProvider.DigitalOcean = &kubeonev1beta2.DigitalOceanSpec{}
	case "gce":
		cluster.CloudProvider.GCE = &kubeonev1beta2.GCESpec{}
	case "hetzner":
		cluster.CloudProvider.Hetzner = &kubeonev1beta2.HetznerSpec{}
	case "none":
		cluster.CloudProvider.None = &kubeonev1beta2.NoneSpec{}
	case "openstack":
		cluster.CloudProvider.Openstack = &kubeonev1beta2.OpenstackSpec{}
	case "vmware-cloud-director":
		cluster.CloudProvider.VMwareCloudDirector = &kubeonev1beta2.VMwareCloudDirectorSpec{}
	case "vsphere":
		cluster.CloudProvider.Vsphere = &kubeonev1beta2.VsphereSpec{}
	default:
		return nil, fmt.Errorf("unknown provider")
	}

	buf, err := yaml.Marshal(&cluster)
	if err != nil {
		return nil, err
	}

	doc, err := yamled.Load(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	toremove := []string{
		"apiEndpoint",
		"clusterNetwork",
		"controlPlane",
		"hosts",
		"features",
		"loggingConfig",
		"proxy",
		"staticWorkers",
	}
	for _, yamlPath := range toremove {
		doc.Remove(yamled.Path{yamlPath})
	}

	return yamlv2.Marshal(doc)
}
