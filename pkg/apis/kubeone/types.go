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

	// Name is the name of the cluster.
	Name string `json:"name"`

	// ControlPlane describes the control plane nodes and how to access them.
	ControlPlane ControlPlaneConfig `json:"controlPlane"`

	// APIEndpoint are pairs of address and port used to communicate with the Kubernetes API.
	APIEndpoint APIEndpoint `json:"apiEndpoint"`

	// CloudProvider configures the cloud provider specific features.
	CloudProvider CloudProviderSpec `json:"cloudProvider"`

	// Versions defines which Kubernetes version will be installed.
	Versions VersionConfig `json:"versions"`

	// ContainerRuntime defines which container runtime will be installed
	ContainerRuntime ContainerRuntimeConfig `json:"containerRuntime,omitempty"`

	// ClusterNetwork configures the in-cluster networking.
	ClusterNetwork ClusterNetworkConfig `json:"clusterNetwork,omitempty"`

	// Proxy configures proxy used while installing Kubernetes and by the Docker daemon.
	Proxy ProxyConfig `json:"proxy,omitempty"`

	// StaticWorkers describes the worker nodes that are managed by KubeOne/kubeadm.
	StaticWorkers StaticWorkersConfig `json:"staticWorkers,omitempty"`

	// DynamicWorkers describes the worker nodes that are managed by Kubermatic machine-controller/Cluster-API.
	DynamicWorkers []DynamicWorkerConfig `json:"dynamicWorkers,omitempty"`

	// MachineController configures the Kubermatic machine-controller component.
	MachineController *MachineControllerConfig `json:"machineController,omitempty"`

	// CABundle PEM encoded global CA
	CABundle string `json:"caBundle,omitempty"`

	// Features enables and configures additional cluster features.
	Features Features `json:"features,omitempty"`

	// Addons are used to deploy additional manifests.
	Addons *Addons `json:"addons,omitempty"`

	// SystemPackages configure kubeone behaviour regarding OS packages.
	SystemPackages *SystemPackages `json:"systemPackages,omitempty"`

	// AssetConfiguration configures how are binaries and container images downloaded
	AssetConfiguration AssetConfiguration `json:"assetConfiguration,omitempty"`

	// RegistryConfiguration configures how Docker images are pulled from an image registry
	RegistryConfiguration *RegistryConfiguration `json:"registryConfiguration,omitempty"`

	// LoggingConfig configures the Kubelet's log rotation
	LoggingConfig LoggingConfig `json:"loggingConfig,omitempty"`
}

// LoggingConfig configures the Kubelet's log rotation
type LoggingConfig struct {
	// ContainerLogMaxSize configures the maximum size of container log file before it is rotated
	// See more at: https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/
	ContainerLogMaxSize string `json:"containerLogMaxSize,omitempty"`

	// ContainerLogMaxFiles configures the maximum number of container log files that can be present for a container
	// See more at: https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/
	ContainerLogMaxFiles int32 `json:"containerLogMaxFiles,omitempty"`
}

// ContainerRuntimeConfig
type ContainerRuntimeConfig struct {
	// Dockerd related configurations
	Docker *ContainerRuntimeDocker `json:"docker,omitempty"`

	// Containerd related configurations
	Containerd *ContainerRuntimeContainerd `json:"containerd,omitempty"`
}

// ContainerRuntimeDocker defines docker container runtime
type ContainerRuntimeDocker struct {
	// Configures dockerd with "registry-mirrors"
	RegistryMirrors []string `json:"registryMirrors"`
}

// ContainerRuntimeContainerd defines docker container runtime
type ContainerRuntimeContainerd struct {
	// A map of registries to use to render configs and mirrors for containerd registries
	Registries map[string]ContainerdRegistry `json:"registries,omitempty"`
}

// ContainerdRegistry defines endpoints and security for given container registry
type ContainerdRegistry struct {
	// List of registry mirrors to use
	Mirrors []string `json:"mirrors,omitempty"`

	// TLSConfig for the registry
	TLSConfig *ContainerdTLSConfig `json:"tlsConfig,omitempty"`

	// Registry authentication
	Auth *ContainerdRegistryAuthConfig `json:"auth,omitempty"`
}

