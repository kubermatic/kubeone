package manifest

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

// Manifest describes the terraform output we expect.
type Manifest struct {
	Hosts     []HostManifest    `yaml:"hosts"`
	APIServer APIServerManifest `yaml:"apiserver"`
	Provider  ProviderManifest  `yaml:"provider"`
	Versions  VersionManifest   `yaml:"versions"`
	Network   NetworkManifest   `yaml:"network"`

	// stuff generated at runtime
	etcdClusterToken string
}

// Validate checks if the manifest makes sense.
func (m *Manifest) Validate() error {
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
func (m *Manifest) EtcdClusterToken() (string, error) {
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

// HostManifest describes a single master node.
type HostManifest struct {
	PublicAddress     string `yaml:"public_address"`
	PrivateAddress    string `yaml:"private_address"`
	SSHPort           int    `yaml:"ssh_port"`
	SSHUsername       string `yaml:"ssh_username"`
	SSHPrivateKeyFile string `yaml:"ssh_private_key_file"`
	SSHAgentSocket    string `yaml:"ssh_agent_socket"`
}

// Validate checks if the manifest makes sense.
func (m *HostManifest) Validate() error {
	if len(m.SSHPrivateKeyFile) == 0 && len(m.SSHAgentSocket) == 0 {
		return errors.New("neither SSH private key nor agent socket given, don't know how to authenticate")
	}

	return nil
}

// EtcdURL with schema
func (m *HostManifest) EtcdURL() string {
	return fmt.Sprintf("https://%s:2379", m.PrivateAddress)
}

// EtcdPeerURL with schema
func (m *HostManifest) EtcdPeerURL() string {
	return fmt.Sprintf("https://%s:2380", m.PrivateAddress)
}

// APIServerManifest describes the load balancer address.
type APIServerManifest struct {
	Address string `yaml:"address"`
}

// ProviderManifest describes the cloud provider that is running the machines.
type ProviderManifest struct {
	Name        string `yaml:"name"`
	CloudConfig string `yaml:"cloud_config"`
}

// VersionManifest describes the versions of Kubernetes and Docker that are installed.
type VersionManifest struct {
	Kubernetes string `yaml:"kubernetes"`
	Docker     string `yaml:"docker"`
}

// Etcd version
func (m *VersionManifest) Etcd() string {
	return "3.1.13"
}

// NetworkManifest describes the node network.
type NetworkManifest struct {
	PodSubnet     string `yaml:"pod_subnet"`
	ServiceSubnet string `yaml:"service_subnet"`
	NodePortRange string `yaml:"node_port_range"`
}
