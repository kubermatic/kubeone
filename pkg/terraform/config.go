package terraform

import (
	"encoding/json"
	"strconv"

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
			ControlPlane []struct {
				PublicAddress    []string `json:"public_address"`
				PrivateAddress   []string `json:"private_address"`
				User             string   `json:"user"`
				SSHPublicKeyFile string   `json:"ssh_public_key_file"`
				SSHPort          string   `json:"ssh_port"`
				SSHAgentSocket   string   `json:"ssh_agent_socket"`
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
	if c.KubeOneAPI.Value.Endpoint != "" {
		m.APIServer.Address = c.KubeOneAPI.Value.Endpoint
	}

	var hosts []manifest.HostManifest
	cp := c.KubeOneHosts.Value.ControlPlane[0]
	sshPort, _ := strconv.Atoi(cp.SSHPort)

	privateIPs := cp.PrivateAddress

	for i, publicIP := range cp.PublicAddress {
		privateIP := publicIP
		if i < len(privateIPs) {
			privateIP = privateIPs[i]
		}

		hosts = append(hosts, manifest.HostManifest{
			PublicAddress:    publicIP,
			PrivateAddress:   privateIP,
			Username:         cp.User,
			SSHPublicKeyFile: cp.SSHPublicKeyFile,
			Port:             sshPort,
			SSHSocket:        cp.SSHAgentSocket,
		})
	}

	if len(hosts) > 0 {
		m.Hosts = hosts
	}
}