// Containerd per-registry credentials config
type ContainerdRegistryAuthConfig struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Auth          string `json:"auth,omitempty"`
	IdentityToken string `json:"identityToken,omitempty"`
}

// Configures containerd TLS for a registry
type ContainerdTLSConfig struct {
	// Don't validate remote TLS certificate
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// OperatingSystemName defines the operating system used on instances
type OperatingSystemName string

const (
	OperatingSystemNameUbuntu     OperatingSystemName = "ubuntu"
	OperatingSystemNameDebian     OperatingSystemName = "debian"
	OperatingSystemNameCentOS     OperatingSystemName = "centos"
	OperatingSystemNameRHEL       OperatingSystemName = "rhel"
	OperatingSystemNameRockyLinux OperatingSystemName = "rockylinux"
	OperatingSystemNameAmazon     OperatingSystemName = "amzn"
	OperatingSystemNameFlatcar    OperatingSystemName = "flatcar"
	OperatingSystemNameUnknown    OperatingSystemName = ""
)

// HostConfig describes a single control plane node.
type HostConfig struct {
	// ID automatically assigned at runtime.
	ID int `json:"-"`

	// PublicAddress is externally accessible IP address from public internet.
	PublicAddress string `json:"publicAddress"`

	// PrivateAddress is internal RFC-1918 IP address.
	PrivateAddress string `json:"privateAddress"`

	// SSHPort is port to connect ssh to.
	// Default value is 22.
	SSHPort int `json:"sshPort,omitempty"`

	// SSHUsername is system login name.
	// Default value is "root".
	SSHUsername string `json:"sshUsername,omitempty"`

	// SSHPrivateKeyFile is path to the file with PRIVATE AND CLEANTEXT ssh key.
	// Default value is "".
	SSHPrivateKeyFile string `json:"sshPrivateKeyFile,omitempty"`

	// SSHAgentSocket path (or reference to the environment) to the SSH agent unix domain socket.
	// Default value is "env:SSH_AUTH_SOCK".
	SSHAgentSocket string `json:"sshAgentSocket,omitempty"`

	// Bastion is an IP or hostname of the bastion (or jump) host to connect to.
	// Default value is "".
	Bastion string `json:"bastion,omitempty"`

	// BastionPort is SSH port to use when connecting to the bastion if it's configured in .Bastion.
	// Default value is 22.
	BastionPort int `json:"bastionPort,omitempty"`

	// BastionUser is system login name to use when connecting to bastion host.
	// Default value is "root".
	BastionUser string `json:"bastionUser,omitempty"`

	// Hostname is the hostname(1) of the host.
	// Default value is populated at the runtime via running `hostname -f` command over ssh.
	Hostname string `json:"hostname,omitempty"`

	// IsLeader indicates this host as a session leader.
	// Default value is populated at the runtime.
	IsLeader bool `json:"isLeader,omitempty"`

	// Taints are taints applied to nodes. Those taints are only applied when the node is being provisioned.
	// If not provided (i.e. nil) for control plane nodes, it defaults to:
	//   * For Kubernetes 1.23 and older: TaintEffectNoSchedule with key node-role.kubernetes.io/master
	//   * For Kubernetes 1.24 and newer: TaintEffectNoSchedule with keys
	//     node-role.kubernetes.io/control-plane and node-role.kubernetes.io/master
	// Explicitly empty (i.e. []corev1.Taint{}) means no taints will be applied (this is default for worker nodes).
	Taints []corev1.Taint `json:"taints,omitempty"`

	// Lables to be used to apply (or remove, with minus symbol suffix, see more kubectl help label) lables to/from node
	Labels map[string]string `json:"labels,omitempty"`

	// Kubelet
	Kubelet KubeletConfig `json:"kubelet,omitempty"`

	// OperatingSystem information, can be populated at the runtime.
	OperatingSystem OperatingSystemName `json:"operatingSystem,omitempty"`
}

// ControlPlaneConfig defines control plane nodes
type ControlPlaneConfig struct {
	// Hosts array of all control plane hosts.
	Hosts []HostConfig `json:"hosts"`
}

// StaticWorkersConfig defines static worker nodes provisioned by KubeOne and kubeadm
type StaticWorkersConfig struct {
	// Hosts
	Hosts []HostConfig `json:"hosts,omitempty"`
}

// KubeletConfig provides some kubelet configuration options
type KubeletConfig struct {
	// SystemReserved configure --system-reserved command-line flag of the kubelet.
	// See more at: https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/
	SystemReserved map[string]string `json:"systemReserved,omitempty"`

	// KubeReserved configure --kube-reserved command-line flag of the kubelet.
	// See more at: https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/
	KubeReserved map[string]string `json:"kubeReserved,omitempty"`

	// EvictionHard configure --eviction-hard command-line flag of the kubelet.
	// See more at: https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/
	EvictionHard map[string]string `json:"evictionHard,omitempty"`

	// MaxPods configures maximum number of pods per node.
	// If not provided, default value provided by kubelet will be used
	// (max. 110 pods per node)
	MaxPods *int32 `json:"maxPods,omitempty"`
}

// APIEndpoint is the endpoint used to communicate with the Kubernetes API
type APIEndpoint struct {
	// Host is the hostname or IP on which API is running.
	Host string `json:"host"`

	// Port is the port used to reach to the API.
	// Default value is 6443.
	Port int `json:"port,omitempty"`

	// AlternativeNames is a list of Subject Alternative Names for the API Server signing cert.
	AlternativeNames []string `json:"alternativeNames,omitempty"`
}

// CloudProviderSpec describes the cloud provider that is running the machines.
// Only one cloud provider must be defined at the single time.
type CloudProviderSpec struct {
	// External
	External bool `json:"external,omitempty"`

	// CloudConfig
	CloudConfig string `json:"cloudConfig,omitempty"`

	// CSIConfig
	CSIConfig string `json:"csiConfig,omitempty"`

	// AWS
	AWS *AWSSpec `json:"aws,omitempty"`

	// Azure
	Azure *AzureSpec `json:"azure,omitempty"`

	// DigitalOcean
	DigitalOcean *DigitalOceanSpec `json:"digitalocean,omitempty"`

	// GCE
	GCE *GCESpec `json:"gce,omitempty"`

	// Hetzner
	Hetzner *HetznerSpec `json:"hetzner,omitempty"`

	// Nutanix
	Nutanix *NutanixSpec `json:"nutanix,omitempty"`

	// Openstack
	Openstack *OpenstackSpec `json:"openstack,omitempty"`

	// EquinixMetal
	EquinixMetal *EquinixMetalSpec `json:"equinixmetal,omitempty"`

	// VMware Cloud Director
	VMwareCloudDirector *VMwareCloudDirectorSpec `json:"vmwareCloudDirector,omitempty"`

	// Vsphere
	Vsphere *VsphereSpec `json:"vsphere,omitempty"`

	// None
	None *NoneSpec `json:"none,omitempty"`
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
	// NetworkID
	NetworkID string `json:"networkID,omitempty"`
}

// NutanixSpec defines the Nutanix provider
type NutanixSpec struct{}

// OpenstackSpec defines the Openstack provider
type OpenstackSpec struct{}

// EquinixMetalSpec defines the Equinix Metal cloud provider
type EquinixMetalSpec struct{}

// VMwareCloudDirectorSpec defines the VMware Cloud Director provider
type VMwareCloudDirectorSpec struct {
	// VApp is the name of vApp for VMs.
	VApp string `json:"vApp,omitempty"`

	// StorageProfile is the name of storage profile to be used for disks.
	StorageProfile string `json:"storageProfile"`
}

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
	// PodSubnet
	// default value is "10.244.0.0/16"
	PodSubnet string `json:"podSubnet,omitempty"`

	// ServiceSubnet
	// default value is "10.96.0.0/12"
	ServiceSubnet string `json:"serviceSubnet,omitempty"`

	// ServiceDomainName
	// default value is "cluster.local"
	ServiceDomainName string `json:"serviceDomainName,omitempty"`

	// NodePortRange
	// default value is "30000-32767"
	NodePortRange string `json:"nodePortRange,omitempty"`

	// CNI
	// default value is {canal: {mtu: 1450}}
	CNI *CNI `json:"cni,omitempty"`

	// KubeProxy config
	KubeProxy *KubeProxyConfig `json:"kubeProxy,omitempty"`
}

