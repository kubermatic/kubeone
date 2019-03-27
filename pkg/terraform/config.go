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
	"strconv"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
)

type controlPlane struct {
	ClusterName       string   `json:"cluster_name"`
	CloudProvider     *string  `json:"cloud_provider"`
	PublicAddress     []string `json:"public_address"`
	PrivateAddress    []string `json:"private_address"`
	SSHUser           string   `json:"ssh_user"`
	SSHPort           string   `json:"ssh_port"`
	SSHPrivateKeyFile string   `json:"ssh_private_key_file"`
	SSHAgentSocket    string   `json:"ssh_agent_socket"`
}

// Validate checks if the control plane conforms to our spec.
func (c *controlPlane) Validate() error {
	if len(c.PublicAddress) < 3 {
		return errors.New("must specify a unique cluster name")
	}

	if len(c.PublicAddress) != len(c.PrivateAddress) {
		return errors.New("number of public addresses must be equal to number of private addresses")
	}

	if len(c.PublicAddress) < 3 {
		return errors.New("must specify at least three public addresses")
	}

	for i := 0; i < len(c.PublicAddress); i++ {
		if len(c.PublicAddress[i]) == 0 {
			return errors.Errorf("public address for host %d is empty", i+1)
		}

		if len(c.PrivateAddress[i]) == 0 {
			return errors.Errorf("private address for host %d is empty", i+1)
		}
	}

	return nil
}

type awsWorkerConfig struct {
	AMI              string   `json:"ami"`
	AvailabilityZone string   `json:"availabilityZone"`
	InstanceProfile  string   `json:"instanceProfile"`
	Region           string   `json:"region"`
	SecurityGroupIDs []string `json:"securityGroupIDs"`
	SubnetID         string   `json:"subnetId"`
	VPCID            string   `json:"vpcId"`
	InstanceType     *string  `json:"instanceType"`
	DiskSize         *int     `json:"diskSize"`
}

// TODO: Add option for sourcing bool values (private networking, monitoring...)
type doWorkerConfig struct {
	DropletSize string `json:"droplet_size"`
	Region      string `json:"region"`
}

type openStackWorkerConfig struct {
	Image            string   `json:"image"`
	Flavor           string   `json:"flavor"`
	SecurityGroups   []string `json:"securityGroups"`
	FloatingIPPool   string   `json:"floatingIPPool"`
	AvailabilityZone string   `json:"availabilityZone"`
	Network          string   `json:"network"`
	Subnet           string   `json:"subnet"`
}

