/*
Copyright 2019 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubeone

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubeOneCluster is KubeOne Cluster API Schema
type KubeOneCluster struct {
	metav1.TypeMeta `json:",inline"`

	// Name is the name of the cluster
	Name string `json:"name"`
	// Hosts describes the control plane nodes and how to access them
	Hosts []HostConfig `json:"hosts,omitempty"`
	// StaticWorkers allows the user to define a list of nodes as workers that are not managed by MachineController
	StaticWorkers []HostConfig `json:"staticWorkers,omitempty"`
	// APIEndpoint are pairs of address and port used to communicate with the Kubernetes API
	APIEndpoint APIEndpoint `json:"apiEndpoint,omitempty"`
	// CloudProvider configures the cloud provider specific features
	CloudProvider CloudProviderSpec `json:"cloudProvider,omitempty"`
	// Versions defines which Kubernetes version will be installed
	Versions VersionConfig `json:"versions,omitempty"`
	// ClusterNetwork configures the in-cluster networking
	ClusterNetwork ClusterNetworkConfig `json:"clusterNetwork,omitempty"`
	// Proxy configures proxy used while installing Kubernetes and by the Docker daemon
	Proxy ProxyConfig `json:"proxy,omitempty"`
	// Workers is used to create worker nodes using the Kubermatic machine-controller
	Workers []WorkerConfig `json:"workers,omitempty"`
	// MachineController configures the Kubermatic machine-controller component
	MachineController *MachineControllerConfig `json:"machineController,omitempty"`
	// Features enables and configures additional cluster features
	Features Features `json:"features,omitempty"`
	// Addons are used to deploy additional manifests
	Addons *Addons `json:"addons,omitempty"`
	// SystemPackages configure kubeone behaviour regarding OS packages
	SystemPackages *SystemPackages `json:"systemPackages,omitempty"`
	// Credentials used for machine-controller and external CCM
	Credentials map[string]string `json:"credentials,omitempty"`
}

// HostConfig describes a single control plane node.
type HostConfig struct {
	ID                int    `json:"-"`
	PublicAddress     string `json:"publicAddress"`
	PrivateAddress    string `json:"privateAddress"`
	SSHPort           int    `json:"sshPort"`
	SSHUsername       string `json:"sshUsername"`
	SSHPrivateKeyFile string `json:"sshPrivateKeyFile"`
	SSHAgentSocket    string `json:"sshAgentSocket"`
	Bastion           string `json:"bastion"`
	BastionPort       int    `json:"bastionPort"`
	BastionUser       string `json:"bastionUser"`
	Hostname          string `json:"hostname"`
	IsLeader          bool   `json:"isLeader"`
	Untaint           bool   `json:"untaint"`

	// Information populated at the runtime
	OperatingSystem string `json:"-"`
}

// APIEndpoint is the endpoint used to communicate with the Kubernetes API
type APIEndpoint struct {
	// Host is the hostname on which API is running
	Host string `json:"host"`

	// Port is the port used to reach to the API
	Port int `json:"port"`
}

// CloudProviderName represents the name of a provider
type CloudProviderName string

// CloudProviderName values
const (
	CloudProviderNameAWS          CloudProviderName = "aws"
	CloudProviderNameAzure        CloudProviderName = "azure"
	CloudProviderNameOpenStack    CloudProviderName = "openstack"
	CloudProviderNameHetzner      CloudProviderName = "hetzner"
	CloudProviderNameDigitalOcean CloudProviderName = "digitalocean"
	CloudProviderNamePacket       CloudProviderName = "packet"
	CloudProviderNameVSphere      CloudProviderName = "vsphere"
	CloudProviderNameGCE          CloudProviderName = "gce"
	CloudProviderNameNone         CloudProviderName = "none"
)

// CloudProviderSpec describes the cloud provider that is running the machines
type CloudProviderSpec struct {
	Name        CloudProviderName `json:"name"`
	External    bool              `json:"external"`
	CloudConfig string            `json:"cloudConfig"`
}

// VersionConfig describes the versions of components that are installed on the machines
type VersionConfig struct {
	// TODO(xmudrii): switch to semver
	Kubernetes string `json:"kubernetes"`
}

// ClusterNetworkConfig describes the cluster network
type ClusterNetworkConfig struct {
	PodSubnet         string `json:"podSubnet"`
	ServiceSubnet     string `json:"serviceSubnet"`
	ServiceDomainName string `json:"serviceDomainName"`
	NodePortRange     string `json:"nodePortRange"`
	CNI               *CNI   `json:"cni,omitempty"`

	// +optional
	NetworkID string `json:"networkID,omitempty"`
}

// CNIProvider type
type CNIProvider string

// List of CNI Providers
const (
	// CNIProviderCanal is a Canal CNI plugin (Flannel + Calico).
	// Highlights:
	// * Support Network Policies
	// * Does not support traffic encryption
	// More info: https://docs.projectcalico.org/v3.7/getting-started/kubernetes/installation/flannel
	CNIProviderCanal CNIProvider = "canal"

	// CNIProviderWeaveNet is a WeaveNet CNI plugin.
	// Highlights:
	// * Support Network Policies
	// * Support optional traffic encryption
	// * In case when encryption is enabled, strong secret will be autogenerated
	// More info: https://www.weave.works/docs/net/latest/kubernetes/kube-addon/
	CNIProviderWeaveNet CNIProvider = "weave-net"

	// CNIProviderExternal is an external CNI plugin.
	// The CNI plugin can be installed as Addon or manually
	// More info: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/#pod-network
	CNIProviderExternal CNIProvider = "external"
)

// CNI config
type CNI struct {
	// Provider choice
	Provider CNIProvider `json:"provider"`
	// Encrypted enables encryption for supported CNI plugins
	Encrypted bool `json:"encrypted"`
}

// ProxyConfig configures proxy for the Docker daemon and is used by KubeOne scripts
type ProxyConfig struct {
	HTTP    string `json:"http"`
	HTTPS   string `json:"https"`
	NoProxy string `json:"noProxy"`
}

// WorkerConfig describes a set of worker machines
type WorkerConfig struct {
	Name     string       `json:"name"`
	Replicas *int         `json:"replicas"`
	Config   ProviderSpec `json:"providerSpec"`
}

// ProviderSpec describes a worker node
type ProviderSpec struct {
	CloudProviderSpec   json.RawMessage   `json:"cloudProviderSpec"`
	Labels              map[string]string `json:"labels"`
	Taints              []corev1.Taint    `json:"taints,omitempty"`
	SSHPublicKeys       []string          `json:"sshPublicKeys"`
	OperatingSystem     string            `json:"operatingSystem"`
	OperatingSystemSpec json.RawMessage   `json:"operatingSystemSpec"`

	// +optional
	Network *NetworkConfig `json:"network,omitempty"`

	// +optional
	OverwriteCloudConfig *string `json:"overwriteCloudConfig,omitempty"`
}

// DNSConfig contains a machine's DNS configuration
type DNSConfig struct {
	Servers []string `json:"servers"`
}

// NetworkConfig contains a machine's static network configuration
type NetworkConfig struct {
	CIDR    string    `json:"cidr"`
	Gateway string    `json:"gateway"`
	DNS     DNSConfig `json:"dns"`
}

// MachineControllerConfig configures kubermatic machine-controller deployment
type MachineControllerConfig struct {
	Deploy bool `json:"deploy"`
	// Provider is provider to be used for machine-controller
	// Defaults and must be same as chosen cloud provider, unless cloud provider is set to None
	Provider CloudProviderName `json:"provider"`
}

// Features controls what features will be enabled on the cluster
type Features struct {
	PodPresets        *PodPresets        `json:"podPresets"`
	PodSecurityPolicy *PodSecurityPolicy `json:"podSecurityPolicy"`
	StaticAuditLog    *StaticAuditLog    `json:"staticAuditLog"`
	DynamicAuditLog   *DynamicAuditLog   `json:"dynamicAuditLog"`
	MetricsServer     *MetricsServer     `json:"metricsServer"`
	OpenIDConnect     *OpenIDConnect     `json:"openidConnect"`
}

// SystemPackages controls configurations of APT/YUM
type SystemPackages struct {
	// ConfigureRepositories (true by default) is a flag to control automatic
	// configuration of kubeadm / docker repositories.
	ConfigureRepositories bool `json:"configureRepositories"`
}

// PodPresets feature flag
type PodPresets struct {
	Enable bool `json:"enable"`
}

// PodSecurityPolicy feature flag
type PodSecurityPolicy struct {
	Enable bool `json:"enable"`
}

// StaticAuditLog feature flag
type StaticAuditLog struct {
	Enable bool                 `json:"enable"`
	Config StaticAuditLogConfig `json:"config"`
}

// StaticAuditLogConfig config
type StaticAuditLogConfig struct {
	// PolicyFilePath is a path on local file system to the audit policy manifest
	// which defines what events should be recorded and what data they should include.
	// PolicyFilePath is a required field.
	// More info: https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#audit-policy
	PolicyFilePath string `json:"policyFilePath"`
	// LogPath is path on control plane instances where audit log files are stored.
	// Default value is /var/log/kubernetes/audit.log
	LogPath string `json:"logPath"`
	// LogMaxAge is maximum number of days to retain old audit log files.
	// Default value is 30
	LogMaxAge int `json:"logMaxAge"`
	// LogMaxBackup is maximum number of audit log files to retain.
	// Default value is 3
	LogMaxBackup int `json:"logMaxBackup"`
	// LogMaxSize is maximum size in megabytes of audit log file before it gets rotated.
	// Default value is 100
	LogMaxSize int `json:"logMaxSize"`
}

// DynamicAuditLog feature flag
type DynamicAuditLog struct {
	Enable bool `json:"enable"`
}

// MetricsServer feature flag
type MetricsServer struct {
	Enable bool `json:"enable"`
}

// OpenIDConnect feature flag
type OpenIDConnect struct {
	Enable bool                `json:"enable"`
	Config OpenIDConnectConfig `json:"config"`
}

// OpenIDConnectConfig config
type OpenIDConnectConfig struct {
	IssuerURL      string `json:"issuerUrl"`
	ClientID       string `json:"clientId"`
	UsernameClaim  string `json:"usernameClaim"`
	UsernamePrefix string `json:"usernamePrefix"`
	GroupsClaim    string `json:"groupsClaim"`
	GroupsPrefix   string `json:"groupsPrefix"`
	RequiredClaim  string `json:"requiredClaim"`
	SigningAlgs    string `json:"signingAlgs"`
	CAFile         string `json:"caFile"`
}

// Addons config
type Addons struct {
	Enable bool `json:"enable"`
	// Path on the local file system to the directory with addons manifests.
	Path string `json:"path"`
}