// KubeProxyConfig defines configured kube-proxy mode, default is iptables mode
type KubeProxyConfig struct {
	// SkipInstallation will skip the installation of kube-proxy
	// default value is false
	SkipInstallation bool `json:"skipInstallation"`

	// IPVS config
	IPVS *IPVSConfig `json:"ipvs"`

	// IPTables config
	IPTables *IPTables `json:"iptables"`
}

// IPVSConfig contains different options to configure IPVS kube-proxy mode
type IPVSConfig struct {
	// ipvs scheduler, if it’s not configured, then round-robin (rr) is the default value.
	// Can be one of:
	// * rr: round-robin
	// * lc: least connection (smallest number of open connections)
	// * dh: destination hashing
	// * sh: source hashing
	// * sed: shortest expected delay
	// * nq: never queue
	Scheduler string `json:"scheduler"`

	// excludeCIDRs is a list of CIDR's which the ipvs proxier should not touch
	// when cleaning up ipvs services.
	ExcludeCIDRs []string `json:"excludeCIDRs"`

	// strict ARP configure arp_ignore and arp_announce to avoid answering ARP queries
	// from kube-ipvs0 interface
	StrictARP bool `json:"strictARP"`

	// tcpTimeout is the timeout value used for idle IPVS TCP sessions.
	// The default value is 0, which preserves the current timeout value on the system.
	TCPTimeout metav1.Duration `json:"tcpTimeout"`

	// tcpFinTimeout is the timeout value used for IPVS TCP sessions after receiving a FIN.
	// The default value is 0, which preserves the current timeout value on the system.
	TCPFinTimeout metav1.Duration `json:"tcpFinTimeout"`

	// udpTimeout is the timeout value used for IPVS UDP packets.
	// The default value is 0, which preserves the current timeout value on the system.
	UDPTimeout metav1.Duration `json:"udpTimeout"`
}

