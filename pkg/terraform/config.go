package terraform

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kubermatic/kubeone/pkg/config"
)

type controlPlane struct {
	ClusterName       string   `json:"cluster_name"`
	PublicAddress     []string `json:"public_address"`
	PrivateAddress    []string `json:"private_address"`
	Hostnames         []string `json:"hostnames"`
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

	if len(c.PublicAddress) != len(c.PrivateAddress) || len(c.PublicAddress) != len(c.Hostnames) {
		return errors.New("number of public addresses must be equal to number of private addresses and hostnames")
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

		if len(c.Hostnames[i]) == 0 {
			return fmt.Errorf("hostname for host %d is empty", i+1)
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

// Validate checks if the Terraform output conforms to our spec.
func (c *Config) Validate() error {
	planes := c.KubeOneHosts.Value.ControlPlane

	if len(planes) == 0 {
		return errors.New("no control plane specified")
	}

	if len(planes) > 1 {
		return errors.New("more than one control plane specified")
	}

	if err := planes[0].Validate(); err != nil {
		return fmt.Errorf("control plane is invalid: %v", err)
	}

	return nil
}

// Apply adds the terraform configuration options to the given
// cluster config.
func (c *Config) Apply(m *config.Cluster) error {
	if c.KubeOneAPI.Value.Endpoint != "" {
		m.APIServer.Address = c.KubeOneAPI.Value.Endpoint
	}

	hosts := make([]config.HostConfig, 0)
	cp := c.KubeOneHosts.Value.ControlPlane[0]
	sshPort, _ := strconv.Atoi(cp.SSHPort)

	m.Name = cp.ClusterName

	// build up a list of master nodes
	for i, publicIP := range cp.PublicAddress {
		privateIP := publicIP
		if i < len(cp.PrivateAddress) {
			privateIP = cp.PrivateAddress[i]
		}

		// strip domain from hostname
		hostname := strings.Split(cp.Hostnames[i], ".")[0]

		hosts = append(hosts, config.HostConfig{
			ID:                i,
			PublicAddress:     publicIP,
			PrivateAddress:    privateIP,
			Hostname:          hostname,
			SSHUsername:       cp.SSHUser,
			SSHPort:           sshPort,
			SSHPrivateKeyFile: cp.SSHPrivateKeyFile,
			SSHAgentSocket:    cp.SSHAgentSocket,
		})
	}

	if len(hosts) > 0 {
		m.Hosts = hosts
	}

	// if there's a cloud provider specific configuration,
	// apply it to the worker nodes
	if len(c.KubeOneWorkers.Value) > 0 {
		var (
			err           error
			workerConfigs []config.WorkerConfig
		)

		switch m.Provider.Name {
		case config.ProviderNameAWS:
			workerConfigs, err = c.updateAWSWorkers(m.Workers)
		case config.ProviderNameDigitalOcean:
			workerConfigs, err = c.updateDigitalOceanWorkers(m.Workers)
		case config.ProviderNameHetzner:
			workerConfigs, err = c.updateHetznerWorkers(m.Workers)
		case config.ProviderNameOpenStack:
			workerConfigs, err = c.updateOpenStackWorkers(m.Workers)
		case config.ProviderNameVSphere:
			workerConfigs, err = c.updateVSphereWorkers(m.Workers)
		default:
			return errors.New("unknown provider")
		}
		if err != nil {
			return err
		}

		m.Workers = workerConfigs
	}

	return nil
}

func (c *Config) updateAWSWorkers(workers []config.WorkerConfig) ([]config.WorkerConfig, error) {
	for idx, workerset := range workers {
		cloudConfRaw, found := c.KubeOneWorkers.Value[workerset.Name]
		if !found {
			continue
		}
		if len(cloudConfRaw) != 1 {
			// TODO: log warning? error?
			continue
		}

		var awsCloudConfig awsWorkerConfig
		if err := json.Unmarshal(cloudConfRaw[0], &awsCloudConfig); err != nil {
			return nil, err
		}

		if err := setWorkersetFlag(&workerset, "ami", awsCloudConfig.AMI); err != nil {
			return nil, err
		}
		if err := setWorkersetFlag(&workerset, "availabilityZone", awsCloudConfig.AvailabilityZone); err != nil {
			return nil, err
		}
		if err := setWorkersetFlag(&workerset, "instanceProfile", awsCloudConfig.InstanceProfile); err != nil {
			return nil, err
		}
		if err := setWorkersetFlag(&workerset, "region", awsCloudConfig.Region); err != nil {
			return nil, err
		}
		if err := setWorkersetFlag(&workerset, "securityGroupIDs", awsCloudConfig.SecurityGroupIDs); err != nil {
			return nil, err
		}
		if err := setWorkersetFlag(&workerset, "subnetId", awsCloudConfig.SubnetID); err != nil {
			return nil, err
		}
		if err := setWorkersetFlag(&workerset, "vpcId", awsCloudConfig.VPCID); err != nil {
			return nil, err
		}

		workers[idx] = workerset
	}

	return workers, nil
}

func (c *Config) updateDigitalOceanWorkers(workers []config.WorkerConfig) ([]config.WorkerConfig, error) {
	return nil, errors.New("DigitalOcean is not implemented yet")
}

func (c *Config) updateHetznerWorkers(workers []config.WorkerConfig) ([]config.WorkerConfig, error) {
	return nil, errors.New("Hetzner is not implemented yet")
}

func (c *Config) updateOpenStackWorkers(workers []config.WorkerConfig) ([]config.WorkerConfig, error) {
	return nil, errors.New("OpenStack is not implemented yet")
}

func (c *Config) updateVSphereWorkers(workers []config.WorkerConfig) ([]config.WorkerConfig, error) {
	return nil, errors.New("VSphere is not implemented yet")
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
