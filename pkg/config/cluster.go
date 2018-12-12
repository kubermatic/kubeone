package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
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
	Backup    BackupConfig    `yaml:"backup"`

	// stuff generated at runtime
	etcdClusterToken string
}

// ApplyEnvironment overwrites empty values inside the configuration for
// certain subsections of a cluster configuration, like the provider or
// Ark credentials.
func (m *Cluster) ApplyEnvironment() error {
	if err := m.Provider.ApplyEnvironment(); err != nil {
		return fmt.Errorf("failed to apply cloud provider credentials: %v", err)
	}

	if err := m.Backup.ApplyEnvironment(); err != nil {
		return fmt.Errorf("failed to apply backup environment variables: %v", err)
	}

	return nil
}

func (m *Cluster) AddDefaults() error {
	for i, _ := range m.Hosts {
		if err := m.Hosts[i].AddDefaults(); err != nil {
			return fmt.Errorf("host %d could not be defaulted: %v", i+1, err)
		}
	}

	return nil
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

	if err := m.Backup.Validate(); err != nil {
		return fmt.Errorf("backup configuration is invalid: %v", err)
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

func (m *HostConfig) AddDefaults() error {
	if len(m.PublicAddress) == 0 && len(m.PrivateAddress) > 0 {
		m.PublicAddress = m.PrivateAddress
	}
	if len(m.PrivateAddress) == 0 && len(m.PublicAddress) > 0 {
		m.PrivateAddress = m.PublicAddress
	}
	if len(m.Hostname) == 0 {
		m.Hostname = m.PublicAddress
	}
	if len(m.SSHPrivateKeyFile) == 0 && len(m.SSHAgentSocket) == 0 {
		m.SSHAgentSocket = "env:SSH_AUTH_SOCK"
	}
	return nil
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

func (p ProviderName) CredentialsEnvironmentVariables() []string {
	switch p {
	case ProviderNameAWS:
		return []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"}
	case ProviderNameOpenStack:
		return []string{"OS_AUTH_URL", "OS_USER_NAME", "OS_PASSWORD", "OS_DOMAIN_NAME", "OS_TENANT_NAME"}
	case ProviderNameHetzner:
		return []string{"HZ_TOKEN"}
	case ProviderNameDigitalOcean:
		return []string{"DO_TOKEN"}
	case ProviderNameVSphere:
		return []string{"VSPHERE_ADDRESS", "VSPHERE_USERNAME", "VSPHERE_PASSWORD"}
	}

	return nil
}

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

	for _, varName := range p.Name.CredentialsEnvironmentVariables() {
		if p.Credentials[varName] == "" {
			return fmt.Errorf("environment variable %s is not set", varName)
		}
	}

	return nil
}

// ApplyEnvironment reads cloud provider credentials from
// environment variables.
func (p *ProviderConfig) ApplyEnvironment() error {
	if p.Credentials == nil {
		p.Credentials = make(map[string]string)
	}

	for _, varName := range p.Name.CredentialsEnvironmentVariables() {
		p.Credentials[varName] = strings.TrimSpace(os.Getenv(varName))
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
func (m *NetworkConfig) Validate() error {
	if m.PodSubnetVal != "" {
		if _, _, err := net.ParseCIDR(m.PodSubnetVal); err != nil {
			return fmt.Errorf("invalid pod subnet specified: %v", err)
		}
	}

	if m.ServiceSubnetVal != "" {
		if _, _, err := net.ParseCIDR(m.ServiceSubnetVal); err != nil {
			return fmt.Errorf("invalid service subnet specified: %v", err)
		}
	}

	return nil
}

type providerConfig struct {
	CloudProviderSpec   map[string]interface{} `yaml:"cloudProviderSpec"`
	Labels              map[string]string      `yaml:"labels"`
	SSHPublicKeys       []string               `yaml:"sshPublicKeys"`
	OperatingSystem     string                 `yaml:"operatingSystem"`
	OperatingSystemSpec map[string]interface{} `yaml:"operatingSystemSpec"`
}

// WorkerConfig describes a set of worker machines.
type WorkerConfig struct {
	Name     string         `yaml:"name"`
	Replicas int            `yaml:"replicas"`
	Config   providerConfig `yaml:"config"`
}

// Validate checks if the Config makes sense.
func (m *WorkerConfig) Validate() error {
	if m.Name == "" {
		return errors.New("no name given")
	}

	if m.Replicas < 1 {
		return errors.New("replicas must be >= 1")
	}

	return nil
}

// BackupConfig describes where and how to store Ark backups
type BackupConfig struct {
	// Provider is provider for buckets and volume snapshots.
	// Possible values are: AWS (includes compatible AWS S3 storages), Azure and GCP
	// TODO(xmudrii): By default uses specified control plane provider if compatible with Ark
	Provider string `yaml:"provider"`

	// S3AccessKey is Access Key used to access backups S3 bucket.
	// This variable is sourced from BACKUP_AWS_ACCESS_KEY_ID,
	// or if unset from AWS_ACCESS_KEY_ID environment variable
	S3AccessKey string `yaml:"s3_access_key"`
	// S3SecretAccessKey is secret key used to access backups S3 bucket.
	// This variable is sourced from BACKUP_AWS_SECRET_ACCESS_KEY environment variable,
	// or if unset from AWS_SECRET_ACCESS_KEY environment variable
	S3SecretAccessKey string `yaml:"s3_secret_access_key"`

	// BucketName is name of the S3 bucket where backups are stored
	BucketName string `yaml:"bucket_name"`

	// BackupStorageConfig is optional configuration depending on the provider specified
	// Details: https://heptio.github.io/ark/v0.10.0/api-types/backupstoragelocation.html
	BackupStorageConfig map[string]string `yaml:"backup_storage_config"`

	// VolumesSnapshotConfig is optional configuration depending on the provider specified
	// Details: https://heptio.github.io/ark/v0.10.0/api-types/volumesnapshotlocation.html
	VolumesSnapshotConfig map[string]string `yaml:"volumes_snapshot_region"`
}

// Enabled checks if a provider is set and Ark should be deployed.
func (m *BackupConfig) Enabled() bool {
	return m.Provider != ""
}

// Validate valides the BackupConfig structure, ensuring credentials and bucket name are provided
func (m *BackupConfig) Validate() error {
	// if the backup is not enabled, nothing else matters
	if !m.Enabled() {
		return nil
	}

	if len(m.S3AccessKey) == 0 {
		return errors.New("S3 access key must be given")
	}

	if len(m.S3SecretAccessKey) == 0 {
		return errors.New("S3 secret access key must be given")
	}

	if len(m.BucketName) == 0 {
		return errors.New("S3 bucket name must be given")
	}

	if m.Provider != "aws" && m.Provider != "azure" && m.Provider != "gcp" {
		return fmt.Errorf("invalid provider %s; supported values: \"aws\", \"azure\" or \"gcp\"", m.Provider)
	}

	return nil
}

// ApplyEnvironment reads credentials from environment variables,
// returning an error if a required variable is not set.
func (m *BackupConfig) ApplyEnvironment() error {
	const envPrefix = "env:"

	if strings.HasPrefix(m.S3AccessKey, envPrefix) {
		envName := strings.TrimPrefix(m.S3AccessKey, envPrefix)
		m.S3AccessKey = os.Getenv(envName)
	}

	if strings.HasPrefix(m.S3SecretAccessKey, envPrefix) {
		envName := strings.TrimPrefix(m.S3AccessKey, envPrefix)
		m.S3SecretAccessKey = os.Getenv(envName)
	}

	return nil
}