// IPTables
type IPTables struct{}

// CNI config. Only one CNI provider must be used at the single time.
type CNI struct {
	// Canal
	Canal *CanalSpec `json:"canal,omitempty"`

	// Cilium
	Cilium *CiliumSpec `json:"cilium,omitempty"`

	// WeaveNet
	WeaveNet *WeaveNetSpec `json:"weaveNet,omitempty"`

	// External
	External *ExternalCNISpec `json:"external,omitempty"`
}

// CanalSpec defines the Canal CNI plugin
type CanalSpec struct {
	// MTU automatically detected based on the cloudProvider
	// default value is 1450
	MTU int `json:"mtu,omitempty"`
}

type KubeProxyReplacementType string

const (
	KubeProxyReplacementStrict   KubeProxyReplacementType = "strict"
	KubeProxyReplacementDisabled KubeProxyReplacementType = "disabled"
)

// CiliumSpec defines the Cilium CNI plugin
type CiliumSpec struct {
	// KubeProxyReplacement defines weather cilium relies on underlying Kernel support
	// to replace kube-proxy functionality by eBPF (strict), or disables a subset of those
	// features so cilium does not bail out if the kernel support is missing (disabled).
	// default is "disabled"
	KubeProxyReplacement KubeProxyReplacementType `json:"kubeProxyReplacement"`

	// EnableHubble to deploy Hubble relay and UI
	// default value is false
	EnableHubble bool `json:"enableHubble"`
}

// WeaveNetSpec defines the WeaveNet CNI plugin
type WeaveNetSpec struct {
	// Encrypted
	Encrypted bool `json:"encrypted,omitempty"`
}

// ExternalCNISpec defines the external CNI plugin.
// It's up to the user's responsibility to deploy the external CNI plugin manually or as an addon
type ExternalCNISpec struct{}

// ProxyConfig configures proxy for the Docker daemon and is used by KubeOne scripts
type ProxyConfig struct {
	// HTTP
	HTTP string `json:"http,omitempty"`

	// HTTPS
	HTTPS string `json:"https,omitempty"`

	// NoProxy
	NoProxy string `json:"noProxy,omitempty"`
}

