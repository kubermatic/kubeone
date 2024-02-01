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

package initcmd

import (
	"github.com/MakeNowJust/heredoc/v2"
)

type initProvider struct {
	title           string
	alternativeName string
	terraformPath   string
	external        bool
	cloudConfig     string
	csiConfig       string
	requiredTFVars  []terraformVariable
	optionalTFVars  []terraformVariable
	workerPerAZ     bool
}

type terraformVariable struct {
	Name         string
	Description  string
	DefaultValue string
	Choices      []terraformVariableChoice
}

type terraformVariableChoice struct {
	Name  string
	Value string
}

var (
	osUbuntu = terraformVariableChoice{
		Name:  "Ubuntu",
		Value: "ubuntu",
	}
	osCentos = terraformVariableChoice{
		Name:  "CentOS",
		Value: "centos",
	}
	osRockyLinux = terraformVariableChoice{
		Name:  "Rocky Linux",
		Value: "rockylinux",
	}
	osOracleLinux = terraformVariableChoice{
		Name:  "Oracle Linux",
		Value: "rockylinux",
	}
	osRHEL = terraformVariableChoice{
		Name:  "Red Hat Enterprise Linux (RHEL)",
		Value: "rhel",
	}
	osFlatcar = terraformVariableChoice{
		Name:  "Flatcar",
		Value: "flatcar",
	}
	osAmazonLinux2 = terraformVariableChoice{
		Name:  "Amazon Linux 2",
		Value: "amzn2",
	}
)

