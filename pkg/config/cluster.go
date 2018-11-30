package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
)

// Cluster describes our entire configuration.
type Cluster struct {
	Name      string          `yaml:"name"`
	Hosts     []HostConfig    `yaml:"hosts"`
	APIServer APIServerConfig `yaml:"apiserver"`
	Provider  ProviderConfig  `yaml:"provider"`
	Versions  VersionConfig   `yaml:"versions"`
	Network   NetworkConfig   `yaml:"network"`
	Workers   []WorkerConfig  `yaml:"workers"`

	// stuff generated at runtime
	etcdClusterToken string
}

// Validate checks if the cluster config makes sense.
func (m *Cluster) Validate() error {
	if len(m.Hosts) == 0 {
		return errors.New("no master hosts specified")
	}

	for idx, host := range m.Hosts {
		// define a unique ID for each host
		m.Hosts[idx].ID = idx

		if err := host.Validate(); err != nil {
			return fmt.Errorf("host %d is invalid: %v", idx+1, err)
		}
	}

	for idx, workerset := range m.Workers {
		if err := workerset.Validate(); err != nil {
			return fmt.Errorf("worker set %d is invalid: %v", idx+1, err)
		}
	}

	if err := m.Network.Validate(); err != nil {
		return fmt.Errorf("network configuration is invalid: %v", err)
	}

	return nil
}

// EtcdClusterToken returns a randomly generated token.
func (m *Cluster) EtcdClusterToken() (string, error) {
	if m.etcdClusterToken == "" {
		b := make([]byte, 16)

		_, err := rand.Read(b)
		if err != nil {
			return "", err
		}

		m.etcdClusterToken = hex.EncodeToString(b)
	}

	return m.etcdClusterToken, nil
}

// Leader returns the first configured host. Only call this after
// validating the cluster config to ensure a leader exists.
func (m *Cluster) Leader() HostConfig {
	return m.Hosts[0]
}

// Followers returns all but the first configured host. Only call
// this after validating the cluster config to ensure hosts exist.
func (m *Cluster) Followers() []HostConfig {
	return m.Hosts[1:]
}

// HostConfig describes a single master node.
type HostConfig struct {
	ID                int    `yaml:"-"`
	PublicAddress     string `yaml:"public_address"`
	PrivateAddress    string `yaml:"private_address"`
	Hostname          string `yaml:"hostname"`
	SSHPort           int    `yaml:"ssh_port"`
	SSHUsername       string `yaml:"ssh_username"`
	SSHPrivateKeyFile string `yaml:"ssh_private_key_file"`
	SSHAgentSocket    string `yaml:"ssh_agent_socket"`
}

// Validate checks if the Config makes sense.
func (m *HostConfig) Validate() error {
	if len(m.PublicAddress) == 0 {
		return errors.New("no public IP/address given")
	}

	if len(m.PrivateAddress) == 0 {
		return errors.New("no private IP/address given")
	}

	if len(m.Hostname) == 0 {
		return errors.New("no hostname given")
	}

	if len(m.SSHPrivateKeyFile) == 0 && len(m.SSHAgentSocket) == 0 {
		return errors.New("neither SSH private key nor agent socket given, don't know how to authenticate")
	}

	if len(m.SSHUsername) == 0 {
		return errors.New("no SSH username given")
	}

	return nil
}

// EtcdURL with schema
func (m *HostConfig) EtcdURL() string {
	return fmt.Sprintf("https://%s:2379", m.PrivateAddress)
}

// EtcdPeerURL with schema
func (m *HostConfig) EtcdPeerURL() string {
	return fmt.Sprintf("https://%s:2380", m.PrivateAddress)
}

// APIServerConfig describes the load balancer address.
type APIServerConfig struct {
	Address string `yaml:"address"`
}

// ProviderName represents the name of an provider
type ProviderName string

// ProviderName values
const (
	ProviderNameAWS          ProviderName = "aws"
	ProviderNameOpenStack    ProviderName = "openstack"
	ProviderNameHetzner      ProviderName = "hetzner"
	ProviderNameDigitalOcean ProviderName = "digitalocean"
	ProviderNameVSphere      ProviderName = "vshere"
)

// ProviderConfig describes the cloud provider that is running the machines.
type ProviderConfig struct {
	Name        ProviderName      `yaml:"name"`
	CloudConfig string            `yaml:"cloud_config"`
	Credentials map[string]string `yaml:"credentials"`
}

// Validate checks the ProviderConfig for errors
func (p *ProviderConfig) Validate() error {
	switch p.Name {
	case ProviderNameAWS, ProviderNameOpenStack, ProviderNameHetzner, ProviderNameDigitalOcean, ProviderNameVSphere:
	default:
		return fmt.Errorf("unknown provider name %q", p.Name)
	}
	return nil
}

// VersionConfig describes the versions of Kubernetes and Docker that are installed.
type VersionConfig struct {
	Kubernetes string `yaml:"kubernetes"`
	Docker     string `yaml:"docker"`
}

// Etcd version
func (m *VersionConfig) Etcd() string {
	return "3.1.13"
}

// NetworkConfig describes the node network.
type NetworkConfig struct {
	PodSubnetVal     string `yaml:"pod_subnet"`
	ServiceSubnetVal string `yaml:"service_subnet"`
	NodePortRangeVal string `yaml:"node_port_range"`
}

// PodSubnet returns the pod subnet or the default value.
func (m *NetworkConfig) PodSubnet() string {
	if m.PodSubnetVal != "" {
		return m.PodSubnetVal
	}

	return "10.244.0.0/16"
}

// ServiceSubnet returns the service subnet or the default value.
func (m *NetworkConfig) ServiceSubnet() string {
	if m.ServiceSubnetVal != "" {
		return m.ServiceSubnetVal
	}

	return "10.96.0.0/12"
}

// NodePortRange returns the node port range or the default value.
func (m *NetworkConfig) NodePortRange() string {
	if m.NodePortRangeVal != "" {
		return m.NodePortRangeVal
	}

	return "30000-32767"
}

// Validate checks the NetworkConfig for errors
func (p *NetworkConfig) Validate() error {
	if p.PodSubnetVal != "" {
		if _, _, err := net.ParseCIDR(p.PodSubnetVal); err != nil {
			return fmt.Errorf("invalid pod subnet specified: %v", err)
		}
	}

	if p.ServiceSubnetVal != "" {
		if _, _, err := net.ParseCIDR(p.ServiceSubnetVal); err != nil {
			return fmt.Errorf("invalid service subnet specified: %v", err)
		}
	}

	return nil
}

// WorkerConfig describes a set of worker machines.
type WorkerConfig struct {
	Replicas        int                    `yaml:"replicas"`
	Name            string                 `yaml:"name"`
	Spec            map[string]interface{} `yaml:"spec"`
	OperatingSystem struct {
		Name string                 `yaml:"name"`
		Spec map[string]interface{} `yaml:"spec"`
	} `yaml:"operating_system"`
}

// Validate checks if the Config makes sense.
func (m *WorkerConfig) Validate() error {
	if len(m.Name) == 0 {
		return errors.New("no name given")
	}

	if m.Replicas < 1 {
		return errors.New("replicas must be >= 1")
	}

	return nil
}