// DynamicWorkerConfig describes a set of worker machines
type DynamicWorkerConfig struct {
	// Name
	Name string `json:"name"`

	// Replicas
	Replicas *int `json:"replicas"`

	// Config
	Config ProviderSpec `json:"providerSpec"`
}

// ProviderSpec describes a worker node
type ProviderSpec struct {
	// CloudProviderSpec
	CloudProviderSpec json.RawMessage `json:"cloudProviderSpec"`

	// Annotations set MachineDeployment.ObjectMeta.Annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// MachineAnnotations set MachineDeployment.Spec.Template.Spec.ObjectMeta.Annotations
	// as a way to annotate resulting Nodes
	// Deprecated: Use NodeAnnotations instead.
	MachineAnnotations map[string]string `json:"machineAnnotations,omitempty"`

	// NodeAnnotations set MachineDeployment.Spec.Template.Spec.ObjectMeta.Annotations
	// as a way to annotate resulting Nodes
	NodeAnnotations map[string]string `json:"nodeAnnotations,omitempty"`

	// MachineObjectAnnotations set MachineDeployment.Spec.Template.Metadata.Annotations
	// as a way to annotate resulting Machine objects. Those annotations are not
	// propagated to Node objects. If you want to annotate resulting Nodes as well,
	// see NodeAnnotations
	MachineObjectAnnotations map[string]string `json:"machineObjectAnnotations,omitempty"`

	// Labels
	Labels map[string]string `json:"labels,omitempty"`

	// Taints
	Taints []corev1.Taint `json:"taints,omitempty"`

	// SSHPublicKeys
	SSHPublicKeys []string `json:"sshPublicKeys,omitempty"`

	// OperatingSystem
	OperatingSystem string `json:"operatingSystem"`

	// OperatingSystemSpec
	OperatingSystemSpec json.RawMessage `json:"operatingSystemSpec,omitempty"`

	// Network
	Network *ProviderStaticNetworkConfig `json:"network,omitempty"`

	// OverwriteCloudConfig
	OverwriteCloudConfig *string `json:"overwriteCloudConfig,omitempty"`
}

// DNSConfig contains a machine's DNS configuration
type DNSConfig struct {
	// Servers
	Servers []string `json:"servers"`
}

// ProviderStaticNetworkConfig contains a machine's static network configuration
type ProviderStaticNetworkConfig struct {
	// CIDR
	CIDR string `json:"cidr"`

	// Gateway
	Gateway string `json:"gateway"`

	// DNS
	DNS DNSConfig `json:"dns"`
}

// MachineControllerConfig configures kubermatic machine-controller deployment
type MachineControllerConfig struct {
	// Deploy
	Deploy bool `json:"deploy,omitempty"`
}

// Features controls what features will be enabled on the cluster
type Features struct {
	// PodNodeSelector
	PodNodeSelector *PodNodeSelector `json:"podNodeSelector,omitempty"`

	// PodSecurityPolicy
	// Deprecated: will be removed once Kubernetes 1.24 reaches EOL
	PodSecurityPolicy *PodSecurityPolicy `json:"podSecurityPolicy,omitempty"`

	// StaticAuditLog
	StaticAuditLog *StaticAuditLog `json:"staticAuditLog,omitempty"`

	// DynamicAuditLog
	DynamicAuditLog *DynamicAuditLog `json:"dynamicAuditLog,omitempty"`

	// MetricsServer
	MetricsServer *MetricsServer `json:"metricsServer,omitempty"`

	// OpenIDConnect
	OpenIDConnect *OpenIDConnect `json:"openidConnect,omitempty"`

	// Encryption Providers
	EncryptionProviders *EncryptionProviders `json:"encryptionProviders,omitempty"`
}

// SystemPackages controls configurations of APT/YUM
type SystemPackages struct {
	// ConfigureRepositories (true by default) is a flag to control automatic
	// configuration of kubeadm / docker repositories.
	ConfigureRepositories bool `json:"configureRepositories,omitempty"`
}