var (
	ValidProviders = map[string]initProvider{
		"aws": {
			title:         "AWS",
			terraformPath: "terraform/aws",
			external:      true,
			optionalTFVars: []terraformVariable{
				{
					Name:         "os",
					Description:  "Operating system to use for this cluster",
					DefaultValue: osUbuntu.Name,
					Choices:      []terraformVariableChoice{osUbuntu, osCentos, osRockyLinux, osOracleLinux, osRHEL, osFlatcar, osAmazonLinux2},
				},
			},
			workerPerAZ: true,
		},
		"azure": {
			title:         "Azure",
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
			optionalTFVars: []terraformVariable{
				{
					Name:         "os",
					Description:  "Operating system to use for this cluster",
					DefaultValue: osUbuntu.Name,
					Choices:      []terraformVariableChoice{osUbuntu, osCentos, osRockyLinux, osOracleLinux, osRHEL, osFlatcar},
				},
			},
		},
		"digitalocean": {
			title:         "DigitalOcean",
			terraformPath: "terraform/digitalocean",
			external:      true,
			optionalTFVars: []terraformVariable{
				{
					Name:         "os",
					Description:  "Operating system to use for this cluster",
					DefaultValue: osUbuntu.Name,
					Choices:      []terraformVariableChoice{osUbuntu, osCentos, osRockyLinux, osOracleLinux},
				},
			},
		},
		"equinixmetal": {
			title:         "Equinix Metal",
			terraformPath: "terraform/equinixmetal",
			external:      true,
			requiredTFVars: []terraformVariable{
				{
					Name:        "project_id",
					Description: "ID of your Equinix Metal project",
				},
			},
			optionalTFVars: []terraformVariable{
				{
					Name:         "os",
					Description:  "Operating system to use for this cluster",
					DefaultValue: osUbuntu.Name,
					Choices:      []terraformVariableChoice{osUbuntu, osCentos, osRockyLinux, osOracleLinux, osFlatcar},
				},
			},
		},
		"gce": {
			title:         "GCE",
			terraformPath: "terraform/gce",
			requiredTFVars: []terraformVariable{
				{
					Name:        "project",
					Description: "Name of your GCE project",
				},
			},
		},
		"hetzner": {
			title:         "Hetzner",
			terraformPath: "terraform/hetzner",
			external:      true,
			optionalTFVars: []terraformVariable{
				{
					Name:         "os",
					Description:  "Operating system to use for this cluster",
					DefaultValue: osUbuntu.Name,
					Choices:      []terraformVariableChoice{osUbuntu, osCentos, osRockyLinux, osOracleLinux},
				},
			},
		},
		"none": {
			title:         "None (e.g. baremetal)",
			terraformPath: "",
		},
		"nutanix": {
			title:         "Nutanix",
			terraformPath: "terraform/nutanix",
			requiredTFVars: []terraformVariable{
				{
					Name:        "nutanix_cluster_name",
					Description: "Name of Nutanix Cluster object",
				},
				{
					Name:        "project_name",
					Description: "Name of your Nutanix project",
				},
				{
					Name:        "subnet_name",
					Description: "Name of subnet to be used",
				},
				{
					Name:        "image_name",
					Description: "Name of image to be used for nodes",
				},
			},
			optionalTFVars: []terraformVariable{
				{
					Name:         "worker_os",
					Description:  "Operating system of the provided image",
					DefaultValue: osUbuntu.Name,
					Choices:      []terraformVariableChoice{osUbuntu, osCentos, osRockyLinux, osOracleLinux, osRHEL, osFlatcar, osAmazonLinux2},
				},
			},
		},
		"openstack": {
			title:         "OpenStack",
			terraformPath: "terraform/openstack",
			external:      true,
			requiredTFVars: []terraformVariable{
				{
					Name:        "external_network_name",
					Description: "Name of the external network object to be used",
				},
				{
					Name:        "image",
					Description: "Name of image to be used for nodes",
				},
				{
					Name:        "subnet_cidr",
					Description: "Subnet CIDR to be used for this cluster",
				},
			},
			optionalTFVars: []terraformVariable{
				{
					Name:         "worker_os",
					Description:  "Operating system of the provided image",
					DefaultValue: osUbuntu.Name,
					Choices:      []terraformVariableChoice{osUbuntu, osCentos, osRockyLinux, osOracleLinux, osRHEL, osFlatcar, osAmazonLinux2},
				},
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
			title:           "VMware Cloud Director",
			alternativeName: "vmwareCloudDirector",
			terraformPath:   "terraform/vmware-cloud-director",
			requiredTFVars: []terraformVariable{
				{
					Name:        "vcd_vdc_name",
					Description: "TODO",
				},
				{
					Name:        "vcd_edge_gateway_name",
					Description: "Name of the edge gateway to be used",
				},
				{
					Name:        "catalog_name",
					Description: "Name of the catalog to be used",
				},
				{
					Name:        "template_name",
					Description: "Name of the template to be used",
				},
			},
			optionalTFVars: []terraformVariable{
				{
					Name:         "worker_os",
					Description:  "Operating system of the provided image",
					DefaultValue: "Ubuntu",
					Choices:      []terraformVariableChoice{osUbuntu, osCentos, osRockyLinux, osOracleLinux, osRHEL, osFlatcar, osAmazonLinux2},
				},
			},
		},
		"vsphere": {
			title:         "vSphere (Ubuntu)",
			terraformPath: "terraform/vsphere",
			external:      true,
			requiredTFVars: []terraformVariable{
				{
					Name:        "datastore_name",
					Description: "Datastore name",
				},
				{
					Name:        "network_name",
					Description: "Network name",
				},
				{
					Name:        "template_name",
					Description: "Template name",
				},
				{
					Name:        "resource_pool_name",
					Description: "Resource pool name",
				},
			},
			optionalTFVars: []terraformVariable{
				{
					Name:         "worker_os",
					Description:  "Operating system of the provided image",
					DefaultValue: "Ubuntu",
					Choices:      []terraformVariableChoice{osUbuntu, osCentos, osRockyLinux, osOracleLinux, osRHEL, osFlatcar, osAmazonLinux2},
				},
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
			title:           "vSphere (Flatcar)",
			alternativeName: "vsphere",
			terraformPath:   "terraform/vsphere_flatcar",
			external:        true,
			requiredTFVars: []terraformVariable{
				{
					Name:        "datastore_name",
					Description: "Datastore name",
				},
				{
					Name:        "network_name",
					Description: "Network name",
				},
				{
					Name:        "template_name",
					Description: "Template name",
				},
				{
					Name:        "resource_pool_name",
					Description: "Resource pool name",
				},
			},
			optionalTFVars: []terraformVariable{
				{
					Name:         "worker_os",
					Description:  "Operating system of the provided image",
					DefaultValue: "Ubuntu",
					Choices:      []terraformVariableChoice{osUbuntu, osCentos, osRockyLinux, osOracleLinux, osRHEL, osFlatcar, osAmazonLinux2},
				},
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
	}
)
