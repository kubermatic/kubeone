package config

import (
	"net"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"
)

// Cluster describes our entire configuration.
type Cluster struct {
	Name              string                  `json:"name"`
	Hosts             []*HostConfig           `json:"hosts"`
	APIServer         APIServerConfig         `json:"apiserver"`
	Provider          ProviderConfig          `json:"provider"`
	Versions          VersionConfig           `json:"versions"`
	Network           NetworkConfig           `json:"network"`
	Proxy             ProxyConfig             `json:"proxy"`
	Workers           []WorkerConfig          `json:"workers"`
	Backup            BackupConfig            `json:"backup"`
	MachineController MachineControllerConfig `json:"machine_controller"`
	Features          Features                `json:"features"`
}

// DefaultAndValidate checks if the cluster config makes sense.
func (m *Cluster) DefaultAndValidate() error {
	if err := m.Provider.Validate(); err != nil {
		return errors.Wrap(err, "provider configuration is invalid")
	}

	if err := m.Backup.ApplyEnvironment(); err != nil {
		return errors.Wrap(err, "failed to apply backup environment variables")
	}

	if len(m.Hosts) == 0 {
		return errors.New("no master hosts specified")
	}

	m.Hosts[0].IsLeader = true

	for idx, host := range m.Hosts {
		// define a unique ID for each host
		m.Hosts[idx].ID = idx

		if err := host.AddDefaultsAndValidate(); err != nil {
			return errors.WithMessagef(err, "host %d is invalid", idx+1)
		}
	}

	if err := m.MachineController.DefaultAndValidate(m.Provider.Name); err != nil {
		return errors.Wrap(err, "failed to configure machine-controller")
	}
	if *m.MachineController.Deploy {
		for idx, workerset := range m.Workers {
			if err := workerset.Validate(); err != nil {
				return errors.WithMessagef(err, "worker set %d is invalid", idx+1)
			}
		}
	} else if len(m.Workers) > 0 {
		return errors.New("machine-controller deployment is disabled, but configuration still contains worker definitions")
	}

	if err := m.Network.Validate(); err != nil {
		return errors.Wrap(err, "network configuration is invalid")
	}

	if err := m.Backup.Validate(); err != nil {
		return errors.Wrap(err, "backup configuration is invalid")
	}

	if m.APIServer.Address == "" {
		m.APIServer.Address = m.Hosts[0].PublicAddress
	}

	return nil
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

func (m *HostConfig) addDefaults() {
	if len(m.PublicAddress) == 0 && len(m.PrivateAddress) > 0 {
		m.PublicAddress = m.PrivateAddress
	}
	if len(m.PrivateAddress) == 0 && len(m.PublicAddress) > 0 {
		m.PrivateAddress = m.PublicAddress
	}
	if len(m.SSHPrivateKeyFile) == 0 && len(m.SSHAgentSocket) == 0 {
		m.SSHAgentSocket = "env:SSH_AUTH_SOCK"
	}
	if m.SSHUsername == "" {
		m.SSHUsername = "root"
	}
}

// AddDefaultsAndValidate checks if the Config makes sense.
func (m *HostConfig) AddDefaultsAndValidate() error {
	m.addDefaults()

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

// APIServerConfig describes the load balancer address.
type APIServerConfig struct {
	Address string `json:"address"`
}

// ProxyConfig object
type ProxyConfig struct {
	HTTPProxy  string `json:"http_proxy"`
	HTTPSProxy string `json:"https_proxy"`
	NoProxy    string `json:"no_proxy"`
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
	ProviderNameNone         ProviderName = "none"
)

// ProviderConfig describes the cloud provider that is running the machines.
type ProviderConfig struct {
	Name        ProviderName `json:"name"`
	External    bool         `json:"external"`
	CloudConfig string       `json:"cloud_config"`
}

// Validate checks the ProviderConfig for errors
func (p *ProviderConfig) Validate() error {
	switch p.Name {
	case ProviderNameAWS:
	case ProviderNameOpenStack:
		if p.CloudConfig == "" {
			return errors.New("`provider.cloud_config` is required for openstack provider")
		}
	case ProviderNameHetzner:
	case ProviderNameDigitalOcean:
	case ProviderNameVSphere:
	case ProviderNameNone:
	default:
		return errors.Errorf("unknown provider name %q", p.Name)
	}

	return nil
}

// CloudProviderInTree detects is there in-tree cloud provider implementation for specified provider.
// List of in-tree provider can be found here: https://github.com/kubernetes/kubernetes/tree/master/pkg/cloudprovider
func (p *ProviderConfig) CloudProviderInTree() bool {
	switch p.Name {
	case ProviderNameAWS, ProviderNameOpenStack, ProviderNameVSphere:
		return true
	default:
		return false
	}
}

// VersionConfig describes the versions of Kubernetes that is installed.
type VersionConfig struct {
	Kubernetes string `json:"kubernetes"`
}

// Validate semversion of config
func (m *VersionConfig) Validate() error {
	v, err := semver.NewVersion(m.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "unable to parse version string")
	}
	if v.Major() != 1 || v.Minor() < 13 {
		return errors.New("kubernetes versions lower than 1.13 are not supported")
	}
	return nil
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
			return errors.Wrap(err, "invalid pod subnet specified")
		}
	}

	if m.ServiceSubnetVal != "" {
		if _, _, err := net.ParseCIDR(m.ServiceSubnetVal); err != nil {
			return errors.Wrap(err, "invalid service subnet specified")
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
	Replicas *int           `json:"replicas"`
	Config   providerConfig `json:"config"`
}

// Validate checks if the Config makes sense.
func (m *WorkerConfig) Validate() error {
	if m.Name == "" {
		return errors.New("no name given")
	}

	if m.Replicas == nil || *m.Replicas < 1 {
		return errors.New("replicas must be specified and >= 1")
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
		return errors.New("object storage access key must be given")
	}

	if len(m.S3SecretAccessKey) == 0 {
		return errors.New("object storage secret access key must be given")
	}

	if len(m.BucketName) == 0 {
		return errors.New("object storage bucket name must be given")
	}

	if m.Provider != "aws" && m.Provider != "azure" && m.Provider != "gcp" {
		return errors.Errorf(`invalid provider %s; supported values: "aws", "azure" or "gcp"`, m.Provider)
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

// Features switches
type Features struct {
	EnablePodSecurityPolicy bool `json:"enable_pod_security_policy"`
	EnableDynamicAuditLog   bool `json:"enable_dynamic_audit_log"`
}

// MachineControllerConfig controls
type MachineControllerConfig struct {
	Deploy *bool `json:"deploy"`
	// Provider is provider to be used for machine-controller
	// Defaults and must be same as chosen cloud provider, unless cloud provider is set to None
	Provider    ProviderName      `json:"provider"`
	Credentials map[string]string `json:"credentials"`
}

// DefaultAndValidate checks if the machine-controller config makes sense.
func (m *MachineControllerConfig) DefaultAndValidate(cloudProvider ProviderName) error {
	if m.Deploy == nil {
		m.Deploy = boolPtr(true)
	}
	if *m.Deploy == false {
		return nil
	}

	// If ProviderName is not None default to cloud provider and ensure user have not
	// manually provided machine-controller provider different than cloud provider.
	// If ProviderName is None, take user input or default to None.
	if cloudProvider != ProviderNameNone {
		if m.Provider == "" {
			m.Provider = cloudProvider
		}
		if m.Provider != cloudProvider {
			return errors.New("cloud provider must be same as machine-controller provider")
		}
	} else if cloudProvider == ProviderNameNone && m.Provider == "" {
		return errors.New("machine-controller deployed but no provider selected")
	}

	var err error
	m.Credentials, err = m.Provider.ProviderCredentials()
	if err != nil {
		return errors.Wrap(err, "failed to apply cloud provider credentials")
	}

	return nil
}

// ProviderEnvironmentVariable is used to match environment variable used by KubeOne to environment variable used by
// machine-controller.
type ProviderEnvironmentVariable struct {
	Name                  string
	MachineControllerName string
}

// ProviderCredentials match the cloudprovider and parses its credentials from
// environment
func (p ProviderName) ProviderCredentials() (map[string]string, error) {
	switch p {
	case ProviderNameAWS:
		creds := make(map[string]string)
		envCredsProvider := credentials.NewEnvCredentials()
		envCreds, err := envCredsProvider.Get()
		if err != nil {
			return nil, err
		}
		if envCreds.AccessKeyID != "" && envCreds.SecretAccessKey != "" {
			creds["AWS_ACCESS_KEY_ID"] = envCreds.AccessKeyID
			creds["AWS_SECRET_ACCESS_KEY"] = envCreds.SecretAccessKey
			return creds, nil
		}

		// If env fails resort to config file
		configCredsProvider := credentials.NewSharedCredentials("", "")
		configCreds, err := configCredsProvider.Get()
		if err != nil {
			return nil, err
		}
		if configCreds.AccessKeyID != "" && configCreds.SecretAccessKey != "" {
			creds["AWS_ACCESS_KEY_ID"] = configCreds.AccessKeyID
			creds["AWS_SECRET_ACCESS_KEY"] = configCreds.SecretAccessKey
			return creds, nil
		}

		return nil, errors.New("error parsing aws credentials")
	case ProviderNameOpenStack:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "OS_AUTH_URL"},
			{Name: "OS_USERNAME", MachineControllerName: "OS_USER_NAME"},
			{Name: "OS_PASSWORD"},
			{Name: "OS_DOMAIN_NAME"},
			{Name: "OS_TENANT_NAME"},
		})
	case ProviderNameHetzner:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "HCLOUD_TOKEN", MachineControllerName: "HZ_TOKEN"},
		})
	case ProviderNameDigitalOcean:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "DIGITALOCEAN_TOKEN", MachineControllerName: "DO_TOKEN"},
		})
	case ProviderNameVSphere:
		return parseCredentialVariables([]ProviderEnvironmentVariable{
			{Name: "VSPHERE_ADDRESS"},
			{Name: "VSPHERE_USERNAME"},
			{Name: "VSPHERE_PASSWORD"},
		})
	}

	return nil, errors.New("no provider matched")
}

func parseCredentialVariables(envVars []ProviderEnvironmentVariable) (map[string]string, error) {
	creds := make(map[string]string)
	for _, env := range envVars {
		if len(env.MachineControllerName) == 0 {
			env.MachineControllerName = env.Name
		}
		creds[env.MachineControllerName] = strings.TrimSpace(os.Getenv(env.Name))
		if creds[env.MachineControllerName] == "" {
			return nil, errors.Errorf("environment variable %s is not set, but is required", env.Name)
		}
	}
	return creds, nil
}

func boolPtr(val bool) *bool {
	return &val
}