// AssetConfiguration controls how assets (e.g. CNI, Kubelet, kube-apiserver, and more)
// are pulled.
// The AssetConfiguration API is a deprecated API removed in the v1beta2 API.
// The AssetConfiguration API will be completely removed in KubeOne 1.6+
// Currently, configuring BinaryAssets works only on Amazon Linux 2.
type AssetConfiguration struct {
	// Kubernetes configures the image registry and repository for the core Kubernetes
	// images (kube-apiserver, kube-controller-manager, kube-scheduler, and kube-proxy).
	// Kubernetes respects only ImageRepository (ImageTag is ignored).
	// Default image repository and tag: defaulted dynamically by Kubeadm.
	// Defaults to RegistryConfiguration.OverwriteRegistry if left empty
	// and RegistryConfiguration.OverwriteRegistry is specified.
	Kubernetes ImageAsset `json:"kubernetes,omitempty"`

	// Pause configures the sandbox (pause) image to be used by Kubelet.
	// Default image repository and tag: defaulted dynamically by Kubeadm.
	// Defaults to RegistryConfiguration.OverwriteRegistry if left empty
	// and RegistryConfiguration.OverwriteRegistry is specified.
	Pause ImageAsset `json:"pause,omitempty"`

	// CoreDNS configures the image registry and tag to be used for deploying
	// the CoreDNS component.
	// Default image repository and tag: defaulted dynamically by Kubeadm.
	// Defaults to RegistryConfiguration.OverwriteRegistry if left empty
	// and RegistryConfiguration.OverwriteRegistry is specified.
	CoreDNS ImageAsset `json:"coreDNS,omitempty"`

	// Etcd configures the image registry and tag to be used for deploying
	// the Etcd component.
	// Default image repository and tag: defaulted dynamically by Kubeadm.
	// Defaults to RegistryConfiguration.OverwriteRegistry if left empty
	// and RegistryConfiguration.OverwriteRegistry is specified.
	Etcd ImageAsset `json:"etcd,omitempty"`

	// MetricsServer configures the image registry and tag to be used for deploying
	// the metrics-server component.
	// Default image repository and tag: defaulted dynamically by KubeOne.
	// Defaults to RegistryConfiguration.OverwriteRegistry if left empty
	// and RegistryConfiguration.OverwriteRegistry is specified.
	MetricsServer ImageAsset `json:"metricsServer,omitempty"`

	// CNI configures the source for downloading the CNI binaries.
	// If not specified, kubernetes-cni package will be installed.
	// Default: none
	CNI BinaryAsset `json:"cni,omitempty"`

	// NodeBinaries configures the source for downloading the
	// Kubernetes Node Binaries tarball (e.g. kubernetes-node-linux-amd64.tar.gz).
	// The tarball must have .tar.gz as the extension and must contain the
	// following files:
	// - kubernetes/node/bin/kubelet
	// - kubernetes/node/bin/kubeadm
	// If not specified, kubelet and kubeadm packages will be installed.
	// Default: none
	NodeBinaries BinaryAsset `json:"nodeBinaries,omitempty"`

	// Kubectl configures the source for downloading the Kubectl binary.
	// If not specified, kubelet package will be installed.
	// Default: none
	Kubectl BinaryAsset `json:"kubectl,omitempty"`
}

// ImageAsset is used to customize the image repository and the image tag
type ImageAsset struct {
	// ImageRepository customizes the registry/repository
	ImageRepository string `json:"imageRepository,omitempty"`

	// ImageTag customizes the image tag
	ImageTag string `json:"imageTag,omitempty"`
}

// BinaryAsset is used to customize the URL of the binary asset
type BinaryAsset struct {
	// URL from where to download the binary
	URL string `json:"url,omitempty"`
}

