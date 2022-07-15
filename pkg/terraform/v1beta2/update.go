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

package v1beta2

import (
	"bytes"
	"encoding/json"

	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/templates/machinecontroller"
)

func unmarshalStrict(buf []byte, obj interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(buf))
	dec.DisallowUnknownFields()

	return fail.Runtime(dec.Decode(obj), "strict unmarshal of %T", obj)
}

func updateAWSWorkerset(existingWorkerSet *kubeonev1beta2.DynamicWorkerConfig, cfg json.RawMessage) error {
	var awsCloudConfig machinecontroller.AWSSpec

	if err := unmarshalStrict(cfg, &awsCloudConfig); err != nil {
		return fail.Config(err, "unmarshalling DynamicWorkerConfig AWS spec")
	}

	flags := []cloudProviderFlags{
		{key: "ami", value: awsCloudConfig.AMI},
		{key: "assignPublicIP", value: awsCloudConfig.AssignPublicIP},
		{key: "availabilityZone", value: awsCloudConfig.AvailabilityZone},
		{key: "diskIops", value: awsCloudConfig.DiskIops},
		{key: "diskSize", value: awsCloudConfig.DiskSize},
		{key: "diskType", value: awsCloudConfig.DiskType},
		{key: "ebsVolumeEncrypted", value: awsCloudConfig.EBSVolumeEncrypted},
		{key: "instanceProfile", value: awsCloudConfig.InstanceProfile},
		{key: "instanceType", value: awsCloudConfig.InstanceType},
		{key: "isSpotInstance", value: awsCloudConfig.IsSpotInstance},
		{key: "region", value: awsCloudConfig.Region},
		{key: "securityGroupIDs", value: awsCloudConfig.SecurityGroupIDs},
		{key: "subnetId", value: awsCloudConfig.SubnetID},
		{key: "tags", value: awsCloudConfig.Tags},
		{key: "vpcId", value: awsCloudConfig.VPCID},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return err
		}
	}

	return nil
}

func updateAzureWorkerset(existingWorkerSet *kubeonev1beta2.DynamicWorkerConfig, cfg json.RawMessage) error {
	var azureCloudConfig machinecontroller.AzureSpec

	if err := unmarshalStrict(cfg, &azureCloudConfig); err != nil {
		return fail.Config(err, "unmarshalling DynamicWorkerConfig Azure spec")
	}

	flags := []cloudProviderFlags{
		{key: "location", value: azureCloudConfig.Location},
		{key: "resourceGroup", value: azureCloudConfig.ResourceGroup},
		{key: "vnetResourceGroup", value: azureCloudConfig.VNetResourceGroup},
		{key: "vmSize", value: azureCloudConfig.VMSize},
		{key: "vnetName", value: azureCloudConfig.VNetName},
		{key: "subnetName", value: azureCloudConfig.SubnetName},
		{key: "loadBalancerSku", value: azureCloudConfig.LoadBalancerSku},
		{key: "routeTableName", value: azureCloudConfig.RouteTableName},
		{key: "availabilitySet", value: azureCloudConfig.AvailabilitySet},
		{key: "assignAvailabilitySet", value: azureCloudConfig.AssignAvailabilitySet},
		{key: "securityGroupName", value: azureCloudConfig.SecurityGroupName},
		{key: "zones", value: azureCloudConfig.Zones},
		{key: "imagePlan", value: azureCloudConfig.ImagePlan},
		{key: "imageReference", value: azureCloudConfig.ImageReference},
		{key: "imageID", value: azureCloudConfig.ImageID},
		{key: "osDiskSize", value: azureCloudConfig.OSDiskSize},
		{key: "osDiskSKU", value: azureCloudConfig.OSDiskSKU},
		{key: "dataDiskSize", value: azureCloudConfig.DataDiskSize},
		{key: "dataDiskSKU", value: azureCloudConfig.DataDiskSKU},
		{key: "assignPublicIP", value: azureCloudConfig.AssignPublicIP},
		{key: "tags", value: azureCloudConfig.Tags},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return err
		}
	}

	return nil
}