type gceWorkerConfig struct {
	DiskSize    int    `json:"diskSize"`
	DiskType    string `json:"diskType"`
	MachineType string `json:"machineType"`
	Network     string `json:"network"`
	Subnetwork  string `json:"subnetwork"`
	Zone        string `json:"zone"`
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
			ControlPlane []controlPlane `json:"control_plane"`
		} `json:"value"`
	} `json:"kubeone_hosts"`

	KubeOneWorkers struct {
		Value map[string][]json.RawMessage `json:"value"`
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
func (c *Config) Apply(cluster *config.Cluster) error {
	if c.KubeOneAPI.Value.Endpoint != "" {
		cluster.APIServer.Address = c.KubeOneAPI.Value.Endpoint
	}

	if len(c.KubeOneHosts.Value.ControlPlane) == 0 {
		return errors.New("no control plane hosts are given")
	}

	cp := c.KubeOneHosts.Value.ControlPlane[0]

	if cp.CloudProvider != nil {
		cluster.Provider.Name = config.ProviderName(*cp.CloudProvider)
	}

	var sshPort int
	var err error
	if cp.SSHPort != "" {
		sshPort, err = strconv.Atoi(cp.SSHPort)
		if err != nil {
			return errors.Wrapf(err, "failed to convert ssh port string %q to int", cp.SSHPort)
		}
	}

	cluster.Name = cp.ClusterName

	// build up a list of master nodes
	hosts := make([]*config.HostConfig, 0)
	for i, publicIP := range cp.PublicAddress {
		privateIP := publicIP
		if i < len(cp.PrivateAddress) {
			privateIP = cp.PrivateAddress[i]
		}

		hosts = append(hosts, &config.HostConfig{
			ID:                i,
			PublicAddress:     publicIP,
			PrivateAddress:    privateIP,
			SSHUsername:       cp.SSHUser,
			SSHPort:           sshPort,
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
		if len(workersetValue) != 1 {
			// TODO: log warning? error?
			continue
		}

		var existingWorkerSet *config.WorkerConfig
		for idx, workerset := range cluster.Workers {
			if workerset.Name == workersetName {
				existingWorkerSet = &cluster.Workers[idx]
				break
			}
		}
		if existingWorkerSet == nil {
			// Append copies the object when its a literal and not a pointer, hence
			// we have to first append, then create a pointer to the appended object
			cluster.Workers = append(cluster.Workers, config.WorkerConfig{Name: workersetName})
			existingWorkerSet = &cluster.Workers[len(cluster.Workers)-1]
		}

		switch cluster.Provider.Name {
		case config.ProviderNameAWS:
			err = c.updateAWSWorkerset(existingWorkerSet, workersetValue[0])
		case config.ProviderNameGCE:
			err = c.updateGCEWorkerset(existingWorkerSet, workersetValue[0])
		case config.ProviderNameDigitalOcean:
			err = c.updateDigitalOceanWorkerset(existingWorkerSet, workersetValue[0])
		case config.ProviderNameHetzner:
			err = c.updateHetznerWorkerset(existingWorkerSet, workersetValue[0])
		case config.ProviderNameOpenStack:
			err = c.updateOpenStackWorkerset(existingWorkerSet, workersetValue[0])
		case config.ProviderNameVSphere:
			err = c.updateVSphereWorkerset(existingWorkerSet, workersetValue[0])
		default:
			return errors.Errorf("unknown provider %v", cluster.Provider.Name)
		}

		if err != nil {
			return errors.Wrapf(err, "failed to update provider-specific config for workerset %q from terraform config", workersetName)
		}

		// copy over common config
		if err = c.updateCommonWorkerConfig(existingWorkerSet, workersetValue[0]); err != nil {
			return errors.Wrap(err, "failed to update common config from terraform config")
		}
	}

	return nil
}

func (c *Config) updateAWSWorkerset(workerset *config.WorkerConfig, cfg json.RawMessage) error {
	var awsCloudConfig awsWorkerConfig
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
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(workerset, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	// We effectively hardcode it here because we have no sane way to check if it was already defined
	// as workerset.Config is a map[string]interface{}
	// TODO: Use imported provicerConfig structs for workset.Config
	// TODO: Add defaulting in the machine-controller for this and remove it here
	if err := setWorkersetFlag(workerset, "diskType", "gp2"); err != nil {
		return errors.WithStack(err)
	}

	// We can not check if its defined in the workset already as workerset.Config is a map[string]interface{}
	// TODO: Use imported provicerConfig structs for workset.Config
	if awsCloudConfig.DiskSize != nil {
		if err := setWorkersetFlag(workerset, "diskSize", *awsCloudConfig.DiskSize); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateGCEWorkerset(workerset *config.WorkerConfig, cfg json.RawMessage) error {
	var gceCloudConfig gceWorkerConfig
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
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(workerset, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateDigitalOceanWorkerset(workerset *config.WorkerConfig, cfg json.RawMessage) error {
	var doCloudConfig doWorkerConfig
	if err := json.Unmarshal(cfg, &doCloudConfig); err != nil {
		return errors.WithStack(err)
	}

	if err := setWorkersetFlag(workerset, "size", doCloudConfig.DropletSize); err != nil {
		return errors.WithStack(err)
	}

	if err := setWorkersetFlag(workerset, "region", doCloudConfig.Region); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *Config) updateHetznerWorkerset(_ *config.WorkerConfig, _ json.RawMessage) error {
	return errors.New("cloudprovider Hetzner is not implemented yet")
}

func (c *Config) updateOpenStackWorkerset(workerset *config.WorkerConfig, cfg json.RawMessage) error {
	var openstackConfig openStackWorkerConfig
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
	}

	for _, flag := range flags {
		if err := setWorkersetFlag(workerset, flag.key, flag.value); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (c *Config) updateVSphereWorkerset(_ *config.WorkerConfig, _ json.RawMessage) error {
	return errors.New("cloudprovider VSphere is not implemented yet")
}

func setWorkersetFlag(w *config.WorkerConfig, name string, value interface{}) error {
	// ignore empty values (i.e. not set in terraform output)
	switch s := value.(type) {
	case int:
		if s == 0 {
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
	default:
		return errors.New("unsupported type")
	}

	// update CloudProviderSpec ONLY IF given terraform output is absent in
	// original CloudProviderSpec
	if w.Config.CloudProviderSpec == nil {
		w.Config.CloudProviderSpec = map[string]interface{}{}
	}
	if _, exists := w.Config.CloudProviderSpec[name]; !exists {
		w.Config.CloudProviderSpec[name] = value
	}

	return nil
}

type commonWorkerConfig struct {
	SSHPublicKeys   []string `json:"sshPublicKeys"`
	Replicas        *int     `json:"replicas"`
	OperatingSystem *string  `json:"operatingSystem"`
}

func (c *Config) updateCommonWorkerConfig(workerset *config.WorkerConfig, cfg json.RawMessage) error {
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

	return nil
}
