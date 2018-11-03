package manifest

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

// Manifest describes the terraform output we expect.
type Manifest struct {
	Hosts        []HostManifest       `json:"hosts"`
	LoadBalancer LoadBalancerManifest `json:"loadbalancer"`
	Provider     ProviderManifest     `json:"provider"`
	Versions     VersionManifest      `json:"versions"`
	Network      NetworkManifest      `json:"network"`

	// stuff generated at runtime
	etcdClusterToken string
}

// Merge terraform output on top of manifest
func (m *Manifest) Merge([]byte) error {
	// TODO
	return nil
}

// Validate checks if the manifest makes sense.
func (m *Manifest) Validate() error {
	if len(m.Hosts) == 0 {
		return errors.New("no master hosts specified")
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
	PublicAddress  string `json:"public_address"`
	PrivateAddress string `json:"private_address"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	SSHKeyFile     string `json:"ssh_key_file"`
	SSHSocket      string `json:"ssh_socket"`
}

// EtcdURL with schema
func (m *HostManifest) EtcdURL() string {
	return fmt.Sprintf("https://%s:2379", m.PrivateAddress)
}

// EtcdPeerURL with schema
func (m *HostManifest) EtcdPeerURL() string {
	return fmt.Sprintf("https://%s:2380", m.PrivateAddress)
}

// LoadBalancerManifest describes the load balancer address.
type LoadBalancerManifest struct {
	Address string `json:"address"`
}

// ProviderManifest describes the cloud provider that is running the machines.
type ProviderManifest struct {
	Name        string `json:"name"`
	CloudConfig string `json:"cloud_config"`
}

// VersionManifest describes the versions of Kubernetes and Docker that are installed.
type VersionManifest struct {
	Kubernetes string `json:"kubernetes"`
	Docker     string `json:"docker"`
}

// Etcd version
func (m *VersionManifest) Etcd() string {
	return "3.1.13"
}

// NetworkManifest describes the node network.
type NetworkManifest struct {
	PodSubnet     string
	ServiceSubnet string
	NodePortRange string
}
