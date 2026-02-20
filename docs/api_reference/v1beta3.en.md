+++
title = "v1beta3 API Reference"
date = 2026-02-23T21:01:35+01:00
weight = 11
+++
## v1beta3

* [APIEndpoint](#apiendpoint)
* [AWSSpec](#awsspec)
* [Addon](#addon)
* [AddonRef](#addonref)
* [Addons](#addons)
* [AzureSpec](#azurespec)
* [CNI](#cni)
* [CanalSpec](#canalspec)
* [CertificateAuthorithyConfig](#certificateauthorithyconfig)
* [CiliumSpec](#ciliumspec)
* [CloudProviderSpec](#cloudproviderspec)
* [ClusterNetworkConfig](#clusternetworkconfig)
* [ContainerRuntimeConfig](#containerruntimeconfig)
* [ContainerRuntimeContainerd](#containerruntimecontainerd)
* [ContainerdRegistry](#containerdregistry)
* [ContainerdRegistryAuthConfig](#containerdregistryauthconfig)
* [ContainerdTLSConfig](#containerdtlsconfig)
* [ControlPlaneComponentConfig](#controlplanecomponentconfig)
* [ControlPlaneComponents](#controlplanecomponents)
* [ControlPlaneConfig](#controlplaneconfig)
* [CoreDNS](#coredns)
* [DNSConfig](#dnsconfig)
* [DigitalOceanSpec](#digitaloceanspec)
* [DynamicAuditLog](#dynamicauditlog)
* [DynamicWorkerConfig](#dynamicworkerconfig)
* [EncryptionProviders](#encryptionproviders)
* [EquinixMetalSpec](#equinixmetalspec)
* [ExternalCNISpec](#externalcnispec)
* [Features](#features)
* [GCESpec](#gcespec)
* [HelmAuth](#helmauth)
* [HelmRelease](#helmrelease)
* [HelmValues](#helmvalues)
* [HetznerSpec](#hetznerspec)
* [HostConfig](#hostconfig)
* [IPTables](#iptables)
* [IPVSConfig](#ipvsconfig)
* [KubeOneCluster](#kubeonecluster)
* [KubeProxyConfig](#kubeproxyconfig)
* [KubeletConfig](#kubeletconfig)
* [KubevirtSpec](#kubevirtspec)
* [LoggingConfig](#loggingconfig)
* [MachineControllerConfig](#machinecontrollerconfig)
* [MetricsServer](#metricsserver)
* [NodeLocalDNS](#nodelocaldns)
* [NoneSpec](#nonespec)
* [NutanixSpec](#nutanixspec)
* [OpenIDConnect](#openidconnect)
* [OpenIDConnectConfig](#openidconnectconfig)
* [OpenstackSpec](#openstackspec)
* [OperatingSystemManagerConfig](#operatingsystemmanagerconfig)
* [PodNodeSelector](#podnodeselector)
* [PodNodeSelectorConfig](#podnodeselectorconfig)
* [ProviderSpec](#providerspec)
* [ProviderStaticNetworkConfig](#providerstaticnetworkconfig)
* [ProxyConfig](#proxyconfig)
* [RegistryConfiguration](#registryconfiguration)
* [StaticAuditLog](#staticauditlog)
* [StaticAuditLogConfig](#staticauditlogconfig)
* [StaticWorkersConfig](#staticworkersconfig)
* [SystemPackages](#systempackages)
* [TLSCipherSuites](#tlsciphersuites)
* [VMwareCloudDirectorSpec](#vmwareclouddirectorspec)
* [VersionConfig](#versionconfig)
* [VsphereSpec](#vspherespec)
* [WeaveNetSpec](#weavenetspec)
* [WebHookAuditLogBatchConfig](#webhookauditlogbatchconfig)
* [WebHookAuditLogThrottleConfig](#webhookauditlogthrottleconfig)
* [WebHookAuditLogTruncateConfig](#webhookauditlogtruncateconfig)
* [WebhookAuditLog](#webhookauditlog)
* [WebhookAuditLogConfig](#webhookauditlogconfig)

### APIEndpoint

APIEndpoint is the endpoint used to communicate with the Kubernetes API

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| host | Host is the hostname or IP on which API is running. | string | true |
| port | Port is the port used to reach to the API. Default value is 6443. | int | false |
| alternativeNames | AlternativeNames is a list of Subject Alternative Names for the API Server signing cert. | []string | false |

[Back to Group](#v1beta3)

### AWSSpec

AWSSpec defines the AWS cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta3)

### Addon

Addon config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of the addon to configure | string | true |
| params | Params to the addon, to render the addon using text/template, this will override globalParams | map[string]string | false |
| disableTemplating | DisableTemplating is used to disable templatization for the addon. | bool | false |
| delete | Delete flag to ensure the named addon with all its contents to be deleted | bool | false |

[Back to Group](#v1beta3)

### AddonRef



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| addon | KubeOne's internal Addon | *[Addon](#addon) | false |
| helmRelease | HelmReleases configure helm charts to reconcile. For each HelmRelease it will run analog of: `helm upgrade --namespace <NAMESPACE> --install --create-namespace <RELEASE> <CHART> [--values=values-override.yaml]` | *[HelmRelease](#helmrelease) | false |

[Back to Group](#v1beta3)

### Addons

Addons config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| path | Path on the local file system to the directory with addons manifests. | string | false |
| addons | Addons is a list of config options for named addon | [][AddonRef](#addonref) | false |

[Back to Group](#v1beta3)

### AzureSpec

AzureSpec defines the Azure cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta3)

### CNI

CNI config. Only one CNI provider must be used at the single time.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| canal | Canal | *[CanalSpec](#canalspec) | false |
| cilium | Cilium | *[CiliumSpec](#ciliumspec) | false |
| weaveNet | WeaveNet | *[WeaveNetSpec](#weavenetspec) | false |
| external | External | *[ExternalCNISpec](#externalcnispec) | false |

[Back to Group](#v1beta3)

### CanalSpec

CanalSpec defines the Canal CNI plugin

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| mtu | MTU automatically detected based on the cloudProvider default value is 1450 | int | false |

[Back to Group](#v1beta3)

### CertificateAuthorithyConfig



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| bundle | Bundle inline PEM encoded global CA | string | false |
| file | File is a path to the CA bundle file, used as a replacement for Bundle | string | false |
| certificateValidityPeriod | CertificateValidityPeriod specifies the validity period for a non-CA certificate generated by kubeadm. Default value: 8760h (365 days * 24 hours = 1 year) | *metav1.Duration | false |
| caCertificateValidityPeriod | CACertificateValidityPeriod specifies the validity period for a CA certificate generated by kubeadm. Default value: 87600h (365 days * 24 hours * 10 = 10 years) | *metav1.Duration | false |

[Back to Group](#v1beta3)

### CiliumSpec

CiliumSpec defines the Cilium CNI plugin

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kubeProxyReplacement | KubeProxyReplacement defines weather cilium relies on underlying Kernel support to replace kube-proxy functionality by eBPF (strict), or disables a subset of those features so cilium does not bail out if the kernel support is missing (disabled). default is false | bool | true |
| enableHubble | EnableHubble to deploy Hubble relay and UI default value is false | bool | true |
| enableL2Announcements | EnableL2Announcements enables the Layer 2 announcement feature for the Cilium CNI plugin. If not set, Cilium will use its default behavior. | bool | true |

[Back to Group](#v1beta3)

### CloudProviderSpec

CloudProviderSpec describes the cloud provider that is running the machines.
Only one cloud provider must be defined at the single time.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| external | External | bool | false |
| disableBundledCSIDrivers | DisableBundledCSIDrivers disables automatic deployment of CSI drivers bundled with KubeOne | bool | true |
| cloudConfig | CloudConfig | string | false |
| csiConfig | CSIConfig | string | false |
| secretProviderClassName | SecretProviderClassName | string | false |
| aws | AWS | *[AWSSpec](#awsspec) | false |
| azure | Azure | *[AzureSpec](#azurespec) | false |
| digitalocean | DigitalOcean | *[DigitalOceanSpec](#digitaloceanspec) | false |
| gce | GCE | *[GCESpec](#gcespec) | false |
| hetzner | Hetzner | *[HetznerSpec](#hetznerspec) | false |
| kubevirt | Kubevirt | *[KubevirtSpec](#kubevirtspec) | false |
| nutanix | Nutanix | *[NutanixSpec](#nutanixspec) | false |
| openstack | Openstack | *[OpenstackSpec](#openstackspec) | false |
| equinixmetal | EquinixMetal | *[EquinixMetalSpec](#equinixmetalspec) | false |
| vmwareCloudDirector | VMware Cloud Director | *[VMwareCloudDirectorSpec](#vmwareclouddirectorspec) | false |
| vsphere | Vsphere | *[VsphereSpec](#vspherespec) | false |
| none | None | *[NoneSpec](#nonespec) | false |

[Back to Group](#v1beta3)

### ClusterNetworkConfig

ClusterNetworkConfig describes the cluster network

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podSubnet | PodSubnet default value is \"10.244.0.0/16\" | string | false |
| podSubnetIPv6 | PodSubnetIPv6 default value is \"\"fd01::/48\"\" | string | false |
| serviceSubnet | ServiceSubnet default value is \"10.96.0.0/12\" | string | false |
| serviceSubnetIPv6 | ServiceSubnetIPv6 default value is \"fd02::/120\" | string | false |
| serviceDomainName | ServiceDomainName default value is \"cluster.local\" | string | false |
| nodePortRange | NodePortRange default value is \"30000-32767\" | string | false |
| cni | CNI default value is {canal: {mtu: 1450}} | *[CNI](#cni) | false |
| kubeProxy | KubeProxy config | *[KubeProxyConfig](#kubeproxyconfig) | false |
| ipFamily | IPFamily allows specifying IP family of a cluster. Valid values are IPv4 \| IPv6 \| IPv4+IPv6 \| IPv6+IPv4. | IPFamily | false |
| nodeCIDRMaskSizeIPv4 | NodeCIDRMaskSizeIPv4 is the mask size used to address the nodes within provided IPv4 Pods CIDR. It has to be larger than the provided IPv4 Pods CIDR. Defaults to 24. | *int | false |
| nodeCIDRMaskSizeIPv6 | NodeCIDRMaskSizeIPv6 is the mask size used to address the nodes within provided IPv6 Pods CIDR. It has to be larger than the provided IPv6 Pods CIDR. Defaults to 64. | *int | false |

[Back to Group](#v1beta3)

### ContainerRuntimeConfig

ContainerRuntimeConfig

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| containerd | Containerd related configurations | *[ContainerRuntimeContainerd](#containerruntimecontainerd) | false |

[Back to Group](#v1beta3)

### ContainerRuntimeContainerd

ContainerRuntimeContainerd defines docker container runtime

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| registries | A map of registries to use to render configs and mirrors for containerd registries | map[string][ContainerdRegistry](#containerdregistry) | false |
| deviceOwnershipFromSecurityContext | Enable or disable device_ownership_from_security_context containerd CRI config. Default to true. | *bool | false |
| sandboxImage | SandboxImage defines the pause image used by containerd's CRI plugin (e.g., \"registry.k8s.io/pause:3.10\") | string | false |

[Back to Group](#v1beta3)

### ContainerdRegistry

ContainerdRegistry defines endpoints and security for given container registry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| mirrors | List of registry mirrors to use | []string | false |
| overridePath | Configure override_path | bool | false |
| tlsConfig | TLSConfig for the registry | *[ContainerdTLSConfig](#containerdtlsconfig) | false |
| auth | Registry authentication | *[ContainerdRegistryAuthConfig](#containerdregistryauthconfig) | false |

[Back to Group](#v1beta3)

### ContainerdRegistryAuthConfig

Containerd per-registry credentials config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| username |  | string | false |
| password |  | string | false |
| auth |  | string | false |
| identityToken |  | string | false |

[Back to Group](#v1beta3)

### ContainerdTLSConfig

Configures containerd TLS for a registry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| insecureSkipVerify | Don't validate remote TLS certificate | bool | false |

[Back to Group](#v1beta3)

### ControlPlaneComponentConfig



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| flags | Flags is a set of additional flags that will be passed to the control plane component. KubeOne internally configures some flags that are eseeential for the cluster to work. Those flags set by KubeOne will be merged with the ones specified in the configuration. In case of conflict the value provided by the user will be used. Usage of `feature-gates` is not allowed here, use `FeatureGates` field instead. IMPORTANT: Use of these flags is at the user's own risk, as KubeOne does not provide support for issues caused by invalid values and configurations. | map[string]string | false |
| featureGates | FeatureGates is a map of additional feature gates that will be passed on to the control plane component. KubeOne internally configures some feature gates that are eseeential for the cluster to work. Those feature gates set by KubeOne will be merged with the ones specified in the configuration. In case of conflict the value provided by the user will be used. IMPORTANT: Use of these featureGates is at the user's own risk, as KubeOne does not provide support for issues caused by invalid values and configurations. | map[string]bool | false |

[Back to Group](#v1beta3)

### ControlPlaneComponents



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| controllerManager | ControllerManagerConfig configures the Kubernetes Controller Manager | *[ControlPlaneComponentConfig](#controlplanecomponentconfig) | false |
| scheduler | Scheduler configures the Kubernetes Scheduler | *[ControlPlaneComponentConfig](#controlplanecomponentconfig) | false |
| apiServer | APIServer configures the Kubernetes API Server | *[ControlPlaneComponentConfig](#controlplanecomponentconfig) | false |

[Back to Group](#v1beta3)

### ControlPlaneConfig

ControlPlaneConfig defines control plane nodes

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| hosts | Hosts array of all control plane hosts. | [][HostConfig](#hostconfig) | true |

[Back to Group](#v1beta3)

### CoreDNS



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| replicas |  | *int32 | false |
| deployPodDisruptionBudget |  | *bool | false |
| imageRepository | ImageRepository allows users to specify the image registry to be used for CoreDNS. Kubeadm automatically appends `/coredns` at the end, so it's not necessary to specify it. By default it's empty, which means it'll be defaulted based on kubeadm defaults and if overwriteRegistry feature is used. ImageRepository has the highest priority, meaning that it'll override overwriteRegistry if specified. | string | false |

[Back to Group](#v1beta3)

### DNSConfig

DNSConfig contains a machine's DNS configuration

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| servers | Servers | []string | true |

[Back to Group](#v1beta3)

### DigitalOceanSpec

DigitalOceanSpec defines the DigitalOcean cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta3)

### DynamicAuditLog

DynamicAuditLog feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable Default value is false. | bool | false |

[Back to Group](#v1beta3)

### DynamicWorkerConfig

DynamicWorkerConfig describes a set of worker machines

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name | string | true |
| replicas | Replicas | *int | true |
| providerSpec | Config | [ProviderSpec](#providerspec) | true |

[Back to Group](#v1beta3)

### EncryptionProviders

Encryption Providers feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | true |
| customEncryptionConfiguration | CustomEncryptionConfiguration | string | true |

[Back to Group](#v1beta3)

### EquinixMetalSpec

EquinixMetalSpec defines the Equinix Metal cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta3)

### ExternalCNISpec

ExternalCNISpec defines the external CNI plugin.
It's up to the user's responsibility to deploy the external CNI plugin manually or as an addon

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta3)

### Features

Features controls what features will be enabled on the cluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| coreDNS | CoreDNS | *[CoreDNS](#coredns) | false |
| podNodeSelector | PodNodeSelector | *[PodNodeSelector](#podnodeselector) | false |
| staticAuditLog | StaticAuditLog | *[StaticAuditLog](#staticauditlog) | false |
| dynamicAuditLog | DynamicAuditLog | *[DynamicAuditLog](#dynamicauditlog) | false |
| webhookAuditLog | WebhookAuditLog | *[WebhookAuditLog](#webhookauditlog) | false |
| metricsServer | MetricsServer | *[MetricsServer](#metricsserver) | false |
| openidConnect | OpenIDConnect | *[OpenIDConnect](#openidconnect) | false |
| encryptionProviders | Encryption Providers | *[EncryptionProviders](#encryptionproviders) | false |
| nodeLocalDNS | NodeLocalDNS config | *[NodeLocalDNS](#nodelocaldns) | false |

[Back to Group](#v1beta3)

### GCESpec

GCESpec defines the GCE cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta3)

### HelmAuth



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| username | Username for chart repository authentication. | string | false |
| password | Password for chart repository authentication. | string | false |

[Back to Group](#v1beta3)

### HelmRelease



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| chart | Chart is [CHART] part of the `helm upgrade [RELEASE] [CHART]` command. | string | true |
| repoURL | RepoURL is a chart repository URL where to locate the requested chart. | string | false |
| chartURL | ChartURL is a direct chart URL location. | string | false |
| version | Version is --version flag of the `helm upgrade` command. Specify the exact chart version to use. If this is not specified, the latest version is used. | string | false |
| releaseName | ReleaseName is [RELEASE] part of the `helm upgrade [RELEASE] [CHART]` command. Empty is defaulted to chart. | string | false |
| namespace | Namespace is --namespace flag of the `helm upgrade` command. A namespace to use for a release. | string | true |
| wait | Wait is --wait flag of the `helm install` command. | bool | false |
| timeout | WaitTimeout --timeout flag of the `helm install` command. Default to 5m | metav1.Duration | false |
| insecure | Insecure enables insecure TLS connection when fetching the chart | bool | false |
| values | Values provide optional overrides of the helm values. | [][HelmValues](#helmvalues) | false |
| auth | Auth is used for chart repository authentication. | *[HelmAuth](#helmauth) | false |

[Back to Group](#v1beta3)

### HelmValues

HelmValues configure inputs to `helm upgrade --install` command analog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| valuesFile | ValuesFile is an optional path on the local file system containing helm values to override. An analog of --values flag of the `helm upgrade` command. | string | false |
| inline | Inline is optionally used as a convenient way to provide short user input overrides to the helm upgrade process. Is written to a temporary file and used as an analog of the `helm upgrade --values=/tmp/inline-helm-values-XXX` command. | [json.RawMessage](https://golang.org/pkg/encoding/json/#RawMessage) | false |

[Back to Group](#v1beta3)

### HetznerSpec

HetznerSpec defines the Hetzner cloud provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| networkID | NetworkID | string | false |

[Back to Group](#v1beta3)

### HostConfig

HostConfig describes a single control plane or worker node.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| publicAddress | PublicAddress is externally accessible IP address from public internet. | string | true |
| ipv6Addresses | IPv6Addresses is a list of IPv6 addresses for the node. Only the first IPv6 address will be announced to the Kubernetes control plane. It is a list because you can request lots of IPv6 addresses (for example in case you want to assign one address per service). | []string | true |
| privateAddress | PrivateAddress is internal RFC-1918 IP address. | string | true |
| sshPort | SSHPort is port to connect ssh to. Default value is 22. | int | false |
| sshUsername | SSHUsername is system login name. Default value is \"root\". | string | false |
| sshPrivateKeyFile | SSHPrivateKeyFile is path to the file with PRIVATE AND CLEANTEXT ssh key. Default value is \"\". | string | false |
| sshCertFile | SSHCertFile is path to the file with the certificate of the private key. Default value is \"\". | string | false |
| sshHostPublicKey | SSHHostPublicKey if not empty, will be used to verify remote host public key | []byte | false |
| sshAgentSocket | SSHAgentSocket path (or reference to the environment) to the SSH agent unix domain socket. Default value is \"env:SSH_AUTH_SOCK\". | string | false |
| bastion | Bastion is an IP or hostname of the bastion (or jump) host to connect to. Default value is \"\". | string | false |
| bastionPort | BastionPort is SSH port to use when connecting to the bastion if it's configured in .Bastion. Default value is 22. | int | false |
| bastionUser | BastionUser is system login name to use when connecting to bastion host. Default value is \"root\". | string | false |
| bastionHostPublicKey | BastionHostPublicKey if not empty, will be used to verify bastion SSH public key | []byte | false |
| bastionPrivateKeyFile | BastionPrivateKeyFile is path to the file with PRIVATE AND CLEANTEXT ssh key. Default value is \"\". | string | false |
| hostname | Hostname is the hostname(1) of the host. Default value is populated at the runtime via running `hostname -f` command over ssh. | string | false |
| isLeader | IsLeader indicates this host as a session leader. Default value is populated at the runtime. | bool | false |
| taints | Taints are taints applied to nodes. Those taints are only applied when the node is being provisioned. If not provided (i.e. nil) for control plane nodes, it defaults to TaintEffectNoSchedule with key\n    node-role.kubernetes.io/control-plane\nExplicitly empty (i.e. []corev1.Taint{}) means no taints will be applied (this is default for worker nodes). | [][corev1.Taint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#taint-v1-core) | false |
| labels | Labels to be used to apply (or remove, with minus symbol suffix, see more kubectl help label) labels to/from node | map[string]string | false |
| annotations | Annotations to be used to apply (or remove, with minus symbol suffix, see more kubectl help annotate) annotations to/from node | map[string]string | false |
| kubelet | Kubelet | [KubeletConfig](#kubeletconfig) | false |
| operatingSystem | OperatingSystem information, can be populated at the runtime. | OperatingSystemName | false |

[Back to Group](#v1beta3)

### IPTables

IPTables

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta3)

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

[Back to Group](#v1beta3)

### KubeOneCluster

KubeOneCluster is KubeOne Cluster API Schema

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name is the name of the cluster. | string | true |
| controlPlane | ControlPlane describes the control plane nodes and how to access them. | [ControlPlaneConfig](#controlplaneconfig) | true |
| apiEndpoint | APIEndpoint are pairs of address and port used to communicate with the Kubernetes API. | [APIEndpoint](#apiendpoint) | true |
| cloudProvider | CloudProvider configures the cloud provider specific features. | [CloudProviderSpec](#cloudproviderspec) | true |
| kubeletConfig | KubeletConfig used to generate cluster's KubeletConfiguration that will be used along with kubeadm | [KubeletConfig](#kubeletconfig) | false |
| versions | Versions defines which Kubernetes version will be installed. | [VersionConfig](#versionconfig) | true |
| containerRuntime | ContainerRuntime defines which container runtime will be installed | [ContainerRuntimeConfig](#containerruntimeconfig) | false |
| clusterNetwork | ClusterNetwork configures the in-cluster networking. | [ClusterNetworkConfig](#clusternetworkconfig) | false |
| proxy | Proxy configures proxy used while installing Kubernetes and by the Docker daemon. | [ProxyConfig](#proxyconfig) | false |
| staticWorkers | StaticWorkers describes the worker nodes that are managed by KubeOne/kubeadm. | [StaticWorkersConfig](#staticworkersconfig) | false |
| dynamicWorkers | DynamicWorkers describes the worker nodes that are managed by Kubermatic machine-controller/Cluster-API. | [][DynamicWorkerConfig](#dynamicworkerconfig) | false |
| machineController | MachineController configures the Kubermatic machine-controller component. | *[MachineControllerConfig](#machinecontrollerconfig) | false |
| operatingSystemManager | OperatingSystemManager configures the Kubermatic operating-system-manager component. | *[OperatingSystemManagerConfig](#operatingsystemmanagerconfig) | false |
| caBundle | CABundle PEM encoded global CA.\n\nDeprecated: Use CertificateAuthorithyConfig instead. Will be overriten by certificateAuthority.bundle if set. | string | false |
| certificateAuthority | CertificateAuthority configures Central Authority certificate. | [CertificateAuthorithyConfig](#certificateauthorithyconfig) | false |
| features | Features enables and configures additional cluster features. | [Features](#features) | false |
| addons | Addons are used to deploy additional manifests. | *[Addons](#addons) | false |
| systemPackages | SystemPackages configure kubeone behaviour regarding OS packages. | *[SystemPackages](#systempackages) | false |
| registryConfiguration | RegistryConfiguration configures how Docker images are pulled from an image registry | *[RegistryConfiguration](#registryconfiguration) | false |
| loggingConfig | LoggingConfig configures the Kubelet's log rotation | [LoggingConfig](#loggingconfig) | false |
| tlsCipherSuites | TLSCipherSuites allows to configure TLS cipher suites for different components. See https://pkg.go.dev/crypto/tls#pkg-constants for possible values. | [TLSCipherSuites](#tlsciphersuites) | true |
| controlPlaneComponents | ControlPlaneComponents configures the Kubernetes control plane components | *[ControlPlaneComponents](#controlplanecomponents) | false |

[Back to Group](#v1beta3)

### KubeProxyConfig

KubeProxyConfig defines configured kube-proxy mode, default is iptables mode

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| skipInstallation | SkipInstallation will skip the installation of kube-proxy default value is false | bool | true |
| ipvs | IPVS config | *[IPVSConfig](#ipvsconfig) | true |
| iptables | IPTables config | *[IPTables](#iptables) | true |

[Back to Group](#v1beta3)

### KubeletConfig

KubeletConfig provides some kubelet configuration options

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| systemReserved | SystemReserved configure --system-reserved command-line flag of the kubelet. See more at: https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/ | map[string]string | false |
| kubeReserved | KubeReserved configure --kube-reserved command-line flag of the kubelet. See more at: https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/ | map[string]string | false |
| evictionHard | EvictionHard configure --eviction-hard command-line flag of the kubelet. See more at: https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/ | map[string]string | false |
| maxPods | MaxPods configures maximum number of pods per node. If not provided, default value provided by kubelet will be used (max. 110 pods per node) | *int32 | false |
| podPidsLimit | PodPidsLimit configures the maximum number of processes running in a Pod If not provided, default value provided by kubelet will be used -1 See more about pid-limiting at: https://kubernetes.io/docs/concepts/policy/pid-limiting/ | *int64 | false |
| imageGCHighThresholdPercent | ImageGCHighThresholdPercent is the percent of disk usage after which image garbage collection is always run. The percent is calculated by dividing this field value by 100, so this field must be between 0 and 100, inclusive. When specified, the value must be greater than imageGCLowThresholdPercent. Default: 85 | *int32 | false |
| imageGCLowThresholdPercent | ImageGCLowThresholdPercent is the percent of disk usage before which image garbage collection is never run. Lowest disk usage to garbage collect to. The percent is calculated by dividing this field value by 100, so the field value must be between 0 and 100, inclusive. When specified, the value must be less than imageGCHighThresholdPercent. Default: 80 | *int32 | false |
| imageMinimumGCAge | ImageMinimumGCAge is the minimum age for an unused image before it is garbage collected. Default: \"2m\" | metav1.Duration | false |
| imageMaximumGCAge | ImageMaximumGCAge is the maximum age an image can be unused before it is garbage collected. The default of this field is \"0s\", which disables this field--meaning images won't be garbage collected based on being unused for too long. Default: \"0s\" (disabled) | metav1.Duration | false |

[Back to Group](#v1beta3)

### KubevirtSpec

KubevirtSpec defines the Kubevirt provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| infraNamespace | InfraNamespace is the namespace that KubeVirt provider will use to create and manage resources in the infra cluster, such as VirtualMachines, VirtualMachineInstances, etc... | string | true |
| zoneAndRegionEnabled | ZoneAndRegionEnabled indicates if need to get Region and zone labels from the cloud provider | bool | false |
| loadBalancerEnabled | LoadBalancerEnabled indicates if the ccm should create and manage the clusters load balancers. | bool | false |

[Back to Group](#v1beta3)

### LoggingConfig

LoggingConfig configures the Kubelet's log rotation

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| containerLogMaxSize | ContainerLogMaxSize configures the maximum size of container log file before it is rotated See more at: https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/ | string | false |
| containerLogMaxFiles | ContainerLogMaxFiles configures the maximum number of container log files that can be present for a container See more at: https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/ | int32 | false |

[Back to Group](#v1beta3)

### MachineControllerConfig

MachineControllerConfig configures kubermatic machine-controller deployment

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| deploy | Deploy | bool | false |

[Back to Group](#v1beta3)

### MetricsServer

MetricsServer feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable deployment of metrics-server. Default value is true. | bool | false |

[Back to Group](#v1beta3)

### NodeLocalDNS



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| deploy | Deploy is enabled by default | bool | false |

[Back to Group](#v1beta3)

### NoneSpec

NoneSpec defines a none provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta3)

### NutanixSpec

NutanixSpec defines the Nutanix provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta3)

### OpenIDConnect

OpenIDConnect feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |
| config | Config | [OpenIDConnectConfig](#openidconnectconfig) | true |

[Back to Group](#v1beta3)

### OpenIDConnectConfig

OpenIDConnectConfig config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| issuerUrl | IssuerURL | string | true |
| clientId | ClientID | string | false |
| usernameClaim | UsernameClaim | string | false |
| usernamePrefix | UsernamePrefix. The value `-` can be used to disable all prefixing. | string | false |
| groupsClaim | GroupsClaim | string | false |
| groupsPrefix | GroupsPrefix. The value `-` can be used to disable all prefixing. | string | false |
| requiredClaim | RequiredClaim | string | true |
| signingAlgs | SigningAlgs | string | false |
| caFile | CAFile | string | true |

[Back to Group](#v1beta3)

### OpenstackSpec

OpenstackSpec defines the Openstack provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta3)

### OperatingSystemManagerConfig

OperatingSystemManagerConfig configures kubermatic operating-system-manager deployment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| deploy | Deploy | bool | false |
| enableNonRootDeviceOwnership | EnableNonRootDeviceOwnership enables the non-root device ownership feature in the container runtime. | bool | false |

[Back to Group](#v1beta3)

### PodNodeSelector

PodNodeSelector feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |
| config | Config | [PodNodeSelectorConfig](#podnodeselectorconfig) | true |

[Back to Group](#v1beta3)

### PodNodeSelectorConfig

PodNodeSelectorConfig config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| configFilePath | ConfigFilePath is a path on the local file system to the PodNodeSelector configuration file. ConfigFilePath is a required field. More info: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#podnodeselector | string | true |

[Back to Group](#v1beta3)

### ProviderSpec

ProviderSpec describes a worker node

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| cloudProviderSpec | CloudProviderSpec | [json.RawMessage](https://golang.org/pkg/encoding/json/#RawMessage) | true |
| annotations | Annotations set MachineDeployment.ObjectMeta.Annotations | map[string]string | false |
| nodeAnnotations | NodeAnnotations set MachineDeployment.Spec.Template.Spec.ObjectMeta.Annotations as a way to annotate resulting Nodes | map[string]string | false |
| machineObjectAnnotations | MachineObjectAnnotations set MachineDeployment.Spec.Template.Metadata.Annotations as a way to annotate resulting Machine objects. Those annotations are not propagated to Node objects. If you want to annotate resulting Nodes as well, see NodeAnnotations | map[string]string | false |
| labels | Labels | map[string]string | false |
| taints | Taints | [][corev1.Taint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#taint-v1-core) | false |
| sshPublicKeys | SSHPublicKeys | []string | false |
| operatingSystem | OperatingSystem | string | true |
| operatingSystemSpec | OperatingSystemSpec | [json.RawMessage](https://golang.org/pkg/encoding/json/#RawMessage) | false |
| network | Network | *[ProviderStaticNetworkConfig](#providerstaticnetworkconfig) | false |
| overwriteCloudConfig | OverwriteCloudConfig | *string | false |

[Back to Group](#v1beta3)

### ProviderStaticNetworkConfig

ProviderStaticNetworkConfig contains a machine's static network configuration

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| cidr | CIDR | string | true |
| gateway | Gateway | string | true |
| dns | DNS | [DNSConfig](#dnsconfig) | true |
| ipFamily | IPFamily | IPFamily | true |

[Back to Group](#v1beta3)

### ProxyConfig

ProxyConfig configures proxy for the Docker daemon and is used by KubeOne scripts

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| http | HTTP | string | false |
| https | HTTPS | string | false |
| noProxy | NoProxy | string | false |

[Back to Group](#v1beta3)

### RegistryConfiguration

RegistryConfiguration controls how images used for components deployed by
KubeOne and kubeadm are pulled from an image registry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| overwriteRegistry | OverwriteRegistry specifies a custom Docker registry which will be used for all images required for KubeOne and kubeadm. This also applies to addons deployed by KubeOne. This field doesn't modify the user/organization part of the image. For example, if OverwriteRegistry is set to 127.0.0.1:5000/example, image called calico/cni would translate to 127.0.0.1:5000/example/calico/cni. Default: \"\" | string | false |
| insecureRegistry | InsecureRegistry configures Docker to threat the registry specified in OverwriteRegistry as an insecure registry. This is also propagated to the worker nodes managed by machine-controller and/or KubeOne. | bool | false |

[Back to Group](#v1beta3)

### StaticAuditLog

StaticAuditLog feature flag

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable | bool | false |
| config | Config | [StaticAuditLogConfig](#staticauditlogconfig) | true |

[Back to Group](#v1beta3)

### StaticAuditLogConfig

StaticAuditLogConfig config

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| policyFilePath | PolicyFilePath is a path on local file system to the audit policy manifest which defines what events should be recorded and what data they should include. PolicyFilePath is a required field. More info: https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#audit-policy | string | true |
| logPath | LogPath is path on control plane instances where audit log files are stored. Default value is /var/log/kubernetes/audit.log | string | false |
| logMaxAge | LogMaxAge is maximum number of days to retain old audit log files. Default value is 30 | int | false |
| logMaxBackup | LogMaxBackup is maximum number of audit log files to retain. Default value is 3. | int | false |
| logMaxSize | LogMaxSize is maximum size in megabytes of audit log file before it gets rotated. Default value is 100. | int | false |

[Back to Group](#v1beta3)

### StaticWorkersConfig

StaticWorkersConfig defines static worker nodes provisioned by KubeOne and kubeadm

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| hosts | Hosts | [][HostConfig](#hostconfig) | false |

[Back to Group](#v1beta3)

### SystemPackages

SystemPackages controls configurations of APT/YUM

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| configureRepositories | ConfigureRepositories (true by default) is a flag to control automatic configuration of kubeadm / docker repositories. | bool | false |

[Back to Group](#v1beta3)

### TLSCipherSuites



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| apiServer | APIServer is a list of TLS cipher suites to use in kube-apiserver. | []string | false |
| etcd | Etcd is a list of TLS cipher suites to use in etcd. | []string | false |
| kubelet | Kubelet is a list of TLS cipher suites to use in kubelet. | []string | false |

[Back to Group](#v1beta3)

### VMwareCloudDirectorSpec

VMwareCloudDirectorSpec defines the VMware Cloud Director provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| vApp | VApp is the name of vApp for VMs. | string | false |
| storageProfile | StorageProfile is the name of storage profile to be used for disks. | string | true |

[Back to Group](#v1beta3)

### VersionConfig

VersionConfig describes the versions of components that are installed on the machines

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kubernetes |  | string | true |

[Back to Group](#v1beta3)

### VsphereSpec

VsphereSpec defines the vSphere provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#v1beta3)

### WeaveNetSpec

WeaveNetSpec defines the WeaveNet CNI plugin

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| encrypted | Encrypted | bool | false |

[Back to Group](#v1beta3)

### WebHookAuditLogBatchConfig



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| bufferSize | BufferSize defines the number of events to buffer before batching. If the rate of incoming events overflows the buffer, events are dropped. | int | false |
| maxSize | MaxSize defines the maximum number of events in one batch. | int | false |
| maxWait | MaxWait defines the maximum amount of time to wait before unconditionally batching events in the queue. | metav1.Duration | false |
| throttle | Throttle defines throttle configuration options. | [WebHookAuditLogThrottleConfig](#webhookauditlogthrottleconfig) | false |

[Back to Group](#v1beta3)

### WebHookAuditLogThrottleConfig



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| disable | Disable disables webhook throttling. Defaults to false, which corresponds to kube-apiservers default of enabling throttling. | bool | false |
| burst | Burst defines the maximum number of batches generated at the same moment if the allowed QPS was underutilized previously. | int | false |
| QPS | QPS defines the maximum average number of batches generated per second. | float32 | false |

[Back to Group](#v1beta3)

### WebHookAuditLogTruncateConfig



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable enables webhook truncating to support limiting the size of events. Defaults to false. | bool | false |
| maxBatchSize | MaxBatchSize defines the maximum size in bytes of the batch sent to the underlying backend. | int | false |
| maxEventSize | MaxEventSize defines the maximum size in bytes of the audit event sent to the underlying backend. | int | false |

[Back to Group](#v1beta3)

### WebhookAuditLog



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enable | Enable Default value is false. | bool | false |
| config | Config | [WebhookAuditLogConfig](#webhookauditlogconfig) | true |

[Back to Group](#v1beta3)

### WebhookAuditLogConfig



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| policyFilePath | PolicyFilePath is a path on local file system to the audit policy manifest which defines what events should be recorded and what data they should include. PolicyFilePath is a required field. More info: https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#audit-policy | string | true |
| configFilePath | ConfigFilePath is a path on local file system to a kubeconfig formatted file that defines how kube-apiserver can connect to the audit webhook. ConfigFilePath is a required field. More info: https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/#webhook-backend | string | true |
| initialBackOff | InitialBackOff defines the amount of time to wait before retrying the first failed request. Defaults to 10s. | metav1.Duration | false |
| mode | Mode defines the strategy for sending audit events. Blocking indicates sending events should block server responses. Batch causes the backend to buffer and write events asynchronously. Known modes are batch,blocking,blocking-strict. Defaults to batch. | string | false |
| version | Version defines API group and version used for serializing audit events written to webhook. Defaults to audit.k8s.io/v1 | string | false |
| batch | Batch defines settings for controlling event batching. Only applicable if webhook mode is set to batch. | [WebHookAuditLogBatchConfig](#webhookauditlogbatchconfig) | false |
| truncate | Truncate defines settings for controlling event truncation. | [WebHookAuditLogTruncateConfig](#webhookauditlogtruncateconfig) | false |

[Back to Group](#v1beta3)
