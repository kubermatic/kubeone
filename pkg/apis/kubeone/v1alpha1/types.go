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

package v1alpha1

import (
	"encoding/json"

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
	// Credentials used for machine-controller and external CCM
	Credentials map[string]string `json:"credentials,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubeOneSecrets hold secrets needed to provision the cluster
type KubeOneSecrets struct {
	metav1.TypeMeta `json:",inline"`

	// Secrets contains list of secrets to be used by KubeOne instead of
	// environment variables
	Secrets map[string]string `json:"secrets"`
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
	Hostname          string `json:"hostname"`

	// Information populated at the runtime
	OperatingSystem string `json:"-"`
	IsLeader        bool   `json:"-"`
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
	PodSecurityPolicy *PodSecurityPolicy `json:"podSecurityPolicy"`
	DynamicAuditLog   *DynamicAuditLog   `json:"dynamicAuditLog"`
	MetricsServer     *MetricsServer     `json:"metricsServer"`
	OpenIDConnect     *OpenIDConnect     `json:"openidConnect"`
}

// PodSecurityPolicy feature flag
type PodSecurityPolicy struct {
	Enable bool `json:"enable"`
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
