package terraform

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/kubermatic/kubeone/pkg/config"
)

type controlPlane struct {
	ClusterName       string   `json:"cluster_name"`
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
			return fmt.Errorf("public address for host %d is empty", i+1)
		}

		if len(c.PrivateAddress[i]) == 0 {
			return fmt.Errorf("private address for host %d is empty", i+1)
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

	cp := c.KubeOneHosts.Value.ControlPlane[0]

	var sshPort int
	var err error
	if cp.SSHPort != "" {
		sshPort, err = strconv.Atoi(cp.SSHPort)
		if err != nil {
			return fmt.Errorf("failed to convert ssh port string '%s' to int: %v", cp.SSHPort, err)
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

	// walk through each of the configured workersets and
	// see if they reference a worker config output from
	// terraform
	for idx, workerset := range cluster.Workers {
		// do we have a matching terraform worker section?
		cloudConfRaw, found := c.KubeOneWorkers.Value[workerset.Name]
		if !found {
			continue
		}

		if len(cloudConfRaw) != 1 {
			// TODO: log warning? error?
			continue
		}
		workerset := workerset

		var err error

		// copy over provider-specific fields in the cloudProviderSpec
		switch cluster.Provider.Name {
		case config.ProviderNameAWS:
			err = c.updateAWSWorkerset(&workerset, cloudConfRaw[0])
		case config.ProviderNameDigitalOcean:
			err = c.updateDigitalOceanWorkerset(&workerset, cloudConfRaw[0])
		case config.ProviderNameHetzner:
			err = c.updateHetznerWorkerset(&workerset, cloudConfRaw[0])
		case config.ProviderNameOpenStack:
			err = c.updateOpenStackWorkerset(&workerset, cloudConfRaw[0])
		case config.ProviderNameVSphere:
			err = c.updateVSphereWorkerset(&workerset, cloudConfRaw[0])
		default:
			return fmt.Errorf("unknown provider %v", cluster.Provider.Name)
		}

		if err != nil {
			return err
		}

		// copy over SSH keys
		err = c.updateSSHKeys(&workerset, cloudConfRaw[0])
		if err != nil {
			return err
		}

		cluster.Workers[idx] = workerset
	}

	return nil
}

func (c *Config) updateAWSWorkerset(workerset *config.WorkerConfig, cfg json.RawMessage) error {
	var awsCloudConfig awsWorkerConfig
	if err := json.Unmarshal(cfg, &awsCloudConfig); err != nil {
		return err
	}

	if err := setWorkersetFlag(workerset, "ami", awsCloudConfig.AMI); err != nil {
		return err
	}
	if err := setWorkersetFlag(workerset, "availabilityZone", awsCloudConfig.AvailabilityZone); err != nil {
		return err
	}
	if err := setWorkersetFlag(workerset, "instanceProfile", awsCloudConfig.InstanceProfile); err != nil {
		return err
	}
	if err := setWorkersetFlag(workerset, "region", awsCloudConfig.Region); err != nil {
		return err
	}
	if err := setWorkersetFlag(workerset, "securityGroupIDs", awsCloudConfig.SecurityGroupIDs); err != nil {
		return err
	}
	if err := setWorkersetFlag(workerset, "subnetId", awsCloudConfig.SubnetID); err != nil {
		return err
	}
	if err := setWorkersetFlag(workerset, "vpcId", awsCloudConfig.VPCID); err != nil {
		return err
	}

	return nil
}

func (c *Config) updateDigitalOceanWorkerset(_ *config.WorkerConfig, _ json.RawMessage) error {
	return errors.New("DigitalOcean is not implemented yet")
}

func (c *Config) updateHetznerWorkerset(_ *config.WorkerConfig, _ json.RawMessage) error {
	return errors.New("Hetzner is not implemented yet")
}

func (c *Config) updateOpenStackWorkerset(_ *config.WorkerConfig, _ json.RawMessage) error {
	return errors.New("OpenStack is not implemented yet")
}

func (c *Config) updateVSphereWorkerset(_ *config.WorkerConfig, _ json.RawMessage) error {
	return errors.New("VSphere is not implemented yet")
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
	case []string:
		if len(s) == 0 {
			return nil
		}
	default:
		return errors.New("unsupported type")
	}

	// update CloudProviderSpec ONLY IF given terraform output is absent in
	// original CloudProviderSpec
	if _, exists := w.Config.CloudProviderSpec[name]; !exists {
		w.Config.CloudProviderSpec[name] = value
	}

	return nil
}

type sshKeyWorkerConfig struct {
	SSHPublicKeys []string `json:"sshPublicKeys"`
}

func (c *Config) updateSSHKeys(workerset *config.WorkerConfig, cfg json.RawMessage) error {
	var cc sshKeyWorkerConfig
	if err := json.Unmarshal(cfg, &cc); err != nil {
		return err
	}

	if len(workerset.Config.SSHPublicKeys) == 0 {
		workerset.Config.SSHPublicKeys = cc.SSHPublicKeys
	}

	return nil
}
