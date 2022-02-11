+++
title = "v1beta2 API Reference"
date = 2022-02-04T17:38:06+02:00
weight = 11
+++
## v1beta2

* [APIEndpoint](#apiendpoint)
* [AWSSpec](#awsspec)
* [Addon](#addon)
* [Addons](#addons)
* [AzureSpec](#azurespec)
* [BinaryAsset](#binaryasset)
* [CNI](#cni)
* [CanalSpec](#canalspec)
* [CiliumSpec](#ciliumspec)
* [CloudProviderSpec](#cloudproviderspec)
* [ClusterNetworkConfig](#clusternetworkconfig)
* [ContainerRuntimeConfig](#containerruntimeconfig)
* [ContainerRuntimeContainerd](#containerruntimecontainerd)
* [ContainerRuntimeDocker](#containerruntimedocker)
* [ContainerdRegistry](#containerdregistry)
* [ContainerdRegistryAuthConfig](#containerdregistryauthconfig)
* [ContainerdTLSConfig](#containerdtlsconfig)
* [ControlPlaneConfig](#controlplaneconfig)
* [DNSConfig](#dnsconfig)
* [DigitalOceanSpec](#digitaloceanspec)
* [DynamicAuditLog](#dynamicauditlog)
* [DynamicWorkerConfig](#dynamicworkerconfig)
* [EncryptionProviders](#encryptionproviders)
* [EquinixMetalSpec](#equinixmetalspec)
* [ExternalCNISpec](#externalcnispec)
* [Features](#features)
* [GCESpec](#gcespec)
* [HetznerSpec](#hetznerspec)
* [HostConfig](#hostconfig)
* [IPTables](#iptables)
* [IPVSConfig](#ipvsconfig)
* [ImageAsset](#imageasset)
* [KubeOneCluster](#kubeonecluster)
* [KubeProxyConfig](#kubeproxyconfig)
* [KubeletConfig](#kubeletconfig)
* [LoggingConfig](#loggingconfig)
* [MachineControllerConfig](#machinecontrollerconfig)
* [MetricsServer](#metricsserver)
* [NoneSpec](#nonespec)
* [NutanixSpec](#nutanixspec)
* [OpenIDConnect](#openidconnect)
* [OpenIDConnectConfig](#openidconnectconfig)
* [OpenstackSpec](#openstackspec)
* [PodNodeSelector](#podnodeselector)
* [PodNodeSelectorConfig](#podnodeselectorconfig)
* [PodSecurityPolicy](#podsecuritypolicy)
* [ProviderSpec](#providerspec)
* [ProviderStaticNetworkConfig](#providerstaticnetworkconfig)
* [ProxyConfig](#proxyconfig)
* [RegistryConfiguration](#registryconfiguration)
* [StaticAuditLog](#staticauditlog)
* [StaticAuditLogConfig](#staticauditlogconfig)
* [StaticWorkersConfig](#staticworkersconfig)
* [SystemPackages](#systempackages)
* [VersionConfig](#versionconfig)
* [VsphereSpec](#vspherespec)
* [WeaveNetSpec](#weavenetspec)

### APIEndpoint

APIEndpoint is the endpoint used to communicate with the Kubernetes API

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| host | Host is the hostname or IP on which API is running. | string | true |
| port | Port is the port used to reach to the API. Default value is 6443. | int | false |
| alternativeNames | AlternativeNames is a list of Subject Alternative Names for the API Server signing cert. | []string | false |

[Back to Group](#v1beta2)

### AWSSpec

AWSSpec defines the AWS cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta2)

### Addon

Addon config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of the addon to configure | string | true |
| params | Params to the addon, to render the addon using text/template, this will override globalParams | map[string]string | false |
| delete | Delete flag to ensure the named addon with all its contents to be deleted | bool | false |

[Back to Group](#v1beta2)

### Addons

Addons config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |
| path | Path on the local file system to the directory with addons manifests. | string | false |
| globalParams | GlobalParams to the addon, to render all addons using text/template | map[string]string | false |
| addons | Addons is a list of config options for named addon | [][Addon](#addon) | false |

[Back to Group](#v1beta2)

### AzureSpec

AzureSpec defines the Azure cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta2)

### BinaryAsset

BinaryAsset is used to customize the URL of the binary asset

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| url | URL from where to download the binary | string | false |

[Back to Group](#v1beta2)

### CNI

CNI config. Only one CNI provider must be used at the single time.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| canal | Canal | *[CanalSpec](#canalspec) | false |
| cilium | Cilium | *[CiliumSpec](#ciliumspec) | false |
| weaveNet | WeaveNet | *[WeaveNetSpec](#weavenetspec) | false |
| external | External | *[ExternalCNISpec](#externalcnispec) | false |

[Back to Group](#v1beta2)

### CanalSpec

CanalSpec defines the Canal CNI plugin

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| mtu | MTU automatically detected based on the cloudProvider default value is 1450 | int | false |

[Back to Group](#v1beta2)

### CiliumSpec

CiliumSpec defines the Cilium CNI plugin

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kubeProxyReplacement | KubeProxyReplacement defines weather cilium relies on underlying Kernel support to replace kube-proxy functionality by eBPF (strict), or disables a subset of those features so cilium does not bail out if the kernel support is missing (disabled). default is \"disabled\" | KubeProxyReplacementType | true |
| enableHubble | EnableHubble to deploy Hubble relay and UI default value is false | bool | true |

[Back to Group](#v1beta2)

### CloudProviderSpec

CloudProviderSpec describes the cloud provider that is running the machines.
Only one cloud provider must be defined at the single time.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| external | External | bool | false |
| cloudConfig | CloudConfig | string | false |
| csiConfig | CSIConfig | string | false |
| aws | AWS | *[AWSSpec](#awsspec) | false |
| azure | Azure | *[AzureSpec](#azurespec) | false |
| digitalocean | DigitalOcean | *[DigitalOceanSpec](#digitaloceanspec) | false |
| gce | GCE | *[GCESpec](#gcespec) | false |
| hetzner | Hetzner | *[HetznerSpec](#hetznerspec) | false |
| nutanix | Nutanix | *[NutanixSpec](#nutanixspec) | false |
| openstack | Openstack | *[OpenstackSpec](#openstackspec) | false |
| equinixmetal | Equinix Metal | *[EquinixMetalSpec](#equinixmetalspec) | false |
| vsphere | Vsphere | *[VsphereSpec](#vspherespec) | false |
| none | None | *[NoneSpec](#nonespec) | false |

[Back to Group](#v1beta2)

### ClusterNetworkConfig

ClusterNetworkConfig describes the cluster network

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podSubnet | PodSubnet default value is \"10.244.0.0/16\" | string | false |
| serviceSubnet | ServiceSubnet default value is \"10.96.0.0/12\" | string | false |
| serviceDomainName | ServiceDomainName default value is \"cluster.local\" | string | false |
| nodePortRange | NodePortRange default value is \"30000-32767\" | string | false |
| cni | CNI default value is {canal: {mtu: 1450}} | *[CNI](#cni) | false |
| kubeProxy | KubeProxy config | *[KubeProxyConfig](#kubeproxyconfig) | false |

[Back to Group](#v1beta2)

### ContainerRuntimeConfig

ContainerRuntimeConfig

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| docker | Dockerd related configurations | *[ContainerRuntimeDocker](#containerruntimedocker) | false |
| containerd | Containerd related configurations | *[ContainerRuntimeContainerd](#containerruntimecontainerd) | false |

[Back to Group](#v1beta2)

### ContainerRuntimeContainerd

ContainerRuntimeContainerd defines docker container runtime

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| registries | A map of registries to use to render configs and mirrors for containerd registries | map[string][ContainerdRegistry](#containerdregistry) | false |

[Back to Group](#v1beta2)

### ContainerRuntimeDocker

ContainerRuntimeDocker defines docker container runtime

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| registryMirrors | Configures dockerd with \"registry-mirrors\" | []string | true |

[Back to Group](#v1beta2)

### ContainerdRegistry

ContainerdRegistry defines endpoints and security for given container registry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| mirrors | List of registry mirrors to use | []string | false |
| tlsConfig | TLSConfig for the registry | *[ContainerdTLSConfig](#containerdtlsconfig) | false |
| auth | Registry authentication | *[ContainerdRegistryAuthConfig](#containerdregistryauthconfig) | false |

[Back to Group](#v1beta2)

### ContainerdRegistryAuthConfig

Containerd per-registry credentials config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| username |  | string | false |
| password |  | string | false |
| auth |  | string | false |
| identityToken |  | string | false |

[Back to Group](#v1beta2)

### ContainerdTLSConfig

Configures containerd TLS for a registry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| insecureSkipVerify | Don't validate remote TLS certificate | bool | false |

[Back to Group](#v1beta2)

### ControlPlaneConfig

ControlPlaneConfig defines control plane nodes

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| hosts | Hosts array of all control plane hosts. | [][HostConfig](#hostconfig) | true |

[Back to Group](#v1beta2)

### DNSConfig

DNSConfig contains a machine's DNS configuration

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| servers | Servers | []string | true |

[Back to Group](#v1beta2)

### DigitalOceanSpec

DigitalOceanSpec defines the DigitalOcean cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta2)

### DynamicAuditLog

DynamicAuditLog feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable Default value is false. | bool | false |

[Back to Group](#v1beta2)

### DynamicWorkerConfig

DynamicWorkerConfig describes a set of worker machines

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name | string | true |
| replicas | Replicas | *int | true |
| providerSpec | Config | [ProviderSpec](#providerspec) | true |

[Back to Group](#v1beta2)

### EncryptionProviders

Encryption Providers feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | true |
| customEncryptionConfiguration | CustomEncryptionConfiguration | string | true |

[Back to Group](#v1beta2)

### EquinixMetalSpec

EquinixMetalSpec defines the Equinix Metal cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta2)

### ExternalCNISpec

ExternalCNISpec defines the external CNI plugin.
It's up to the user's responsibility to deploy the external CNI plugin manually or as an addon

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta2)

### Features

Features controls what features will be enabled on the cluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podNodeSelector | PodNodeSelector | *[PodNodeSelector](#podnodeselector) | false |
| podSecurityPolicy | PodSecurityPolicy Deprecated: will be removed once Kubernetes 1.24 reaches EOL | *[PodSecurityPolicy](#podsecuritypolicy) | false |
| staticAuditLog | StaticAuditLog | *[StaticAuditLog](#staticauditlog) | false |
| dynamicAuditLog | DynamicAuditLog | *[DynamicAuditLog](#dynamicauditlog) | false |
| metricsServer | MetricsServer | *[MetricsServer](#metricsserver) | false |
| openidConnect | OpenIDConnect | *[OpenIDConnect](#openidconnect) | false |
| encryptionProviders | Encryption Providers | *[EncryptionProviders](#encryptionproviders) | false |

[Back to Group](#v1beta2)

### GCESpec

GCESpec defines the GCE cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta2)

### HetznerSpec

HetznerSpec defines the Hetzner cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| networkID | NetworkID | string | false |

[Back to Group](#v1beta2)

### HostConfig

HostConfig describes a single control plane node.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| publicAddress | PublicAddress is externally accessible IP address from public internet. | string | true |
| privateAddress | PrivateAddress is internal RFC-1918 IP address. | string | true |
| sshPort | SSHPort is port to connect ssh to. Default value is 22. | int | false |
| sshUsername | SSHUsername is system login name. Default value is \"root\". | string | false |
| sshPrivateKeyFile | SSHPrivateKeyFile is path to the file with PRIVATE AND CLEANTEXT ssh key. Default value is \"\". | string | false |
| sshAgentSocket | SSHAgentSocket path (or reference to the environment) to the SSH agent unix domain socket. Default value is \"env:SSH_AUTH_SOCK\". | string | false |
| bastion | Bastion is an IP or hostname of the bastion (or jump) host to connect to. Default value is \"\". | string | false |
| bastionPort | BastionPort is SSH port to use when connecting to the bastion if it's configured in .Bastion. Default value is 22. | int | false |
| bastionUser | BastionUser is system login name to use when connecting to bastion host. Default value is \"root\". | string | false |
| hostname | Hostname is the hostname(1) of the host. Default value is populated at the runtime via running `hostname -f` command over ssh. | string | false |
| isLeader | IsLeader indicates this host as a session leader. Default value is populated at the runtime. | bool | false |
| taints | Taints if not provided (i.e. nil) defaults to TaintEffectNoSchedule, with key node-role.kubernetes.io/master for control plane nodes. Explicitly empty (i.e. []corev1.Taint{}) means no taints will be applied (this is default for worker nodes). | [][corev1.Taint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#taint-v1-core) | false |
| kubelet | Kubelet | [KubeletConfig](#kubeletconfig) | false |
| operatingSystem | OperatingSystem information, can be populated at the runtime. | OperatingSystemName | false |

[Back to Group](#v1beta2)

### IPTables

IPTables

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta2)

### IPVSConfig

IPVSConfig contains different options to configure IPVS kube-proxy mode

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| scheduler | ipvs scheduler, if it’s not configured, then round-robin (rr) is the default value. Can be one of: * rr: round-robin * lc: least connection (smallest number of open connections) * dh: destination hashing * sh: source hashing * sed: shortest expected delay * nq: never queue | string | true |
| excludeCIDRs | excludeCIDRs is a list of CIDR's which the ipvs proxier should not touch when cleaning up ipvs services. | []string | true |
| strictARP | strict ARP configure arp_ignore and arp_announce to avoid answering ARP queries from kube-ipvs0 interface | bool | true |
| tcpTimeout | tcpTimeout is the timeout value used for idle IPVS TCP sessions. The default value is 0, which preserves the current timeout value on the system. | metav1.Duration | true |
| tcpFinTimeout | tcpFinTimeout is the timeout value used for IPVS TCP sessions after receiving a FIN. The default value is 0, which preserves the current timeout value on the system. | metav1.Duration | true |
| udpTimeout | udpTimeout is the timeout value used for IPVS UDP packets. The default value is 0, which preserves the current timeout value on the system. | metav1.Duration | true |

[Back to Group](#v1beta2)

### ImageAsset

ImageAsset is used to customize the image repository and the image tag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| imageRepository | ImageRepository customizes the registry/repository | string | false |
| imageTag | ImageTag customizes the image tag | string | false |

[Back to Group](#v1beta2)

### KubeOneCluster

KubeOneCluster is KubeOne Cluster API Schema

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name is the name of the cluster. | string | true |
| controlPlane | ControlPlane describes the control plane nodes and how to access them. | [ControlPlaneConfig](#controlplaneconfig) | true |
| apiEndpoint | APIEndpoint are pairs of address and port used to communicate with the Kubernetes API. | [APIEndpoint](#apiendpoint) | true |
| cloudProvider | CloudProvider configures the cloud provider specific features. | [CloudProviderSpec](#cloudproviderspec) | true |
| versions | Versions defines which Kubernetes version will be installed. | [VersionConfig](#versionconfig) | true |
| containerRuntime | ContainerRuntime defines which container runtime will be installed | [ContainerRuntimeConfig](#containerruntimeconfig) | false |
| clusterNetwork | ClusterNetwork configures the in-cluster networking. | [ClusterNetworkConfig](#clusternetworkconfig) | false |
| proxy | Proxy configures proxy used while installing Kubernetes and by the Docker daemon. | [ProxyConfig](#proxyconfig) | false |
| staticWorkers | StaticWorkers describes the worker nodes that are managed by KubeOne/kubeadm. | [StaticWorkersConfig](#staticworkersconfig) | false |
| dynamicWorkers | DynamicWorkers describes the worker nodes that are managed by Kubermatic machine-controller/Cluster-API. | [][DynamicWorkerConfig](#dynamicworkerconfig) | false |
| machineController | MachineController configures the Kubermatic machine-controller component. | *[MachineControllerConfig](#machinecontrollerconfig) | false |
| caBundle | CABundle PEM encoded global CA | string | false |
| features | Features enables and configures additional cluster features. | [Features](#features) | false |
| addons | Addons are used to deploy additional manifests. | *[Addons](#addons) | false |
| systemPackages | SystemPackages configure kubeone behaviour regarding OS packages. | *[SystemPackages](#systempackages) | false |
| registryConfiguration | RegistryConfiguration configures how Docker images are pulled from an image registry | *[RegistryConfiguration](#registryconfiguration) | false |
| loggingConfig | LoggingConfig configures the Kubelet's log configuration | [LoggingConfig](#loggingconfig) | false |

[Back to Group](#v1beta2)

### KubeProxyConfig

KubeProxyConfig defines configured kube-proxy mode, default is iptables mode

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| skipInstallation | SkipInstallation will skip the installation of kube-proxy default value is false | bool | true |
| ipvs | IPVS config | *[IPVSConfig](#ipvsconfig) | true |
| iptables | IPTables config | *[IPTables](#iptables) | true |

[Back to Group](#v1beta2)

### KubeletConfig

KubeletConfig provides some kubelet configuration options

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| systemReserved | SystemReserved configure --system-reserved command-line flag of the kubelet. See more at: https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/ | map[string]string | false |
| kubeReserved | KubeReserved configure --kube-reserved command-line flag of the kubelet. See more at: https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/ | map[string]string | false |
| evictionHard | EvictionHard configure --eviction-hard command-line flag of the kubelet. See more at: https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/ | map[string]string | false |

[Back to Group](#v1beta2)

### LoggingConfig

LoggingConfig configures the Kubelet's log rotation

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| containerLogMaxSize | ContainerLogMaxSize configures the maximum size of container log file before it is rotated See more at: https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/ | string | false |
| containerLogMaxFiles | ContainerLogMaxFiles configures the maximum number of container log files that can be present for a container See more at: https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/ | int32 | false |

[Back to Group](#v1beta2)

### MachineControllerConfig

MachineControllerConfig configures kubermatic machine-controller deployment

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| deploy | Deploy | bool | false |

[Back to Group](#v1beta2)

### MetricsServer

MetricsServer feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable deployment of metrics-server. Default value is true. | bool | false |

[Back to Group](#v1beta2)

### NoneSpec

NoneSpec defines a none provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta2)

### NutanixSpec

NutanixSpec defines the Nutanix provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta2)

### OpenIDConnect

OpenIDConnect feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |
| config | Config | [OpenIDConnectConfig](#openidconnectconfig) | true |

[Back to Group](#v1beta2)

### OpenIDConnectConfig

OpenIDConnectConfig config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| issuerUrl | IssuerURL | string | true |
| clientId | ClientID | string | true |
| usernameClaim | UsernameClaim | string | true |
| usernamePrefix | UsernamePrefix | string | true |
| groupsClaim | GroupsClaim | string | true |
| groupsPrefix | GroupsPrefix | string | true |
| requiredClaim | RequiredClaim | string | true |
| signingAlgs | SigningAlgs | string | true |
| caFile | CAFile | string | true |

[Back to Group](#v1beta2)

### OpenstackSpec

OpenstackSpec defines the Openstack provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta2)

### PodNodeSelector

PodNodeSelector feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |
| config | Config | [PodNodeSelectorConfig](#podnodeselectorconfig) | true |

[Back to Group](#v1beta2)

### PodNodeSelectorConfig

PodNodeSelectorConfig config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| configFilePath | ConfigFilePath is a path on the local file system to the PodNodeSelector configuration file. ConfigFilePath is a required field. More info: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#podnodeselector | string | true |

[Back to Group](#v1beta2)

### PodSecurityPolicy

PodSecurityPolicy feature flag
This feature is deprecated and will be removed from the API once
Kubernetes 1.24 reaches EOL.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |

[Back to Group](#v1beta2)

### ProviderSpec

ProviderSpec describes a worker node

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| cloudProviderSpec | CloudProviderSpec | [json.RawMessage](https://golang.org/pkg/encoding/json/#RawMessage) | true |
| annotations | Annotations set MachineDeployment.ObjectMeta.Annotations | map[string]string | false |
| machineAnnotations | MachineAnnotations set MachineDeployment.Spec.Template.Spec.ObjectMeta.Annotations a way to annotate resulted Nodes | map[string]string | false |
| labels | Labels | map[string]string | false |
| taints | Taints | [][corev1.Taint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#taint-v1-core) | false |
| sshPublicKeys | SSHPublicKeys | []string | false |
| operatingSystem | OperatingSystem | string | true |
| operatingSystemSpec | OperatingSystemSpec | [json.RawMessage](https://golang.org/pkg/encoding/json/#RawMessage) | false |
| network | Network | *[ProviderStaticNetworkConfig](#providerstaticnetworkconfig) | false |
| overwriteCloudConfig | OverwriteCloudConfig | *string | false |

[Back to Group](#v1beta2)

### ProviderStaticNetworkConfig

ProviderStaticNetworkConfig contains a machine's static network configuration

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| cidr | CIDR | string | true |
| gateway | Gateway | string | true |
| dns | DNS | [DNSConfig](#dnsconfig) | true |

[Back to Group](#v1beta2)

### ProxyConfig

ProxyConfig configures proxy for the Docker daemon and is used by KubeOne scripts

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| http | HTTP | string | false |
| https | HTTPS | string | false |
| noProxy | NoProxy | string | false |

[Back to Group](#v1beta2)

### RegistryConfiguration

RegistryConfiguration controls how images used for components deployed by
KubeOne and kubeadm are pulled from an image registry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| overwriteRegistry | OverwriteRegistry specifies a custom Docker registry which will be used for all images required for KubeOne and kubeadm. This also applies to addons deployed by KubeOne. This field doesn't modify the user/organization part of the image. For example, if OverwriteRegistry is set to 127.0.0.1:5000/example, image called calico/cni would translate to 127.0.0.1:5000/example/calico/cni. Default: \"\" | string | false |
| insecureRegistry | InsecureRegistry configures Docker to threat the registry specified in OverwriteRegistry as an insecure registry. This is also propagated to the worker nodes managed by machine-controller and/or KubeOne. | bool | false |

[Back to Group](#v1beta2)

### StaticAuditLog

StaticAuditLog feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |
| config | Config | [StaticAuditLogConfig](#staticauditlogconfig) | true |

[Back to Group](#v1beta2)

### StaticAuditLogConfig

StaticAuditLogConfig config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| policyFilePath | PolicyFilePath is a path on local file system to the audit policy manifest which defines what events should be recorded and what data they should include. PolicyFilePath is a required field. More info: https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#audit-policy | string | true |
| logPath | LogPath is path on control plane instances where audit log files are stored. Default value is /var/log/kubernetes/audit.log | string | false |
| logMaxAge | LogMaxAge is maximum number of days to retain old audit log files. Default value is 30 | int | false |
| logMaxBackup | LogMaxBackup is maximum number of audit log files to retain. Default value is 3. | int | false |
| logMaxSize | LogMaxSize is maximum size in megabytes of audit log file before it gets rotated. Default value is 100. | int | false |

[Back to Group](#v1beta2)

### StaticWorkersConfig

StaticWorkersConfig defines static worker nodes provisioned by KubeOne and kubeadm

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| hosts | Hosts | [][HostConfig](#hostconfig) | false |

[Back to Group](#v1beta2)

### SystemPackages

SystemPackages controls configurations of APT/YUM

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| configureRepositories | ConfigureRepositories (true by default) is a flag to control automatic configuration of kubeadm / docker repositories. | bool | false |

[Back to Group](#v1beta2)

### VersionConfig

VersionConfig describes the versions of components that are installed on the machines

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kubernetes |  | string | true |

[Back to Group](#v1beta2)

### VsphereSpec

VsphereSpec defines the vSphere provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta2)

### WeaveNetSpec

WeaveNetSpec defines the WeaveNet CNI plugin

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| encrypted | Encrypted | bool | false |

[Back to Group](#v1beta2)
