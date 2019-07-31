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

package terraform

import (
	"encoding/json"

	"github.com/pkg/errors"

	kubeonev1alpha1 "github.com/kubermatic/kubeone/pkg/apis/kubeone/v1alpha1"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
)

type controlPlane struct {
	ClusterName       string   `json:"cluster_name"`
	CloudProvider     *string  `json:"cloud_provider"`
	PublicAddress     []string `json:"public_address"`
	PrivateAddress    []string `json:"private_address"`
	Hostnames         []string `json:"hostnames"`
	SSHUser           string   `json:"ssh_user"`
	SSHPort           int      `json:"ssh_port"`
	SSHPrivateKeyFile string   `json:"ssh_private_key_file"`
	SSHAgentSocket    string   `json:"ssh_agent_socket"`
	Bastion           string   `json:"bastion"`
	BastionPort       int      `json:"bastion_port"`
}

// Config represents configuration in the terraform output format
type Config struct {
	KubeOneAPI struct {
		Value struct {
			Endpoint string `json:"endpoint"`
		} `json:"value"`
	} `json:"kubeone_api"`

	KubeOneHosts struct {
		Value struct {
			ControlPlane controlPlane `json:"control_plane"`
		} `json:"value"`
	} `json:"kubeone_hosts"`

	KubeOneWorkers struct {
		Value map[string]kubeonev1alpha1.WorkerConfig `json:"value"`
	} `json:"kubeone_workers"`
}

type cloudProviderFlags struct {
	key   string
	value interface{}
}

// NewConfigFromJSON creates a new config object from json
func NewConfigFromJSON(j []byte) (c *Config, err error) {
	c = &Config{}
	return c, json.Unmarshal(j, c)
}

// Apply adds the terraform configuration options to the given
// cluster config.
func (c *Config) Apply(cluster *kubeonev1alpha1.KubeOneCluster) error {
	if c.KubeOneAPI.Value.Endpoint != "" {
		cluster.APIEndpoint = kubeonev1alpha1.APIEndpoint{
			Host: c.KubeOneAPI.Value.Endpoint,
		}
	}

	cp := c.KubeOneHosts.Value.ControlPlane

	if cp.CloudProvider != nil {
		cluster.CloudProvider.Name = kubeonev1alpha1.CloudProviderName(*cp.CloudProvider)
	}

	var err error

	cluster.Name = cp.ClusterName

	// build up a list of master nodes
	hosts := make([]kubeonev1alpha1.HostConfig, 0)
	for i, publicIP := range cp.PublicAddress {
		privateIP := publicIP
		if i < len(cp.PrivateAddress) {
			privateIP = cp.PrivateAddress[i]
		}

		hostname := ""
		if i < len(cp.Hostnames) {
			hostname = cp.Hostnames[i]
		}

		hosts = append(hosts, newHostConfig(i, publicIP, privateIP, hostname, cp))
	}

	if len(hosts) == 0 {
		// there was no public IPs available
		for i, privateIP := range cp.PrivateAddress {
			hostname := ""
			if i < len(cp.Hostnames) {
				hostname = cp.Hostnames[i]
			}

			hosts = append(hosts, newHostConfig(i, "", privateIP, hostname, cp))
		}
	}

	if len(hosts) > 0 {
		cluster.Hosts = hosts
	}

	// Walk through all configued workersets from terraform and apply their config
	// by either merging it into an existing workerSet or creating a new one
	for workersetName, workersetValue := range c.KubeOneWorkers.Value {
		var existingWorkerSet *kubeonev1alpha1.WorkerConfig

		// Check do we have a workerset with the same name defined
		// in the KubeOneCluster object
		for idx, workerset := range cluster.Workers {
			if workerset.Name == workersetName {
				existingWorkerSet = &cluster.Workers[idx]
				break
			}
		}

		// If we didn't found a workerset defined in the cluster object,
		// append a workerset from the terraform output to the cluster object
		if existingWorkerSet == nil {
			// no existing workerset found, use what we have from terraform
			workersetValue.Name = workersetName
			cluster.Workers = append(cluster.Workers, workersetValue)
			continue
		}

		// If we found a workerset defined in the cluster object,
		// merge values from the object and the terraform output
		switch cluster.CloudProvider.Name {
		case kubeonev1alpha1.CloudProviderNameAWS:
			err = c.updateAWSWorkerset(existingWorkerSet, workersetValue.Config.CloudProviderSpec)
		case kubeonev1alpha1.CloudProviderNameAzure:
			err = c.updateAzureWorkerset(existingWorkerSet, workersetValue.Config.CloudProviderSpec)
		case kubeonev1alpha1.CloudProviderNameGCE:
			err = c.updateGCEWorkerset(existingWorkerSet, workersetValue.Config.CloudProviderSpec)
		case kubeonev1alpha1.CloudProviderNameDigitalOcean:
			err = c.updateDigitalOceanWorkerset(existingWorkerSet, workersetValue.Config.CloudProviderSpec)
		case kubeonev1alpha1.CloudProviderNameHetzner:
			err = c.updateHetznerWorkerset(existingWorkerSet, workersetValue.Config.CloudProviderSpec)
		case kubeonev1alpha1.CloudProviderNameOpenStack:
			err = c.updateOpenStackWorkerset(existingWorkerSet, workersetValue.Config.CloudProviderSpec)
		case kubeonev1alpha1.CloudProviderNameVSphere:
			err = c.updateVSphereWorkerset(existingWorkerSet, workersetValue.Config.CloudProviderSpec)
		case kubeonev1alpha1.CloudProviderNamePacket:
			err = c.updatePacketWorkerset(existingWorkerSet, workersetValue.Config.CloudProviderSpec)
		default:
			return errors.Errorf("unknown provider %v", cluster.CloudProvider.Name)
		}

		if err != nil {
			return errors.Wrapf(err, "failed to update provider-specific config for workerset %q from terraform config", workersetName)
		}
	}

	return nil
}