func updateGCEWorkerset(existingWorkerSet *kubeonev1beta2.DynamicWorkerConfig, cfg json.RawMessage) error {
	var gceCloudConfig machinecontroller.GCESpec

	if err := unmarshalStrict(cfg, &gceCloudConfig); err != nil {
		return fail.Config(err, "unmarshalling DynamicWorkerConfig GCE spec")
	}

	flags := []cloudProviderFlags{
		{key: "diskSize", value: gceCloudConfig.DiskSize},
		{key: "diskType", value: gceCloudConfig.DiskType},
		{key: "machineType", value: gceCloudConfig.MachineType},
		{key: "network", value: gceCloudConfig.Network},
		{key: "subnetwork", value: gceCloudConfig.Subnetwork},
		{key: "zone", value: gceCloudConfig.Zone},
		{key: "preemptible", value: gceCloudConfig.Preemptible},
		{key: "assignPublicIPAddress", value: gceCloudConfig.AssignPublicIPAddress},
		{key: "labels", value: gceCloudConfig.Labels},
		{key: "tags", value: gceCloudConfig.Tags},
		{key: "multizone", value: gceCloudConfig.MultiZone},
		{key: "regional", value: gceCloudConfig.Regional},
		{key: "customImage", value: gceCloudConfig.CustomImage},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return err
		}
	}

	return nil
}

func updateDigitalOceanWorkerset(existingWorkerSet *kubeonev1beta2.DynamicWorkerConfig, cfg json.RawMessage) error {
	var doCloudConfig machinecontroller.DigitalOceanSpec

	if err := unmarshalStrict(cfg, &doCloudConfig); err != nil {
		return fail.Config(err, "unmarshalling DynamicWorkerConfig DigitalOcean spec")
	}

	flags := []cloudProviderFlags{
		{key: "region", value: doCloudConfig.Region},
		{key: "size", value: doCloudConfig.Size},
		{key: "backups", value: doCloudConfig.Backups},
		{key: "ipv6", value: doCloudConfig.IPv6},
		{key: "private_networking", value: doCloudConfig.PrivateNetworking},
		{key: "monitoring", value: doCloudConfig.Monitoring},
		{key: "tags", value: doCloudConfig.Tags},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return err
		}
	}

	return nil
}

func updateHetznerWorkerset(existingWorkerSet *kubeonev1beta2.DynamicWorkerConfig, cfg json.RawMessage) error {
	var hetznerConfig machinecontroller.HetznerSpec

	if err := unmarshalStrict(cfg, &hetznerConfig); err != nil {
		return fail.Config(err, "unmarshalling DynamicWorkerConfig Hetzner spec")
	}

	flags := []cloudProviderFlags{
		{key: "serverType", value: hetznerConfig.ServerType},
		{key: "datacenter", value: hetznerConfig.Datacenter},
		{key: "location", value: hetznerConfig.Location},
		{key: "image", value: hetznerConfig.Image},
		{key: "networks", value: hetznerConfig.Networks},
		{key: "labels", value: hetznerConfig.Labels},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return err
		}
	}

	return nil
}

func updateNutanixWorkerset(existingWorkerSet *kubeonev1beta2.DynamicWorkerConfig, cfg json.RawMessage) error {
	var nutanixConfig machinecontroller.NutanixSpec

	if err := unmarshalStrict(cfg, &nutanixConfig); err != nil {
		return fail.Config(err, "unmarshalling DynamicWorkerConfig Nutanix spec")
	}

	flags := []cloudProviderFlags{
		{key: "clusterName", value: nutanixConfig.ClusterName},
		{key: "projectName", value: nutanixConfig.ProjectName},
		{key: "subnetName", value: nutanixConfig.SubnetName},
		{key: "imageName", value: nutanixConfig.ImageName},
		{key: "cpus", value: nutanixConfig.CPUs},
		{key: "cpuCores", value: nutanixConfig.CPUCores},
		{key: "cpuPassthrough", value: nutanixConfig.CPUPassthrough},
		{key: "memoryMB", value: nutanixConfig.MemoryMB},
		{key: "diskSize", value: nutanixConfig.DiskSize},
		{key: "categories", value: nutanixConfig.Categories},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return err
		}
	}

	return nil
}

