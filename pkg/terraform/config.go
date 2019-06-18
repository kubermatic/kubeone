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
	SSHUser           string   `json:"ssh_user"`
	SSHPort           int      `json:"ssh_port"`
	SSHPrivateKeyFile string   `json:"ssh_private_key_file"`
	SSHAgentSocket    string   `json:"ssh_agent_socket"`
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
		Value map[string]json.RawMessage `json:"value"`
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

		hosts = append(hosts, kubeonev1alpha1.HostConfig{
			ID:                i,
			PublicAddress:     publicIP,
			PrivateAddress:    privateIP,
			SSHUsername:       cp.SSHUser,
			SSHPort:           cp.SSHPort,
			SSHPrivateKeyFile: cp.SSHPrivateKeyFile,
			SSHAgentSocket:    cp.SSHAgentSocket,
		})
	}

	if len(hosts) > 0 {
		cluster.Hosts = hosts
	}

	// Walk through all configued workersets from terraform and apply their config
	// by either merging it into an existing workerSet or creating a new one
	for workersetName, workersetValue := range c.KubeOneWorkers.Value {
		var existingWorkerSet *kubeonev1alpha1.WorkerConfig
		for idx, workerset := range cluster.Workers {
			if workerset.Name == workersetName {
				existingWorkerSet = &cluster.Workers[idx]
				break
			}
		}
		if existingWorkerSet == nil {
			// Append copies the object when its a literal and not a pointer, hence
			// we have to first append, then create a pointer to the appended object
			cluster.Workers = append(cluster.Workers, kubeonev1alpha1.WorkerConfig{Name: workersetName})
			existingWorkerSet = &cluster.Workers[len(cluster.Workers)-1]
		}

		switch cluster.CloudProvider.Name {
		case kubeonev1alpha1.CloudProviderNameAWS:
			err = c.updateAWSWorkerset(existingWorkerSet, workersetValue)
		case kubeonev1alpha1.CloudProviderNameAzure:
			err = c.updateAzureWorkerset(existingWorkerSet, workersetValue)
		case kubeonev1alpha1.CloudProviderNameGCE:
			err = c.updateGCEWorkerset(existingWorkerSet, workersetValue)
		case kubeonev1alpha1.CloudProviderNameDigitalOcean:
			err = c.updateDigitalOceanWorkerset(existingWorkerSet, workersetValue)
		case kubeonev1alpha1.CloudProviderNameHetzner:
			err = c.updateHetznerWorkerset(existingWorkerSet, workersetValue)
		case kubeonev1alpha1.CloudProviderNameOpenStack:
			err = c.updateOpenStackWorkerset(existingWorkerSet, workersetValue)
		case kubeonev1alpha1.CloudProviderNameVSphere:
			err = c.updateVSphereWorkerset(existingWorkerSet, workersetValue)
		case kubeonev1alpha1.CloudProviderNamePacket:
			err = c.updatePacketWorkerset(existingWorkerSet, workersetValue)
		default:
			return errors.Errorf("unknown provider %v", cluster.CloudProvider.Name)
		}

		if err != nil {
			return errors.Wrapf(err, "failed to update provider-specific config for workerset %q from terraform config", workersetName)
		}

		// copy over common config
		if err = c.updateCommonWorkerConfig(existingWorkerSet, workersetValue); err != nil {
			return errors.Wrap(err, "failed to update common config from terraform config")
		}
	}

	return nil
}