func newHostConfig(id int, publicIP, privateIP, hostname string, cp controlPlane) kubeonev1alpha1.HostConfig {
	return kubeonev1alpha1.HostConfig{
		ID:                id,
		PublicAddress:     publicIP,
		PrivateAddress:    privateIP,
		Hostname:          hostname,
		SSHUsername:       cp.SSHUser,
		SSHPort:           cp.SSHPort,
		SSHPrivateKeyFile: cp.SSHPrivateKeyFile,
		SSHAgentSocket:    cp.SSHAgentSocket,
		Bastion:           cp.Bastion,
		BastionPort:       cp.BastionPort,
	}
}

func (c *Config) updateAWSWorkerset(existingWorkerSet *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
	var awsCloudConfig machinecontroller.AWSSpec

	if err := json.Unmarshal(cfg, &awsCloudConfig); err != nil {
		return errors.WithStack(err)
	}

	flags := []cloudProviderFlags{
		{key: "ami", value: awsCloudConfig.AMI},
		{key: "assignPublicIP", value: awsCloudConfig.AssignPublicIP},
		{key: "availabilityZone", value: awsCloudConfig.AvailabilityZone},
		{key: "diskIops", value: awsCloudConfig.DiskIops},
		{key: "diskSize", value: awsCloudConfig.DiskSize},
		{key: "diskType", value: awsCloudConfig.DiskType},
		{key: "instanceProfile", value: awsCloudConfig.InstanceProfile},
		{key: "instanceType", value: awsCloudConfig.InstanceType},
		{key: "region", value: awsCloudConfig.Region},
		{key: "securityGroupIDs", value: awsCloudConfig.SecurityGroupIDs},
		{key: "subnetId", value: awsCloudConfig.SubnetID},
		{key: "tags", value: awsCloudConfig.Tags},
		{key: "vpcId", value: awsCloudConfig.VPCID},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateAzureWorkerset(existingWorkerSet *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
	var azureCloudConfig machinecontroller.AzureSpec

	if err := json.Unmarshal(cfg, &azureCloudConfig); err != nil {
		return errors.WithStack(err)
	}

	flags := []cloudProviderFlags{
		{key: "assignPublicIP", value: azureCloudConfig.AssignPublicIP},
		{key: "availabilitySet", value: azureCloudConfig.AvailabilitySet},
		{key: "location", value: azureCloudConfig.Location},
		{key: "resourceGroup", value: azureCloudConfig.ResourceGroup},
		{key: "routeTableName", value: azureCloudConfig.RouteTableName},
		{key: "securityGroupName", value: azureCloudConfig.SecurityGroupName},
		{key: "subnetName", value: azureCloudConfig.SubnetName},
		{key: "tags", value: azureCloudConfig.Tags},
		{key: "vmSize", value: azureCloudConfig.VMSize},
		{key: "vnetName", value: azureCloudConfig.VNetName},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateGCEWorkerset(existingWorkerSet *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
	var gceCloudConfig machinecontroller.GCESpec

	if err := json.Unmarshal(cfg, &gceCloudConfig); err != nil {
		return errors.WithStack(err)
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
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateDigitalOceanWorkerset(existingWorkerSet *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
	var doCloudConfig machinecontroller.DigitalOceanSpec

	if err := json.Unmarshal(cfg, &doCloudConfig); err != nil {
		return errors.WithStack(err)
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
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateHetznerWorkerset(existingWorkerSet *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
	var hetznerConfig machinecontroller.HetznerSpec

	if err := json.Unmarshal(cfg, &hetznerConfig); err != nil {
		return err
	}

	flags := []cloudProviderFlags{
		{key: "serverType", value: hetznerConfig.ServerType},
		{key: "datacenter", value: hetznerConfig.Datacenter},
		{key: "location", value: hetznerConfig.Location},
		{key: "labels", value: hetznerConfig.Labels},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateOpenStackWorkerset(existingWorkerSet *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
	var openstackConfig machinecontroller.OpenStackSpec

	if err := json.Unmarshal(cfg, &openstackConfig); err != nil {
		return err
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
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updatePacketWorkerset(existingWorkerSet *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
	var packetConfig machinecontroller.PacketSpec

	if err := json.Unmarshal(cfg, &packetConfig); err != nil {
		return err
	}

	flags := []cloudProviderFlags{
		{key: "projectID", value: packetConfig.ProjectID},
		{key: "facilities", value: packetConfig.Facilities},
		{key: "instanceType", value: packetConfig.InstanceType},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateVSphereWorkerset(existingWorkerSet *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
	var vsphereConfig machinecontroller.VSphereSpec

	if err := json.Unmarshal(cfg, &vsphereConfig); err != nil {
		return err
	}

	flags := []cloudProviderFlags{
		{key: "allowInsecure", value: vsphereConfig.AllowInsecure},
		{key: "cluster", value: vsphereConfig.Cluster},
		{key: "cpus", value: vsphereConfig.CPUs},
		{key: "datacenter", value: vsphereConfig.Datacenter},
		{key: "datastore", value: vsphereConfig.Datastore},
		{key: "diskSizeGB", value: vsphereConfig.DiskSizeGB},
		{key: "folder", value: vsphereConfig.Folder},
		{key: "memoryMB", value: vsphereConfig.MemoryMB},
		{key: "templateNetName", value: vsphereConfig.TemplateNetName},
		{key: "templateVMName", value: vsphereConfig.TemplateVMName},
		{key: "vmNetName", value: vsphereConfig.VMNetName},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(existingWorkerSet, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func setWorkersetFlag(w *kubeonev1alpha1.WorkerConfig, name string, value interface{}) error {
	// ignore empty values (i.e. not set in terraform output)
	switch s := value.(type) {
	case int:
		if s == 0 {
			return nil
		}
	case *int:
		if s == nil {
			return nil
		}
	case *uint:
		if s == nil {
			return nil
		}
	case string:
		if s == "" {
			return nil
		}
	case *string:
		if s == nil {
			return nil
		}
	case []string:
		if len(s) == 0 {
			return nil
		}
	case map[string]string:
		if s == nil {
			return nil
		}
	case bool:
	case *bool:
		if s == nil {
			return nil
		}
	default:
		return errors.New("unsupported type")
	}

	// update CloudProviderSpec ONLY IF given terraform output is absent in
	// original CloudProviderSpec
	jsonSpec := make(map[string]interface{})
	if w.Config.CloudProviderSpec != nil {
		if err := json.Unmarshal(w.Config.CloudProviderSpec, &jsonSpec); err != nil {
			return errors.Wrap(err, "unable to parse the provided cloud provider")
		}
	}

	if _, exists := jsonSpec[name]; !exists {
		jsonSpec[name] = value
	}

	var err error
	w.Config.CloudProviderSpec, err = json.Marshal(jsonSpec)
	if err != nil {
		return errors.Wrap(err, "unable to update the cloud provider spec")
	}

	return nil
}
