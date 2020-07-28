/*
Copyright 2020 The KubeOne Authors.

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

package v1beta1

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
	// ControlPlane describes the control plane nodes and how to access them
	ControlPlane ControlPlaneConfig `json:"controlPlane,omitempty"`
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
	// StaticWorkers describes the worker nodes that are managed by KubeOne/kubeadm
	StaticWorkers StaticWorkersConfig `json:"staticWorkers,omitempty"`
	// DynamicWorkers describes the worker nodes that are managed by
	// Kubermatic machine-controller/Cluster-API
	DynamicWorkers []DynamicWorkerConfig `json:"dynamicWorkers,omitempty"`
	// MachineController configures the Kubermatic machine-controller component
	MachineController *MachineControllerConfig `json:"machineController,omitempty"`
	// Features enables and configures additional cluster features
	Features Features `json:"features,omitempty"`
	// Addons are used to deploy additional manifests
	Addons *Addons `json:"addons,omitempty"`
	// SystemPackages configure kubeone behaviour regarding OS packages
	SystemPackages *SystemPackages `json:"systemPackages,omitempty"`
}

// OperatingSystemName defines the operating system used on instances
type OperatingSystemName string

var (
	OperatingSystemNameUbuntu  OperatingSystemName = "ubuntu"
	OperatingSystemNameCentOS  OperatingSystemName = "centos"
	OperatingSystemNameRHEL    OperatingSystemName = "rhel"
	OperatingSystemNameCoreOS  OperatingSystemName = "coreos"
	OperatingSystemNameFlatcar OperatingSystemName = "flatcar"
	OperatingSystemNameUnknown OperatingSystemName = ""
)

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

	// If not provided (i.e. nil) defaults to TaintEffectNoSchedule, with key
	// node-role.kubernetes.io/master for control plane nodes.
	//
	// Explicitly empty (i.e. []corev1.Taint{}) means no taints will be applied
	// (this is default for worker nodes).
	Taints []corev1.Taint `json:"taints,omitempty"`

	// Information populated at the runtime
	OperatingSystem OperatingSystemName `json:"-"`
}

// ControlPlaneConfig defines control plane nodes
type ControlPlaneConfig struct {
	Hosts []HostConfig `json:"hosts"`
}

// StaticWorkersConfig defines static worker nodes provisioned by KubeOne and kubeadm
type StaticWorkersConfig struct {
	Hosts []HostConfig `json:"hosts"`
}

// APIEndpoint is the endpoint used to communicate with the Kubernetes API
type APIEndpoint struct {
	// Host is the hostname on which API is running
	Host string `json:"host"`

	// Port is the port used to reach to the API
	Port int `json:"port"`
}

// CloudProviderSpec describes the cloud provider that is running the machines.
// Only one cloud provider must be defined at the single time.
type CloudProviderSpec struct {
	External     bool              `json:"external"`
	CloudConfig  string            `json:"cloudConfig"`
	AWS          *AWSSpec          `json:"aws"`
	Azure        *AzureSpec        `json:"azure"`
	DigitalOcean *DigitalOceanSpec `json:"digitalocean"`
	GCE          *GCESpec          `json:"gce"`
	Hetzner      *HetznerSpec      `json:"hetzner"`
	Openstack    *OpenstackSpec    `json:"openstack"`
	Packet       *PacketSpec       `json:"packet"`
	Vsphere      *VsphereSpec      `json:"vsphere"`
	None         *NoneSpec         `json:"none"`
}

// AWSSpec defines the AWS cloud provider
type AWSSpec struct{}

// AzureSpec defines the Azure cloud provider
type AzureSpec struct{}

// DigitalOceanSpec defines the DigitalOcean cloud provider
type DigitalOceanSpec struct{}

// GCESpec defines the GCE cloud provider
type GCESpec struct{}

// HetznerSpec defines the Hetzner cloud provider
type HetznerSpec struct {
	NetworkID string `json:"networkID"`
}

// OpenstackSpec defines the Openstack provider
type OpenstackSpec struct{}

// PacketSpec defines the Packet cloud provider
type PacketSpec struct{}

// VsphereSpec defines the vSphere provider
type VsphereSpec struct{}

// NoneSpec defines a none provider
type NoneSpec struct{}

// VersionConfig describes the versions of components that are installed on the machines
type VersionConfig struct {
	Kubernetes string `json:"kubernetes"`
}

// ClusterNetworkConfig describes the cluster network
type ClusterNetworkConfig struct {
	PodSubnet         string `json:"podSubnet"`
	ServiceSubnet     string `json:"serviceSubnet"`
	ServiceDomainName string `json:"serviceDomainName"`
	NodePortRange     string `json:"nodePortRange"`
	CNI               *CNI   `json:"cni,omitempty"`
}

// CNI config. Only one CNI provider must be used at the single time.
type CNI struct {
	Canal    *CanalSpec       `json:"canal"`
	WeaveNet *WeaveNetSpec    `json:"weaveNet"`
	External *ExternalCNISpec `json:"external"`
}

// CanalSpec defines the Canal CNI plugin
type CanalSpec struct {
	MTU string `json:"mtu"`
}

// WeaveNetSpec defines the WeaveNet CNI plugin
type WeaveNetSpec struct {
	Encrypted bool `json:"encrypted"`
}

// ExternalCNISpec defines the external CNI plugin.
// It's up to the user's responsibility to deploy the external CNI plugin manually or as an addon
type ExternalCNISpec struct{}

// ProxyConfig configures proxy for the Docker daemon and is used by KubeOne scripts
type ProxyConfig struct {
	HTTP    string `json:"http"`
	HTTPS   string `json:"https"`
	NoProxy string `json:"noProxy"`
}

// DynamicWorkerConfig describes a set of worker machines
type DynamicWorkerConfig struct {
	Name     string       `json:"name"`
	Replicas *int         `json:"replicas"`
	Config   ProviderSpec `json:"providerSpec"`
}

// ProviderSpec describes a worker node
type ProviderSpec struct {
	CloudProviderSpec   json.RawMessage   `json:"cloudProviderSpec"`
	Labels              map[string]string `json:"labels,omitempty"`
	Taints              []corev1.Taint    `json:"taints,omitempty"`
	SSHPublicKeys       []string          `json:"sshPublicKeys,omitempty"`
	OperatingSystem     string            `json:"operatingSystem"`
	OperatingSystemSpec json.RawMessage   `json:"operatingSystemSpec"`

	// +optional
	Network *ProviderStaticNetworkConfig `json:"network,omitempty"`

	// +optional
	OverwriteCloudConfig *string `json:"overwriteCloudConfig,omitempty"`
}

// DNSConfig contains a machine's DNS configuration
type DNSConfig struct {
	Servers []string `json:"servers"`
}

// ProviderStaticNetworkConfig contains a machine's static network configuration
type ProviderStaticNetworkConfig struct {
	CIDR    string    `json:"cidr"`
	Gateway string    `json:"gateway"`
	DNS     DNSConfig `json:"dns"`
}

// MachineControllerConfig configures kubermatic machine-controller deployment
type MachineControllerConfig struct {
	Deploy bool `json:"deploy"`
}

// Features controls what features will be enabled on the cluster
type Features struct {
	PodNodeSelector   *PodNodeSelector   `json:"podNodeSelector"`
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

// PodNodeSelector feature flag
type PodNodeSelector struct {
	Enable bool                  `json:"enable"`
	Config PodNodeSelectorConfig `json:"config"`
}

// PodNodeSelectorConfig config
type PodNodeSelectorConfig struct {
	// ConfigFilePath is a path on the local file system to the PodNodeSelector
	// configuration file.
	// ConfigFilePath is a required field.
	// More info: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#podnodeselector
	ConfigFilePath string `json:"configFilePath"`
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