// RegistryConfiguration controls how images used for components deployed by
// KubeOne and kubeadm are pulled from an image registry
type RegistryConfiguration struct {
	// OverwriteRegistry specifies a custom Docker registry which will be used
	// for all images required for KubeOne and kubeadm. This also applies to
	// addons deployed by KubeOne.
	// This field doesn't modify the user/organization part of the image. For example,
	// if OverwriteRegistry is set to 127.0.0.1:5000/example, image called
	// calico/cni would translate to 127.0.0.1:5000/example/calico/cni.
	// Default: ""
	OverwriteRegistry string `json:"overwriteRegistry,omitempty"`

	// InsecureRegistry configures Docker to threat the registry specified
	// in OverwriteRegistry as an insecure registry. This is also propagated
	// to the worker nodes managed by machine-controller and/or KubeOne.
	InsecureRegistry bool `json:"insecureRegistry,omitempty"`
}

// PodNodeSelector feature flag
type PodNodeSelector struct {
	// Enable
	Enable bool `json:"enable,omitempty"`

	// Config
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

// PodSecurityPolicy feature flag
// This feature is deprecated and will be removed from the API once
// Kubernetes 1.24 reaches EOL.
type PodSecurityPolicy struct {
	// Enable
	Enable bool `json:"enable,omitempty"`
}

// StaticAuditLog feature flag
type StaticAuditLog struct {
	// Enable
	Enable bool `json:"enable,omitempty"`

	// Config
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
	LogPath string `json:"logPath,omitempty"`

	// LogMaxAge is maximum number of days to retain old audit log files.
	// Default value is 30
	LogMaxAge int `json:"logMaxAge,omitempty"`

	// LogMaxBackup is maximum number of audit log files to retain.
	// Default value is 3.
	LogMaxBackup int `json:"logMaxBackup,omitempty"`

	// LogMaxSize is maximum size in megabytes of audit log file before it gets rotated.
	// Default value is 100.
	LogMaxSize int `json:"logMaxSize,omitempty"`
}

// DynamicAuditLog feature flag
type DynamicAuditLog struct {
	// Enable
	// Default value is false.
	Enable bool `json:"enable,omitempty"`
}

// MetricsServer feature flag
type MetricsServer struct {
	// Enable deployment of metrics-server.
	// Default value is true.
	Enable bool `json:"enable,omitempty"`
}

// OpenIDConnect feature flag
type OpenIDConnect struct {
	// Enable
	Enable bool `json:"enable,omitempty"`

	// Config
	Config OpenIDConnectConfig `json:"config"`
}

// OpenIDConnectConfig config
type OpenIDConnectConfig struct {
	// IssuerURL
	IssuerURL string `json:"issuerUrl"`

	// ClientID
	ClientID string `json:"clientId,omitempty"`

	// UsernameClaim
	UsernameClaim string `json:"usernameClaim,omitempty"`

	// UsernamePrefix. The value `-` can be used to disable all prefixing.
	UsernamePrefix string `json:"usernamePrefix,omitempty"`

	// GroupsClaim
	GroupsClaim string `json:"groupsClaim,omitempty"`

	// GroupsPrefix. The value `-` can be used to disable all prefixing.
	GroupsPrefix string `json:"groupsPrefix,omitempty"`

	// RequiredClaim
	RequiredClaim string `json:"requiredClaim"`

	// SigningAlgs
	SigningAlgs string `json:"signingAlgs,omitempty"`

	// CAFile
	CAFile string `json:"caFile"`
}

// Addon config
type Addon struct {
	// Name of the addon to configure
	Name string `json:"name"`

	// Params to the addon, to render the addon using text/template, this will override globalParams
	Params map[string]string `json:"params,omitempty"`

	// Delete flag to ensure the named addon with all its contents to be deleted
	Delete bool `json:"delete,omitempty"`
}

// Addons config
type Addons struct {
	// Enable
	Enable bool `json:"enable,omitempty"`

	// Path on the local file system to the directory with addons manifests.
	Path string `json:"path,omitempty"`

	// GlobalParams to the addon, to render all addons using text/template
	GlobalParams map[string]string `json:"globalParams,omitempty"`

	// Addons is a list of config options for named addon
	Addons []Addon `json:"addons,omitempty"`
}

// Encryption Providers feature flag
type EncryptionProviders struct {
	// Enable
	Enable bool `json:"enable"`

	// CustomEncryptionConfiguration
	CustomEncryptionConfiguration string `json:"customEncryptionConfiguration"`
}
