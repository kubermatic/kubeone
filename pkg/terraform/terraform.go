package terraform

import (
	"encoding/json"

	"github.com/kubermatic/kubeone/pkg/manifest"
)

// Config represents configuration in the terraform output format
type Config struct {
	KubeOneAPI struct {
		Value struct {
			Endpoint string `json:"endpoint"`
		} `json:"value"`
	} `json:"kubeone_api"`

	KubeOneHosts struct {
		Value struct {
			ControlPlane struct {
				PublicAddress  []string `json:"public_address"`
				PrivateAddress []string `json:"private_address"`
				User           string   `json:"user"`
				SSHKeyFile     string   `json:"ssh_key_file"`
				SSHPort        int      `json:"ssh_port"`
				SSHAgentSocket string   `json:"ssh_agent_socket"`
			} `json:"control_plane"`
		} `json:"value"`
	} `json:"kubeone_hosts"`
}

// NewConfigFromJSON creates a new config object from json
func NewConfigFromJSON(j []byte) (c *Config, err error) {
	c = &Config{}
	return c, json.Unmarshal(j, c)
}

// Apply adds the terraform configuration options to the given manifest
func (c Config) Apply(m *manifest.Manifest) {
	var hosts []manifest.HostManifest
	privateIPs := c.KubeOneHosts.Value.ControlPlane.PrivateAddress
	for i, publicIP := range c.KubeOneHosts.Value.ControlPlane.PublicAddress {
		privateIP := publicIP
		if i < len(privateIPs) {
			privateIP = privateIPs[i]
		}

		hosts = append(hosts, manifest.HostManifest{
			PublicAddress:  publicIP,
			PrivateAddress: privateIP,
			Username:       c.KubeOneHosts.Value.ControlPlane.User,
			SSHKeyFile:     c.KubeOneHosts.Value.ControlPlane.SSHKeyFile,
			Port:           c.KubeOneHosts.Value.ControlPlane.SSHPort,
			SSHSocket:      c.KubeOneHosts.Value.ControlPlane.SSHAgentSocket,
		})
	}

	if len(hosts) > 0 {
		m.Hosts = hosts
	}
	if c.KubeOneAPI.Value.Endpoint != "" {
		m.LoadBalancer.Address = c.KubeOneAPI.Value.Endpoint
	}
}