func updateOpenStackWorkerset(existingWorkerSet *kubeonev1beta2.DynamicWorkerConfig, cfg json.RawMessage) error {
	var openstackConfig machinecontroller.OpenStackSpec

	if err := unmarshalStrict(cfg, &openstackConfig); err != nil {
		return fail.Config(err, "unmarshalling DynamicWorkerConfig OpenStack spec")
	}

	flags := []cloudProviderFlags{
		{key: "floatingIPPool", value: openstackConfig.FloatingIPPool},
		{key: "image", value: openstackConfig.Image},
		{key: "flavor", value: openstackConfig.Flavor},
		{key: "securityGroups", value: openstackConfig.SecurityGroups},
		{key: "availabilityZone", value: openstackConfig.AvailabilityZone},
		{key: "network", value: openstackConfig.Network},
		{key: "subnet", value: openstackConfig.Subnet},
		{key: "rootDiskSizeGB", value: openstackConfig.RootDiskSizeGB},
		{key: "nodeVolumeAttachLimit", value: openstackConfig.NodeVolumeAttachLimit},
		{key: "tags", value: openstackConfig.Tags},
		{key: "trustDevicePath", value: openstackConfig.TrustDevicePath},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return err
		}
	}

	return nil
}

func updateEquinixMetalWorkerset(existingWorkerSet *kubeonev1beta2.DynamicWorkerConfig, cfg json.RawMessage) error {
	var metalConfig machinecontroller.EquinixMetalSpec

	if err := unmarshalStrict(cfg, &metalConfig); err != nil {
		return fail.Config(err, "unmarshalling DynamicWorkerConfig EquinixMetal spec")
	}

	flags := []cloudProviderFlags{
		{key: "projectID", value: metalConfig.ProjectID},
		{key: "metro", value: metalConfig.Metro},
		{key: "facilities", value: metalConfig.Facilities},
		{key: "instanceType", value: metalConfig.InstanceType},
		{key: "billingCycle", value: metalConfig.BillingCycle},
		{key: "tags", value: metalConfig.Tags},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return err
		}
	}

	return nil
}

func updateVMwareCloudDirectorWorkerset(existingWorkerSet *kubeonev1beta2.DynamicWorkerConfig, cfg json.RawMessage) error {
	var config machinecontroller.VMWareCloudDirectorSpec

	if err := unmarshalStrict(cfg, &config); err != nil {
		return fail.Config(err, "unmarshalling DynamicWorkerConfig VMware Cloud Director spec")
	}

	flags := []cloudProviderFlags{
		{key: "organization", value: config.Organization},
		{key: "vdc", value: config.VDC},
		{key: "vapp", value: config.CPUs},
		{key: "catalog", value: config.Catalog},
		{key: "template", value: config.Template},
		{key: "network", value: config.Network},
		{key: "cpus", value: config.CPUs},
		{key: "cpuCores", value: config.CPUCores},
		{key: "memoryMB", value: config.MemoryMB},
		{key: "diskSizeGB", value: config.DiskSizeGB},
		{key: "storageProfile", value: config.StorageProfile},
		{key: "ipAllocationMode", value: config.IPAllocationMode},
		{key: "metadata", value: config.Metadata},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return err
		}
	}

	return nil
}

func updateVSphereWorkerset(existingWorkerSet *kubeonev1beta2.DynamicWorkerConfig, cfg json.RawMessage) error {
	var vsphereConfig machinecontroller.VSphereSpec

	if err := unmarshalStrict(cfg, &vsphereConfig); err != nil {
		return fail.Config(err, "unmarshalling DynamicWorkerConfig vSphere spec")
	}

	flags := []cloudProviderFlags{
		{key: "allowInsecure", value: vsphereConfig.AllowInsecure},
		{key: "cluster", value: vsphereConfig.Cluster},
		{key: "cpus", value: vsphereConfig.CPUs},
		{key: "datacenter", value: vsphereConfig.Datacenter},
		{key: "datastore", value: vsphereConfig.Datastore},
		{key: "datastoreCluster", value: vsphereConfig.DatastoreCluster},
		{key: "diskSizeGB", value: vsphereConfig.DiskSizeGB},
		{key: "folder", value: vsphereConfig.Folder},
		{key: "resourcePool", value: vsphereConfig.ResourcePool},
		{key: "memoryMB", value: vsphereConfig.MemoryMB},
		{key: "templateVMName", value: vsphereConfig.TemplateVMName},
		{key: "vmNetName", value: vsphereConfig.VMNetName},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return err
		}
	}

	return nil
}
