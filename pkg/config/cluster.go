package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

// Cluster describes our entire configuration.
type Cluster struct {
	Hosts     []HostConfig    `yaml:"hosts"`
	APIServer APIServerConfig `yaml:"apiserver"`
	Provider  ProviderConfig  `yaml:"provider"`
	Versions  VersionConfig   `yaml:"versions"`
	Network   NetworkConfig   `yaml:"network"`

	// stuff generated at runtime
	etcdClusterToken string
}

// Validate checks if the cluster config makes sense.
func (m *Cluster) Validate() error {
	if len(m.Hosts) == 0 {
		return errors.New("no master hosts specified")
	}

	for idx, host := range m.Hosts {
		if err := host.Validate(); err != nil {
			return fmt.Errorf("host %d is invalid: %v", idx+1, err)
		}
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

// HostConfig describes a single master node.
type HostConfig struct {
	PublicAddress     string `yaml:"public_address"`
	PrivateAddress    string `yaml:"private_address"`
	PublicDNS         string `yaml:"public_dns"`
	PrivateDNS        string `yaml:"private_dns"`
	Hostname          string `yaml:"hostname"`
	SSHPort           int    `yaml:"ssh_port"`
	SSHUsername       string `yaml:"ssh_username"`
	SSHPrivateKeyFile string `yaml:"ssh_private_key_file"`
	SSHAgentSocket    string `yaml:"ssh_agent_socket"`
}

// Validate checks if the Config makes sense.
func (m *HostConfig) Validate() error {
	if len(m.SSHPrivateKeyFile) == 0 && len(m.SSHAgentSocket) == 0 {
		return errors.New("neither SSH private key nor agent socket given, don't know how to authenticate")
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

// ProviderConfig describes the cloud provider that is running the machines.
type ProviderConfig struct {
	Name        string `yaml:"name"`
	CloudConfig string `yaml:"cloud_config"`
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
