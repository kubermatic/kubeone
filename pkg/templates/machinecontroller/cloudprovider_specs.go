/*
Copyright 2019 The KubeOne Authors.

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

package machinecontroller

// AWSSpec holds cloudprovider spec for AWS
type AWSSpec struct {
	AMI                string            `json:"ami"`
	AssignPublicIP     *bool             `json:"assignPublicIP"`
	AvailabilityZone   string            `json:"availabilityZone"`
	DiskIops           *int              `json:"diskIops,omitempty"`
	DiskSize           *int              `json:"diskSize"`
	DiskType           string            `json:"diskType"`
	EBSVolumeEncrypted bool              `json:"ebsVolumeEncrypted"`
	InstanceProfile    string            `json:"instanceProfile"`
	InstanceType       *string           `json:"instanceType"`
	IsSpotInstance     *bool             `json:"isSpotInstance,omitempty"`
	Region             string            `json:"region"`
	SecurityGroupIDs   []string          `json:"securityGroupIDs"`
	SubnetID           string            `json:"subnetId"`
	Tags               map[string]string `json:"tags"`
	VPCID              string            `json:"vpcId"`
}

// DigitalOceanSpec holds cloudprovider spec for DigitalOcean
type DigitalOceanSpec struct {
	Region            string   `json:"region"`
	Size              string   `json:"size"`
	Backups           bool     `json:"backups"`
	IPv6              bool     `json:"ipv6"`
	PrivateNetworking bool     `json:"private_networking"`
	Monitoring        bool     `json:"monitoring"`
	Tags              []string `json:"tags"`
}

// OpenStackSpec holds cloudprovider spec for OpenStack
type OpenStackSpec struct {
	Image                 string            `json:"image"`
	Flavor                string            `json:"flavor"`
	SecurityGroups        []string          `json:"securityGroups"`
	FloatingIPPool        string            `json:"floatingIPPool"`
	AvailabilityZone      string            `json:"availabilityZone"`
	Network               string            `json:"network"`
	Subnet                string            `json:"subnet"`
	RootDiskSizeGB        *int              `json:"rootDiskSizeGB,omitempty"`
	NodeVolumeAttachLimit *uint             `json:"nodeVolumeAttachLimit,omitempty"`
	TrustDevicePath       bool              `json:"trustDevicePath"`
	Tags                  map[string]string `json:"tags"`
}

// GCESpec holds cloudprovider spec for GCE
type GCESpec struct {
	DiskSize              int               `json:"diskSize"`
	DiskType              string            `json:"diskType"`
	MachineType           string            `json:"machineType"`
	Network               string            `json:"network"`
	Subnetwork            string            `json:"subnetwork"`
	Zone                  string            `json:"zone"`
	Preemptible           bool              `json:"preemptible"`
	AssignPublicIPAddress *bool             `json:"assignPublicIPAddress"`
	Labels                map[string]string `json:"labels"`
	Tags                  []string          `json:"tags"`
	MultiZone             *bool             `json:"multizone"`
	Regional              *bool             `json:"regional"`
	CustomImage           string            `json:"customImage,omitempty"`
}

// HetznerSpec holds cloudprovider spec for Hetzner
type HetznerSpec struct {
	ServerType string            `json:"serverType"`
	Datacenter string            `json:"datacenter"`
	Location   string            `json:"location"`
	Image      string            `json:"image"`
	Networks   []string          `json:"networks"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// PacketSpec holds cloudprovider spec for Packet
type PacketSpec struct {
	ProjectID    string   `json:"projectID"`
	BillingCycle string   `json:"billingCycle"`
	Facilities   []string `json:"facilities"`
	InstanceType string   `json:"instanceType"`
	Tags         []string `json:"tags,omitempty"`
}

// VSphereSpec holds cloudprovider spec for vSphere
type VSphereSpec struct {
	AllowInsecure    bool   `json:"allowInsecure"`
	Cluster          string `json:"cluster"`
	CPUs             int    `json:"cpus"`
	Datacenter       string `json:"datacenter"`
	Datastore        string `json:"datastore"`
	DatastoreCluster string `json:"datastoreCluster"`
	DiskSizeGB       *int   `json:"diskSizeGB,omitempty"`
	Folder           string `json:"folder"`
	ResourcePool     string `json:"resourcePool"`
	MemoryMB         int    `json:"memoryMB"`
	TemplateVMName   string `json:"templateVMName"`
	VMNetName        string `json:"vmNetName,omitempty"`
}

// AzureSpec holds cloudprovider spec for Azure
type AzureSpec struct {
	AssignPublicIP    bool              `json:"assignPublicIP"`
	AvailabilitySet   string            `json:"availabilitySet"`
	Location          string            `json:"location"`
	ResourceGroup     string            `json:"resourceGroup"`
	RouteTableName    string            `json:"routeTableName"`
	SecurityGroupName string            `json:"securityGroupName"`
	Zones             []string          `json:"zones"`
	ImagePlan         *AzureImagePlan   `json:"imagePlan"`
	SubnetName        string            `json:"subnetName"`
	Tags              map[string]string `json:"tags"`
	VMSize            string            `json:"vmSize"`
	VNetName          string            `json:"vnetName"`
	ImageID           string            `json:"imageID"`
	OSDiskSize        int               `json:"osDiskSize"`
	DataDiskSize      int               `json:"dataDiskSize"`
}

type AzureImagePlan struct {
	Name      string `json:"name,omitempty"`
	Publisher string `json:"publisher,omitempty"`
	Product   string `json:"product,omitempty"`
}
