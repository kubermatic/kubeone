package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Cluster describes our entire configuration.
type Cluster struct {
	Name      string          `json:"name"`
	Hosts     []*HostConfig   `json:"hosts"`
	APIServer APIServerConfig `json:"apiserver"`
	ETCD      ETCDConfig      `json:"etcd"`
	Provider  ProviderConfig  `json:"provider"`
	Versions  VersionConfig   `json:"versions"`
	Network   NetworkConfig   `json:"network"`
	Workers   []WorkerConfig  `json:"workers"`
	Backup    BackupConfig    `json:"backup"`
}

const (
	envSupportedETCDVersion     = "K1_ETCD_VERSION"
	defaultSupportedETCDVersion = "3.2.24"
)

func getSupportedETCDVersion() []string {
	if ss := os.Getenv(envSupportedETCDVersion); ss != "" {
		return strings.Split(ss, ",")
	}
	return []string{defaultSupportedETCDVersion}
}

// DefaultAndValidate checks if the cluster config makes sense.
func (m *Cluster) DefaultAndValidate() error {
	if err := m.Provider.ApplyEnvironment(); err != nil {
		return fmt.Errorf("failed to apply cloud provider credentials: %v", err)
	}

	if err := m.Backup.ApplyEnvironment(); err != nil {
		return fmt.Errorf("failed to apply backup environment variables: %v", err)
	}

	if len(m.Hosts) == 0 {
		return errors.New("no master hosts specified")
	}

	if m.ETCD.Version == "" {
		m.ETCD.Version = defaultSupportedETCDVersion
	}

	etcdSupported := false
	for _, v := range getSupportedETCDVersion() {
		if m.ETCD.Version == v {
			etcdSupported = true
			break
		}
	}

	if !etcdSupported {
		return fmt.Errorf("Only supported etcd version is 3.2.24")
	}

	m.EtcdClusterToken()

	m.Hosts[0].IsLeader = true

	for idx, host := range m.Hosts {
		// define a unique ID for each host
		m.Hosts[idx].ID = idx

		if err := host.AddDefaultsAndValidate(); err != nil {
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

	if err := m.Backup.Validate(); err != nil {
		return fmt.Errorf("backup configuration is invalid: %v", err)
	}

	return nil
}

// EtcdClusterToken returns the cluster name
// It must be deterministic across multiple runs
func (m *Cluster) EtcdClusterToken() string {
	return m.Name
}

// Leader returns the first configured host. Only call this after
// validating the cluster config to ensure a leader exists.
func (m *Cluster) Leader() (*HostConfig, error) {
	for i := range m.Hosts {
		if m.Hosts[i].IsLeader {
			return m.Hosts[i], nil
		}
	}
	return nil, errors.New("leader not found")
}

// Followers returns all but the first configured host. Only call
// this after validating the cluster config to ensure hosts exist.
func (m *Cluster) Followers() []*HostConfig {
	return m.Hosts[1:]
}

// APIServerConfig describes the load balancer address.
type APIServerConfig struct {
	Address string `json:"address"`
}

type ETCDConfig struct {
	Version string `json:"address"`
}

// VersionConfig describes the versions of Kubernetes and Docker that are installed.
type VersionConfig struct {
	Kubernetes string `json:"kubernetes"`
	Docker     string `json:"docker"`
}

// Etcd version
func (m *VersionConfig) Etcd() string {
	return "3.1.13"
}
