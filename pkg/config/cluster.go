package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

// Cluster describes our entire configuration.
type Cluster struct {
	Name              string                  `json:"name"`
	Hosts             []*HostConfig           `json:"hosts"`
	APIServer         APIServerConfig         `json:"apiserver"`
	ETCD              ETCDConfig              `json:"etcd"`
	Provider          ProviderConfig          `json:"provider"`
	Versions          VersionConfig           `json:"versions"`
	Network           NetworkConfig           `json:"network"`
	Workers           []WorkerConfig          `json:"workers"`
	Backup            BackupConfig            `json:"backup"`
	MachineController MachineControllerConfig `json:"machine_controller"`
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
		m.ETCD.Version = "3.2.24"
	}

	if m.ETCD.Version != "3.2.24" {
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

	if err := m.MachineController.DefaultAndValidate(); err != nil {
		return fmt.Errorf("failed to configure machine-controller: %v", err)
	}

	if *m.MachineController.Deploy {
		for idx, workerset := range m.Workers {
			if err := workerset.Validate(); err != nil {
				return fmt.Errorf("worker set %d is invalid: %v", idx+1, err)
			}
		}
	} else if len(m.Workers) > 0 {
		return errors.New("machine-controller deployment is disabled, but configuration still contains worker definitions")
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

// HostConfig describes a single master node.
type HostConfig struct {
	ID                int    `json:"-"`
	PublicAddress     string `json:"public_address"`
	PrivateAddress    string `json:"private_address"`
	SSHPort           int    `json:"ssh_port"`
	SSHUsername       string `json:"ssh_username"`
	SSHPrivateKeyFile string `json:"ssh_private_key_file"`
	SSHAgentSocket    string `json:"ssh_agent_socket"`

	// runtime information
	Hostname        string `json:"-"`
	OperatingSystem string `json:"-"`
	IsLeader        bool   `json:"-"`
}

func (m *HostConfig) addDefaults() error {
	if len(m.PublicAddress) == 0 && len(m.PrivateAddress) > 0 {
		m.PublicAddress = m.PrivateAddress
	}
	if len(m.PrivateAddress) == 0 && len(m.PublicAddress) > 0 {
		m.PrivateAddress = m.PublicAddress
	}
	if len(m.SSHPrivateKeyFile) == 0 && len(m.SSHAgentSocket) == 0 {
		m.SSHAgentSocket = "env:SSH_AUTH_SOCK"
	}
	return nil
}

// AddDefaultsAndValidate checks if the Config makes sense.
func (m *HostConfig) AddDefaultsAndValidate() error {
	if err := m.addDefaults(); err != nil {
		return fmt.Errorf("defaulting failed: %v", err)
	}

	if len(m.PublicAddress) == 0 {
		return errors.New("no public IP/address given")
	}

	if len(m.PrivateAddress) == 0 {
		return errors.New("no private IP/address given")
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
	Address string `json:"address"`
}

type ETCDConfig struct {
	Version string `json:"address"`
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
	Name        ProviderName      `json:"name"`
	CloudConfig string            `json:"cloud_config"`
	Credentials map[string]string `json:"credentials"`
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
	Kubernetes string `json:"kubernetes"`
	Docker     string `json:"docker"`
}

// Etcd version
func (m *VersionConfig) Etcd() string {
	return "3.1.13"
}

// NetworkConfig describes the node network.
type NetworkConfig struct {
	PodSubnetVal     string `json:"pod_subnet"`
	ServiceSubnetVal string `json:"service_subnet"`
	NodePortRangeVal string `json:"node_port_range"`
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
	CloudProviderSpec   map[string]interface{} `json:"cloudProviderSpec"`
	Labels              map[string]string      `json:"labels"`
	SSHPublicKeys       []string               `json:"sshPublicKeys"`
	OperatingSystem     string                 `json:"operatingSystem"`
	OperatingSystemSpec map[string]interface{} `json:"operatingSystemSpec"`
}

// WorkerConfig describes a set of worker machines.
type WorkerConfig struct {
	Name     string         `json:"name"`
	Replicas int            `json:"replicas"`
	Config   providerConfig `json:"config"`
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
	Provider string `json:"provider"`

	// S3AccessKey is Access Key used to access backups S3 bucket.
	// This variable is sourced from BACKUP_AWS_ACCESS_KEY_ID,
	// or if unset from AWS_ACCESS_KEY_ID environment variable
	S3AccessKey string `json:"s3_access_key"`
	// S3SecretAccessKey is secret key used to access backups S3 bucket.
	// This variable is sourced from BACKUP_AWS_SECRET_ACCESS_KEY environment variable,
	// or if unset from AWS_SECRET_ACCESS_KEY environment variable
	S3SecretAccessKey string `json:"s3_secret_access_key"`

	// BucketName is name of the S3 bucket where backups are stored
	BucketName string `json:"bucket_name"`

	// BackupStorageConfig is optional configuration depending on the provider specified
	// Details: https://heptio.github.io/ark/v0.10.0/api-types/backupstoragelocation.html
	BackupStorageConfig map[string]string `json:"backup_storage_config"`

	// VolumesSnapshotConfig is optional configuration depending on the provider specified
	// Details: https://heptio.github.io/ark/v0.10.0/api-types/volumesnapshotlocation.html
	VolumesSnapshotConfig map[string]string `json:"volumes_snapshot_region"`
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
		envName := strings.TrimPrefix(m.S3SecretAccessKey, envPrefix)
		m.S3SecretAccessKey = os.Getenv(envName)
	}

	return nil
}

type MachineControllerConfig struct {
	Deploy *bool `json:"deploy"`
}

// DefaultAndValidate checks if the machine-controller config makes sense.
func (m *MachineControllerConfig) DefaultAndValidate() error {
	if m.Deploy == nil {
		m.Deploy = boolPtr(true)
	}

	return nil
}

func boolPtr(val bool) *bool {
	return &val
}