func (c *Config) updateAWSWorkerset(workerset *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
	var awsCloudConfig machinecontroller.AWSSpec

	if err := json.Unmarshal(cfg, &awsCloudConfig); err != nil {
		return errors.WithStack(err)
	}

	flags := []cloudProviderFlags{
		{key: "ami", value: awsCloudConfig.AMI},
		{key: "availabilityZone", value: awsCloudConfig.AvailabilityZone},
		{key: "instanceProfile", value: awsCloudConfig.InstanceProfile},
		{key: "region", value: awsCloudConfig.Region},
		{key: "securityGroupIDs", value: awsCloudConfig.SecurityGroupIDs},
		{key: "subnetId", value: awsCloudConfig.SubnetID},
		{key: "vpcId", value: awsCloudConfig.VPCID},
		{key: "instanceType", value: awsCloudConfig.InstanceType},
		{key: "tags", value: awsCloudConfig.Tags},
		{key: "diskType", value: awsCloudConfig.DiskType},
		{key: "diskSize", value: awsCloudConfig.DiskSize},
		{key: "diskIops", value: awsCloudConfig.DiskIops},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(workerset, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateAzureWorkerset(workerset *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
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
		if err := setWorkersetFlag(workerset, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateGCEWorkerset(workerset *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
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
		if err := setWorkersetFlag(workerset, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateDigitalOceanWorkerset(workerset *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
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
		if err := setWorkersetFlag(workerset, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateHetznerWorkerset(workerset *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
	var hetznerConfig machinecontroller.HetznerSpec

	if err := json.Unmarshal(cfg, &hetznerConfig); err != nil {
		return err
	}

	flags := []cloudProviderFlags{
		{key: "serverType", value: hetznerConfig.ServerType},
		{key: "datacenter", value: hetznerConfig.Datacenter},
		{key: "location", value: hetznerConfig.Location},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(workerset, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateOpenStackWorkerset(workerset *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
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
		{key: "tags", value: openstackConfig.Tags},
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(workerset, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updatePacketWorkerset(workerset *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
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
		if err := setWorkersetFlag(workerset, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateVSphereWorkerset(workerset *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
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
		if err := setWorkersetFlag(workerset, flag.key, flag.value); err != nil {
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

type commonWorkerConfig struct {
	SSHPublicKeys       []string             `json:"sshPublicKeys"`
	Replicas            *int                 `json:"replicas"`
	OperatingSystem     *string              `json:"operatingSystem"`
	OperatingSystemSpec *operatingSystemSpec `json:"operatingSystemSpec"`

	// +optional
	Network *networkConfig `json:"network,omitempty"`

	// +optional
	OverwriteCloudConfig *string `json:"overwriteCloudConfig,omitempty"`
}

// dnsConfig contains a machine's DNS configuration
type dnsConfig struct {
	Servers []string `json:"servers"`
}

// networkConfig contains a machine's static network configuration
type networkConfig struct {
	CIDR    string    `json:"cidr"`
	Gateway string    `json:"gateway"`
	DNS     dnsConfig `json:"dns"`
}

type operatingSystemSpec struct {
	DistUpgradeOnBoot *bool `json:"distUpgradeOnBoot"`
}

func (c *Config) updateCommonWorkerConfig(workerset *kubeonev1alpha1.WorkerConfig, cfg json.RawMessage) error {
	var cc commonWorkerConfig
	if err := json.Unmarshal(cfg, &cc); err != nil {
		return errors.Wrap(err, "failed to unmarshal common worker config")
	}

	for _, sshKey := range cc.SSHPublicKeys {
		workerset.Config.SSHPublicKeys = append(workerset.Config.SSHPublicKeys, sshKey)
	}

	// Only update if replicas was not configured yet to ensure config from `config.yaml`
	// takes precedence
	if cc.Replicas != nil && workerset.Replicas == nil {
		workerset.Replicas = cc.Replicas
	}

	// Overwrite config from `config.yaml` as the info about the image/AMI/Whatever your cloud calls it
	// comes from Terraform
	if cc.OperatingSystem != nil {
		workerset.Config.OperatingSystem = *cc.OperatingSystem
	}

	osSpecMap := make(map[string]interface{})
	if cc.OperatingSystemSpec.DistUpgradeOnBoot != nil {
		osSpecMap["distUpgradeOnBoot"] = *cc.OperatingSystemSpec.DistUpgradeOnBoot
	}

	if len(osSpecMap) > 0 {
		var err error
		workerset.Config.OperatingSystemSpec, err = json.Marshal(osSpecMap)
		if err != nil {
			return errors.Wrap(err, "unable to update the cloud provider spec")
		}
	}

	if cc.Network != nil {
		workerset.Config.Network.CIDR = cc.Network.CIDR
		workerset.Config.Network.Gateway = cc.Network.Gateway
		workerset.Config.Network.DNS.Servers = cc.Network.DNS.Servers
	}

	if cc.OverwriteCloudConfig != nil {
		workerset.Config.OverwriteCloudConfig = cc.OverwriteCloudConfig
	}

	return nil
}
