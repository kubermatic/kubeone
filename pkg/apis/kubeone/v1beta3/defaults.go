/*
Copyright 2024 The KubeOne Authors.

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

package v1beta3

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"

	"k8c.io/kubeone/pkg/templates/kubernetesconfigs"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

const (
	// DefaultPodSubnet defines the default subnet used by pods
	DefaultPodSubnet = "10.244.0.0/16"
	// DefaultServiceSubnet defines the default subnet used by services
	DefaultServiceSubnet = "10.96.0.0/12"
	// DefaultServiceDNS defines the default DNS domain name used by services
	DefaultServiceDNS = "cluster.local"
	// DefaultNodePortRange defines the default NodePort range
	DefaultNodePortRange = "30000-32767"
	// DefaultStaticNoProxy defined static NoProxy
	DefaultStaticNoProxy = "127.0.0.1/8,localhost"
	// DefaultCanalMTU defines default VXLAN MTU for Canal CNI
	DefaultCanalMTU = 1450
)

const (
	// DefaultPodSubnetIPv6 is the default network range from which IPv6 POD networks are allocated.
	DefaultPodSubnetIPv6 = "fd01::/48"
	// DefaultServiceSubnetIPv6 is the default network range from which IPv6 service VIPs are allocated.
	DefaultServiceSubnetIPv6 = "fd02::/120"
	// DefaultNodeCIDRMaskSizeIPv4 is the default mask size used to address the nodes within provided IPv4 Pods CIDR.
	DefaultNodeCIDRMaskSizeIPv4 = 24
	// DefaultNodeCIDRMaskSizeIPv6 is the default mask size used to address the nodes within provided IPv6 Pods CIDR.
	DefaultNodeCIDRMaskSizeIPv6 = 64
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_KubeOneCluster(obj *KubeOneCluster) {
	SetDefaults_Hosts(obj)
	SetDefaults_CloudProvider(obj)
	SetDefaults_APIEndpoints(obj)
	SetDefaults_Versions(obj)
	SetDefaults_ContainerRuntime(obj)
	SetDefaults_ClusterNetwork(obj)
	SetDefaults_Proxy(obj)
	SetDefaults_MachineController(obj)
	SetDefaults_OperatingSystemManager(obj)
	SetDefaults_Addons(obj)
	SetDefaults_SystemPackages(obj)
	SetDefaults_Features(obj)
	SetDefaults_CloudConfig(obj)
	SetDefaults_TLSCipherSuites(obj)
}

func SetDefaults_CloudProvider(obj *KubeOneCluster) {
	gteKube129Condition, _ := semver.NewConstraint(">= 1.29")
	actualVer, err := semver.NewVersion(obj.Versions.Kubernetes)
	if err != nil {
		return
	}

	// if kubernetes version is 1.29+
	// and cloud provider is configured
	if gteKube129Condition.Check(actualVer) && obj.CloudProvider.None == nil {
		// and cloud provider is NOT VMwareCloudDirector, to prevent kubelet --cloud-provider=external situation where
		// there will be no CCM (VMwareCloudDirector have no CCM) to initialize the Node
		if obj.CloudProvider.VMwareCloudDirector == nil {
			obj.CloudProvider.External = true
		}
	}
}

func SetDefaults_CloudConfig(obj *KubeOneCluster) {
	if obj.CloudProvider.AWS != nil && obj.CloudProvider.External {
		if obj.CloudProvider.CloudConfig == "" {
			obj.CloudProvider.CloudConfig = defaultAWSCCMCloudConfig(obj.Name, obj.ClusterNetwork.IPFamily)
		}
	}
}

func SetDefaults_Hosts(obj *KubeOneCluster) {
	// No hosts, so skip defaulting
	if len(obj.ControlPlane.Hosts) == 0 {
		return
	}

	setDefaultLeader := true

	// Define a unique ID for each host
	for idx := range obj.ControlPlane.Hosts {
		if setDefaultLeader && obj.ControlPlane.Hosts[idx].IsLeader {
			// override setting default leader, as explicit leader already
			// defined
			setDefaultLeader = false
		}
		obj.ControlPlane.Hosts[idx].ID = idx
		defaultHostConfig(&obj.ControlPlane.Hosts[idx])
		if obj.ControlPlane.Hosts[idx].Taints == nil {
			obj.ControlPlane.Hosts[idx].Taints = append(obj.ControlPlane.Hosts[idx].Taints, corev1.Taint{
				Effect: corev1.TaintEffectNoSchedule,
				Key:    "node-role.kubernetes.io/control-plane",
			})
		}
	}
	if setDefaultLeader {
		// In absence of explicitly defined leader set the first host to be the
		// default leader
		obj.ControlPlane.Hosts[0].IsLeader = true
	}

	for idx := range obj.StaticWorkers.Hosts {
		// continue assigning IDs after control plane hosts. This way every node gets a unique ID regardless of the different host slices
		obj.StaticWorkers.Hosts[idx].ID = idx + len(obj.ControlPlane.Hosts)
		defaultHostConfig(&obj.StaticWorkers.Hosts[idx])
		if obj.StaticWorkers.Hosts[idx].Taints == nil {
			obj.StaticWorkers.Hosts[idx].Taints = []corev1.Taint{}
		}
	}
}

func SetDefaults_APIEndpoints(obj *KubeOneCluster) {
	obj.APIEndpoint.Port = defaults(obj.APIEndpoint.Port, 6443)
}

func SetDefaults_Versions(obj *KubeOneCluster) {
	// The cluster provisioning fails if there is a leading "v" in the version
	obj.Versions.Kubernetes = strings.TrimPrefix(obj.Versions.Kubernetes, "v")
}

func SetDefaults_ContainerRuntime(obj *KubeOneCluster) {
	if obj.ContainerRuntime.Containerd == nil {
		obj.ContainerRuntime.Containerd = &ContainerRuntimeContainerd{}
	}
	if obj.ContainerRuntime.Containerd.DeviceOwnershipFromSecurityContext == nil {
		obj.ContainerRuntime.Containerd.DeviceOwnershipFromSecurityContext = ptr.To(true)
	}
}

func SetDefaults_ClusterNetwork(obj *KubeOneCluster) {
	if obj.ClusterNetwork.IPFamily == "" {
		obj.ClusterNetwork.IPFamily = IPFamilyIPv4
	}
	switch obj.ClusterNetwork.IPFamily {
	case IPFamilyIPv4:
		obj.ClusterNetwork.PodSubnet = defaults(obj.ClusterNetwork.PodSubnet, DefaultPodSubnet)
		obj.ClusterNetwork.ServiceSubnet = defaults(obj.ClusterNetwork.ServiceSubnet, DefaultServiceSubnet)
		obj.ClusterNetwork.NodeCIDRMaskSizeIPv4 = defaults(obj.ClusterNetwork.NodeCIDRMaskSizeIPv4, ptr.To(DefaultNodeCIDRMaskSizeIPv4))
	case IPFamilyIPv6:
		obj.ClusterNetwork.PodSubnetIPv6 = defaults(obj.ClusterNetwork.PodSubnetIPv6, DefaultPodSubnetIPv6)
		obj.ClusterNetwork.ServiceSubnetIPv6 = defaults(obj.ClusterNetwork.ServiceSubnetIPv6, DefaultServiceSubnetIPv6)
		obj.ClusterNetwork.NodeCIDRMaskSizeIPv6 = defaults(obj.ClusterNetwork.NodeCIDRMaskSizeIPv6, ptr.To(DefaultNodeCIDRMaskSizeIPv6))
	case IPFamilyIPv4IPv6, IPFamilyIPv6IPv4:
		obj.ClusterNetwork.PodSubnet = defaults(obj.ClusterNetwork.PodSubnet, DefaultPodSubnet)
		obj.ClusterNetwork.ServiceSubnet = defaults(obj.ClusterNetwork.ServiceSubnet, DefaultServiceSubnet)
		obj.ClusterNetwork.PodSubnetIPv6 = defaults(obj.ClusterNetwork.PodSubnetIPv6, DefaultPodSubnetIPv6)
		obj.ClusterNetwork.ServiceSubnetIPv6 = defaults(obj.ClusterNetwork.ServiceSubnetIPv6, DefaultServiceSubnetIPv6)
		obj.ClusterNetwork.NodeCIDRMaskSizeIPv4 = defaults(obj.ClusterNetwork.NodeCIDRMaskSizeIPv4, ptr.To(DefaultNodeCIDRMaskSizeIPv4))
		obj.ClusterNetwork.NodeCIDRMaskSizeIPv6 = defaults(obj.ClusterNetwork.NodeCIDRMaskSizeIPv6, ptr.To(DefaultNodeCIDRMaskSizeIPv6))
	}

	obj.ClusterNetwork.ServiceDomainName = defaults(obj.ClusterNetwork.ServiceDomainName, DefaultServiceDNS)
	obj.ClusterNetwork.NodePortRange = defaults(obj.ClusterNetwork.NodePortRange, DefaultNodePortRange)

	defaultCanal := &CanalSpec{MTU: DefaultCanalMTU}
	switch {
	case obj.CloudProvider.AWS != nil:
		defaultCanal.MTU = defaults(defaultCanal.MTU, 8951) // 9001 AWS Jumbo Frame - 50 VXLAN bytes
	case obj.CloudProvider.GCE != nil:
		defaultCanal.MTU = defaults(defaultCanal.MTU, 1410) // GCE specific 1460 bytes - 50 VXLAN bytes
	case obj.CloudProvider.Hetzner != nil:
		defaultCanal.MTU = defaults(defaultCanal.MTU, 1400) // Hetzner specific 1450 bytes - 50 VXLAN bytes
	case obj.CloudProvider.Openstack != nil:
		defaultCanal.MTU = defaults(defaultCanal.MTU, 1400) // Openstack specific 1450 bytes - 50 VXLAN bytes
	}

	if obj.ClusterNetwork.CNI == nil {
		obj.ClusterNetwork.CNI = &CNI{
			Canal: defaultCanal,
		}
	}
	if obj.ClusterNetwork.CNI.Canal != nil && obj.ClusterNetwork.CNI.Canal.MTU == 0 {
		obj.ClusterNetwork.CNI.Canal.MTU = defaultCanal.MTU
	}

	if obj.ClusterNetwork.CNI.Cilium != nil && obj.ClusterNetwork.CNI.Cilium.KubeProxyReplacement == "" {
		obj.ClusterNetwork.CNI.Cilium.KubeProxyReplacement = "disabled"
	}
}

func SetDefaults_Proxy(obj *KubeOneCluster) {
	if obj.Proxy.HTTP == "" && obj.Proxy.HTTPS == "" {
		return
	}
	noproxy := []string{
		DefaultStaticNoProxy,
		obj.ClusterNetwork.ServiceDomainName,
		obj.ClusterNetwork.PodSubnet,
		obj.ClusterNetwork.ServiceSubnet,
	}
	if obj.Proxy.NoProxy != "" {
		noproxy = append(noproxy, obj.Proxy.NoProxy)
	}
	obj.Proxy.NoProxy = strings.Join(noproxy, ",")
}

func SetDefaults_MachineController(obj *KubeOneCluster) {
	if obj.MachineController == nil {
		obj.MachineController = &MachineControllerConfig{
			Deploy: obj.CloudProvider.None == nil,
		}
	}
}

func SetDefaults_OperatingSystemManager(obj *KubeOneCluster) {
	if obj.OperatingSystemManager == nil {
		obj.OperatingSystemManager = &OperatingSystemManagerConfig{
			Deploy: obj.MachineController.Deploy,
		}
	}
}

func SetDefaults_Addons(obj *KubeOneCluster) {
	if obj.Addons == nil {
		obj.Addons = &Addons{}
	}

	for _, ad := range obj.Addons.Addons {
		if ad.HelmRelease != nil {
			if ad.HelmRelease.ReleaseName == "" {
				ad.HelmRelease.ReleaseName = ad.HelmRelease.Chart
			}

			if ad.HelmRelease.WaitTimeout.Duration == 0 {
				ad.HelmRelease.WaitTimeout.Duration = time.Minute * 5
			}
		}
	}
}

func SetDefaults_SystemPackages(obj *KubeOneCluster) {
	if obj.SystemPackages == nil {
		obj.SystemPackages = &SystemPackages{
			ConfigureRepositories: true,
		}
	}
}

func SetDefaults_Features(obj *KubeOneCluster) {
	if obj.Features.CoreDNS == nil {
		obj.Features.CoreDNS = &CoreDNS{}
	}
	if obj.Features.CoreDNS.Replicas == nil {
		obj.Features.CoreDNS.Replicas = ptr.To(int32(2))
	}
	if obj.Features.CoreDNS.DeployPodDisruptionBudget == nil {
		obj.Features.CoreDNS.DeployPodDisruptionBudget = ptr.To(true)
	}

	if obj.Features.MetricsServer == nil {
		obj.Features.MetricsServer = &MetricsServer{
			Enable: true,
		}
	}
	if obj.Features.StaticAuditLog != nil && obj.Features.StaticAuditLog.Enable {
		defaultStaticAuditLogConfig(&obj.Features.StaticAuditLog.Config)
	}
	if obj.Features.OpenIDConnect != nil && obj.Features.OpenIDConnect.Enable {
		defaultOpenIDConnect(&obj.Features.OpenIDConnect.Config)
	}
	if obj.Features.NodeLocalDNS == nil {
		obj.Features.NodeLocalDNS = &NodeLocalDNS{
			Deploy: true,
		}
	}
}

func SetDefaults_TLSCipherSuites(obj *KubeOneCluster) {
	if obj.TLSCipherSuites.APIServer == nil {
		obj.TLSCipherSuites.APIServer = kubernetesconfigs.APIServerDefaultTLSCipherSuites()
	}

	if obj.TLSCipherSuites.Etcd == nil {
		obj.TLSCipherSuites.Etcd = kubernetesconfigs.TLSCipherSuites(tls.CipherSuites())
	}

	if obj.TLSCipherSuites.Kubelet == nil {
		obj.TLSCipherSuites.Kubelet = kubernetesconfigs.DefaultTLSCipherSuites()
	}
}

func defaultOpenIDConnect(config *OpenIDConnectConfig) {
	config.ClientID = defaults(config.ClientID, "kubernetes")
	config.UsernameClaim = defaults(config.UsernameClaim, "sub")
	config.UsernamePrefix = defaults(config.UsernamePrefix, "oidc:")
	config.GroupsClaim = defaults(config.GroupsClaim, "groups")
	config.GroupsPrefix = defaults(config.GroupsPrefix, "oidc:")
	config.SigningAlgs = defaults(config.SigningAlgs, "RS256")
}

func defaultStaticAuditLogConfig(obj *StaticAuditLogConfig) {
	obj.LogPath = defaults(obj.LogPath, "/var/log/kubernetes/audit.log")
	obj.LogMaxAge = defaults(obj.LogMaxAge, 30)
	obj.LogMaxBackup = defaults(obj.LogMaxBackup, 3)
	obj.LogMaxSize = defaults(obj.LogMaxSize, 100)
}

func defaultHostConfig(obj *HostConfig) {
	if len(obj.PublicAddress) == 0 && len(obj.PrivateAddress) > 0 {
		obj.PublicAddress = obj.PrivateAddress
	}
	if len(obj.PrivateAddress) == 0 && len(obj.PublicAddress) > 0 {
		obj.PrivateAddress = obj.PublicAddress
	}
	if obj.SSHPrivateKeyFile == "" {
		obj.SSHAgentSocket = defaults(obj.SSHAgentSocket, "env:SSH_AUTH_SOCK")
	}
	obj.SSHUsername = defaults(obj.SSHUsername, "root")
	obj.SSHPort = defaults(obj.SSHPort, 22)
	obj.BastionPort = defaults(obj.BastionPort, 22)
	obj.BastionUser = defaults(obj.BastionUser, obj.SSHUsername)
}

func defaultAWSCCMCloudConfig(name string, ipFamily IPFamily) string {
	lines := []string{
		"[global]",
		fmt.Sprintf("KubernetesClusterID=%q", name),
	}

	switch ipFamily {
	case IPFamilyIPv4:
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv4"))
	case IPFamilyIPv6:
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv6"))
	case IPFamilyIPv4IPv6:
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv4"))
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv6"))
	case IPFamilyIPv6IPv4:
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv6"))
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv4"))
	}

	return strings.Join(lines, "\n")
}

func defaults[T comparable](input, defaultValue T) T {
	var zero T

	if input != zero {
		return input
	}

	return defaultValue
}
