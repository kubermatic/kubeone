+++
title = "v1beta1 API Reference"
date = 2020-08-07T12:41:33+03:00
weight = 11
+++
## v1beta1

* [APIEndpoint](#apiendpoint)
* [AWSSpec](#awsspec)
* [Addons](#addons)
* [AzureSpec](#azurespec)
* [CNI](#cni)
* [CanalSpec](#canalspec)
* [CloudProviderSpec](#cloudproviderspec)
* [ClusterNetworkConfig](#clusternetworkconfig)
* [ControlPlaneConfig](#controlplaneconfig)
* [DNSConfig](#dnsconfig)
* [DigitalOceanSpec](#digitaloceanspec)
* [DynamicAuditLog](#dynamicauditlog)
* [DynamicWorkerConfig](#dynamicworkerconfig)
* [ExternalCNISpec](#externalcnispec)
* [Features](#features)
* [GCESpec](#gcespec)
* [HetznerSpec](#hetznerspec)
* [HostConfig](#hostconfig)
* [KubeOneCluster](#kubeonecluster)
* [MachineControllerConfig](#machinecontrollerconfig)
* [MetricsServer](#metricsserver)
* [NoneSpec](#nonespec)
* [OpenIDConnect](#openidconnect)
* [OpenIDConnectConfig](#openidconnectconfig)
* [OpenstackSpec](#openstackspec)
* [PacketSpec](#packetspec)
* [PodNodeSelector](#podnodeselector)
* [PodNodeSelectorConfig](#podnodeselectorconfig)
* [PodPresets](#podpresets)
* [PodSecurityPolicy](#podsecuritypolicy)
* [ProviderSpec](#providerspec)
* [ProviderStaticNetworkConfig](#providerstaticnetworkconfig)
* [ProxyConfig](#proxyconfig)
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

[Back to Group](#v1beta1)

### AWSSpec

AWSSpec defines the AWS cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta1)

### Addons

Addons config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |
| path | Path on the local file system to the directory with addons manifests. | string | true |

[Back to Group](#v1beta1)

### AzureSpec

AzureSpec defines the Azure cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta1)

### CNI

CNI config. Only one CNI provider must be used at the single time.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| canal | Canal | *[CanalSpec](#canalspec) | false |
| weaveNet | WeaveNet | *[WeaveNetSpec](#weavenetspec) | false |
| external | External | *[ExternalCNISpec](#externalcnispec) | false |

[Back to Group](#v1beta1)

### CanalSpec

CanalSpec defines the Canal CNI plugin

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| mtu | MTU automatically detected based on the cloudProvider default value is 1450 | int | false |

[Back to Group](#v1beta1)

### CloudProviderSpec

CloudProviderSpec describes the cloud provider that is running the machines.
Only one cloud provider must be defined at the single time.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| external | External | bool | false |
| cloudConfig | CloudConfig | string | false |
| aws | AWS | *[AWSSpec](#awsspec) | false |
| azure | Azure | *[AzureSpec](#azurespec) | false |
| digitalocean | DigitalOcean | *[DigitalOceanSpec](#digitaloceanspec) | false |
| gce | GCE | *[GCESpec](#gcespec) | false |
| hetzner | Hetzner | *[HetznerSpec](#hetznerspec) | false |
| openstack | Openstack | *[OpenstackSpec](#openstackspec) | false |
| packet | Packet | *[PacketSpec](#packetspec) | false |
| vsphere | Vsphere | *[VsphereSpec](#vspherespec) | false |
| none | None | *[NoneSpec](#nonespec) | false |

[Back to Group](#v1beta1)

### ClusterNetworkConfig

ClusterNetworkConfig describes the cluster network

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podSubnet | PodSubnet default value is \"10.244.0.0/16\" | string | false |
| serviceSubnet | ServiceSubnet default value is \"10.96.0.0/12\" | string | false |
| serviceDomainName | ServiceDomainName default value is \"cluster.local\" | string | false |
| nodePortRange | NodePortRange default value is \"30000-32767\" | string | false |
| cni | CNI default value is {canal: {mtu: 1450}} | *[CNI](#cni) | false |

[Back to Group](#v1beta1)

### ControlPlaneConfig

ControlPlaneConfig defines control plane nodes

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| hosts | Hosts array of all control plane hosts. | [][HostConfig](#hostconfig) | true |

[Back to Group](#v1beta1)

### DNSConfig

DNSConfig contains a machine's DNS configuration

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| servers | Servers | []string | true |

[Back to Group](#v1beta1)

### DigitalOceanSpec

DigitalOceanSpec defines the DigitalOcean cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta1)

### DynamicAuditLog

DynamicAuditLog feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable Default value is false. | bool | false |

[Back to Group](#v1beta1)

### DynamicWorkerConfig

DynamicWorkerConfig describes a set of worker machines

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name | string | true |
| replicas | Replicas | *int | true |
| providerSpec | Config | [ProviderSpec](#providerspec) | true |

[Back to Group](#v1beta1)

### ExternalCNISpec

ExternalCNISpec defines the external CNI plugin.
It's up to the user's responsibility to deploy the external CNI plugin manually or as an addon

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta1)

### Features

Features controls what features will be enabled on the cluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podNodeSelector | PodNodeSelector | *[PodNodeSelector](#podnodeselector) | false |
| podPresets | PodPresets | *[PodPresets](#podpresets) | false |
| podSecurityPolicy | PodSecurityPolicy | *[PodSecurityPolicy](#podsecuritypolicy) | false |
| staticAuditLog | StaticAuditLog | *[StaticAuditLog](#staticauditlog) | false |
| dynamicAuditLog | DynamicAuditLog | *[DynamicAuditLog](#dynamicauditlog) | false |
| metricsServer | MetricsServer | *[MetricsServer](#metricsserver) | false |
| openidConnect | OpenIDConnect | *[OpenIDConnect](#openidconnect) | false |

[Back to Group](#v1beta1)

### GCESpec

GCESpec defines the GCE cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta1)

### HetznerSpec

HetznerSpec defines the Hetzner cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| networkID | NetworkID | string | false |

[Back to Group](#v1beta1)

### HostConfig

HostConfig describes a single control plane node.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| publicAddress | PublicAddress is externally accessible IP address from public internet. | string | true |
| privateAddress | PrivateAddress is internal RFC-1918 IP address. | string | true |
| sshPort | SSHPort is port to connect ssh to. Default value is 22. | int | false |
| sshUsername | SSHUsername is system login name. Default value is \"root\". | string | false |
| sshPrivateKeyFile | SSHPrivateKeyFile is path to the file with PRIVATE AND CLEANTEXT ssh key. Default value is \"\". | string | false |
| sshAgentSocket | SSHAgentSocket path (or reference to the environment) to the SSH agent unix domain socket. Default vaulue is \"env:SSH_AUTH_SOCK\". | string | false |
| bastion | Bastion is an IP or hostname of the bastion (or jump) host to connect to. Default value is \"\". | string | false |
| bastionPort | BastionPort is SSH port to use when connecting to the bastion if it's configured in .Bastion. Default value is 22. | int | false |
| bastionUser | BastionUser is system login name to use when connecting to bastion host. Default value is \"root\". | string | false |
| hostname | Hostname is the hostname(1) of the host. Default value is populated at the runtime via running `hostname -f` command over ssh. | string | false |
| isLeader | IsLeader indicates this host as a session leader. Default vaule is populated at the runtime. | bool | false |
| taints | Taints if not provided (i.e. nil) defaults to TaintEffectNoSchedule, with key node-role.kubernetes.io/master for control plane nodes. Explicitly empty (i.e. []corev1.Taint{}) means no taints will be applied (this is default for worker nodes). | [][corev1.Taint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#taint-v1-core) | false |

[Back to Group](#v1beta1)

### KubeOneCluster

KubeOneCluster is KubeOne Cluster API Schema

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name is the name of the cluster. | string | true |
| controlPlane | ControlPlane describes the control plane nodes and how to access them. | [ControlPlaneConfig](#controlplaneconfig) | true |
| apiEndpoint | APIEndpoint are pairs of address and port used to communicate with the Kubernetes API. | [APIEndpoint](#apiendpoint) | true |
| cloudProvider | CloudProvider configures the cloud provider specific features. | [CloudProviderSpec](#cloudproviderspec) | true |
| versions | Versions defines which Kubernetes version will be installed. | [VersionConfig](#versionconfig) | true |
| clusterNetwork | ClusterNetwork configures the in-cluster networking. | [ClusterNetworkConfig](#clusternetworkconfig) | false |
| proxy | Proxy configures proxy used while installing Kubernetes and by the Docker daemon. | [ProxyConfig](#proxyconfig) | false |
| staticWorkers | StaticWorkers describes the worker nodes that are managed by KubeOne/kubeadm. | [StaticWorkersConfig](#staticworkersconfig) | false |
| dynamicWorkers | DynamicWorkers describes the worker nodes that are managed by Kubermatic machine-controller/Cluster-API. | [][DynamicWorkerConfig](#dynamicworkerconfig) | false |
| machineController | MachineController configures the Kubermatic machine-controller component. | *[MachineControllerConfig](#machinecontrollerconfig) | false |
| features | Features enables and configures additional cluster features. | [Features](#features) | false |
| addons | Addons are used to deploy additional manifests. | *[Addons](#addons) | false |
| systemPackages | SystemPackages configure kubeone behaviour regarding OS packages. | *[SystemPackages](#systempackages) | false |

[Back to Group](#v1beta1)

### MachineControllerConfig

MachineControllerConfig configures kubermatic machine-controller deployment

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| deploy | Deploy | bool | false |

[Back to Group](#v1beta1)

### MetricsServer

MetricsServer feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable deployment of metrics-server. Default value is true. | bool | false |

[Back to Group](#v1beta1)

### NoneSpec

NoneSpec defines a none provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta1)

### OpenIDConnect

OpenIDConnect feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |
| config | Config | [OpenIDConnectConfig](#openidconnectconfig) | true |

[Back to Group](#v1beta1)

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

[Back to Group](#v1beta1)

### OpenstackSpec

OpenstackSpec defines the Openstack provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta1)

### PacketSpec

PacketSpec defines the Packet cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta1)

### PodNodeSelector

PodNodeSelector feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |
| config | Config | [PodNodeSelectorConfig](#podnodeselectorconfig) | true |

[Back to Group](#v1beta1)

### PodNodeSelectorConfig

PodNodeSelectorConfig config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| configFilePath | ConfigFilePath is a path on the local file system to the PodNodeSelector configuration file. ConfigFilePath is a required field. More info: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#podnodeselector | string | true |

[Back to Group](#v1beta1)

### PodPresets

PodPresets feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |

[Back to Group](#v1beta1)

### PodSecurityPolicy

PodSecurityPolicy feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |

[Back to Group](#v1beta1)

### ProviderSpec

ProviderSpec describes a worker node

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| cloudProviderSpec | CloudProviderSpec | [json.RawMessage](https://golang.org/pkg/encoding/json/#RawMessage) | true |
| labels | Labels | map[string]string | false |
| taints | Taints | [][corev1.Taint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#taint-v1-core) | false |
| sshPublicKeys | SSHPublicKeys | []string | false |
| operatingSystem | OperatingSystem | string | true |
| operatingSystemSpec | OperatingSystemSpec | [json.RawMessage](https://golang.org/pkg/encoding/json/#RawMessage) | false |
| network | Network | *[ProviderStaticNetworkConfig](#providerstaticnetworkconfig) | false |
| overwriteCloudConfig | OverwriteCloudConfig | *string | false |

[Back to Group](#v1beta1)

### ProviderStaticNetworkConfig

ProviderStaticNetworkConfig contains a machine's static network configuration

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| cidr | CIDR | string | true |
| gateway | Gateway | string | true |
| dns | DNS | [DNSConfig](#dnsconfig) | true |

[Back to Group](#v1beta1)

### ProxyConfig

ProxyConfig configures proxy for the Docker daemon and is used by KubeOne scripts

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| http | HTTP | string | false |
| https | HTTPS | string | false |
| noProxy | NoProxy | string | false |

[Back to Group](#v1beta1)

### StaticAuditLog

StaticAuditLog feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |
| config | Config | [StaticAuditLogConfig](#staticauditlogconfig) | true |

[Back to Group](#v1beta1)

### StaticAuditLogConfig

StaticAuditLogConfig config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| policyFilePath | PolicyFilePath is a path on local file system to the audit policy manifest which defines what events should be recorded and what data they should include. PolicyFilePath is a required field. More info: https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#audit-policy | string | true |
| logPath | LogPath is path on control plane instances where audit log files are stored. Default value is /var/log/kubernetes/audit.log | string | false |
| logMaxAge | LogMaxAge is maximum number of days to retain old audit log files. Default value is 30 | int | false |
| logMaxBackup | LogMaxBackup is maximum number of audit log files to retain. Default value is 3. | int | false |
| logMaxSize | LogMaxSize is maximum size in megabytes of audit log file before it gets rotated. Default value is 100. | int | false |

[Back to Group](#v1beta1)

### StaticWorkersConfig

StaticWorkersConfig defines static worker nodes provisioned by KubeOne and kubeadm

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| hosts | Hosts | [][HostConfig](#hostconfig) | false |

[Back to Group](#v1beta1)

### SystemPackages

SystemPackages controls configurations of APT/YUM

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| configureRepositories | ConfigureRepositories (true by default) is a flag to control automatic configuration of kubeadm / docker repositories. | bool | false |

[Back to Group](#v1beta1)

### VersionConfig

VersionConfig describes the versions of components that are installed on the machines

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kubernetes |  | string | true |

[Back to Group](#v1beta1)

### VsphereSpec

VsphereSpec defines the vSphere provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta1)

### WeaveNetSpec

WeaveNetSpec defines the WeaveNet CNI plugin

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| encrypted | Encrypted | bool | false |

[Back to Group](#v1beta1)
