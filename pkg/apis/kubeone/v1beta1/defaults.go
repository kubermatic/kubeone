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
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	// DefaultVXLanMTU defines default VXLAN MTU for Canal CNI
	DefaultCanalMTU = 1450
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_KubeOneCluster(obj *KubeOneCluster) {
	SetDefaults_Hosts(obj)
	SetDefaults_APIEndpoints(obj)
	SetDefaults_Versions(obj)
	SetDefaults_ContainerRuntime(obj)
	SetDefaults_ClusterNetwork(obj)
	SetDefaults_Proxy(obj)
	SetDefaults_MachineController(obj)
	SetDefaults_SystemPackages(obj)
	SetDefaults_AssetConfiguration(obj)
	SetDefaults_Features(obj)
	SetDefaults_Addons(obj)
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
			obj.ControlPlane.Hosts[idx].Taints = []corev1.Taint{
				{
					Effect: corev1.TaintEffectNoSchedule,
					Key:    "node-role.kubernetes.io/master",
				},
			}
		}
	}
	if setDefaultLeader {
		// In absence of explicitly defined leader set the first host to be the
		// default leader
		obj.ControlPlane.Hosts[0].IsLeader = true
	}

	for idx := range obj.StaticWorkers.Hosts {
		// continue assinging IDs after control plane hosts. This way every node gets a unique ID regardless of the different host slices
		obj.StaticWorkers.Hosts[idx].ID = idx + len(obj.ControlPlane.Hosts)
		defaultHostConfig(&obj.StaticWorkers.Hosts[idx])
		if obj.StaticWorkers.Hosts[idx].Taints == nil {
			obj.StaticWorkers.Hosts[idx].Taints = []corev1.Taint{}
		}
	}
}

func SetDefaults_APIEndpoints(obj *KubeOneCluster) {
	// If no API endpoint is provided, assume the public address is an endpoint
	if len(obj.APIEndpoint.Host) == 0 {
		if len(obj.ControlPlane.Hosts) == 0 {
			// No hosts, so can't default to the first one
			return
		}
		obj.APIEndpoint.Host = obj.ControlPlane.Hosts[0].PublicAddress
	}
	if obj.APIEndpoint.Port == 0 {
		obj.APIEndpoint.Port = 6443
	}
}

func SetDefaults_Versions(obj *KubeOneCluster) {
	// The cluster provisioning fails if there is a leading "v" in the version
	obj.Versions.Kubernetes = strings.TrimPrefix(obj.Versions.Kubernetes, "v")
}

func SetDefaults_ContainerRuntime(obj *KubeOneCluster) {
	switch {
	case obj.ContainerRuntime.Docker != nil:
	case obj.ContainerRuntime.Containerd != nil:
	default:
		obj.ContainerRuntime = ContainerRuntimeConfig{
			Docker: &ContainerRuntimeDocker{},
		}
	}
}

func SetDefaults_ClusterNetwork(obj *KubeOneCluster) {
	if len(obj.ClusterNetwork.PodSubnet) == 0 {
		obj.ClusterNetwork.PodSubnet = DefaultPodSubnet
	}
	if len(obj.ClusterNetwork.ServiceSubnet) == 0 {
		obj.ClusterNetwork.ServiceSubnet = DefaultServiceSubnet
	}
	if len(obj.ClusterNetwork.ServiceDomainName) == 0 {
		obj.ClusterNetwork.ServiceDomainName = DefaultServiceDNS
	}
	if len(obj.ClusterNetwork.NodePortRange) == 0 {
		obj.ClusterNetwork.NodePortRange = DefaultNodePortRange
	}

	defaultCanal := &CanalSpec{MTU: DefaultCanalMTU}
	switch {
	case obj.CloudProvider.AWS != nil:
		defaultCanal.MTU = 8951 // 9001 AWS Jumbo Frame - 50 VXLAN bytes
	case obj.CloudProvider.GCE != nil:
		defaultCanal.MTU = 1410 // GCE specific 1460 bytes - 50 VXLAN bytes
	case obj.CloudProvider.Hetzner != nil:
		defaultCanal.MTU = 1400 // Hetzner specific 1450 bytes - 50 VXLAN bytes
	case obj.CloudProvider.Openstack != nil:
		defaultCanal.MTU = 1400 // Openstack specific 1450 bytes - 50 VXLAN bytes
	}

	if obj.ClusterNetwork.CNI == nil {
		obj.ClusterNetwork.CNI = &CNI{
			Canal: defaultCanal,
		}
	}
	if obj.ClusterNetwork.CNI.Canal != nil && obj.ClusterNetwork.CNI.Canal.MTU == 0 {
		obj.ClusterNetwork.CNI.Canal.MTU = defaultCanal.MTU
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
			Deploy: true,
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

func SetDefaults_AssetConfiguration(obj *KubeOneCluster) {
	if obj.RegistryConfiguration == nil || obj.RegistryConfiguration.OverwriteRegistry == "" {
		// We default AssetConfiguration only if RegistryConfiguration.OverwriteRegistry
		// is used
		return
	}

	if obj.AssetConfiguration.Kubernetes.ImageRepository == "" {
		obj.AssetConfiguration.Kubernetes.ImageRepository = obj.RegistryConfiguration.OverwriteRegistry
	}
	if obj.AssetConfiguration.CoreDNS.ImageRepository == "" {
		obj.AssetConfiguration.CoreDNS.ImageRepository = obj.RegistryConfiguration.OverwriteRegistry
	}
	if obj.AssetConfiguration.Etcd.ImageRepository == "" {
		obj.AssetConfiguration.Etcd.ImageRepository = obj.RegistryConfiguration.OverwriteRegistry
	}
	if obj.AssetConfiguration.MetricsServer.ImageRepository == "" {
		obj.AssetConfiguration.MetricsServer.ImageRepository = obj.RegistryConfiguration.OverwriteRegistry
	}
}

func SetDefaults_Features(obj *KubeOneCluster) {
	if obj.Features.MetricsServer == nil {
		obj.Features.MetricsServer = &MetricsServer{
			Enable: true,
		}
	}
	if obj.Features.StaticAuditLog != nil && obj.Features.StaticAuditLog.Enable {
		defaultStaticAuditLogConfig(&obj.Features.StaticAuditLog.Config)
	}
}

func SetDefaults_Addons(obj *KubeOneCluster) {
	if obj.Addons != nil && obj.Addons.Enable {
		if len(obj.Addons.Path) == 0 {
			obj.Addons.Path = "./addons"
		}
	}
}

func defaultStaticAuditLogConfig(obj *StaticAuditLogConfig) {
	if obj.LogPath == "" {
		obj.LogPath = "/var/log/kubernetes/audit.log"
	}
	if obj.LogMaxAge == 0 {
		obj.LogMaxAge = 30
	}
	if obj.LogMaxBackup == 0 {
		obj.LogMaxBackup = 3
	}
	if obj.LogMaxSize == 0 {
		obj.LogMaxSize = 100
	}
}

func defaultHostConfig(obj *HostConfig) {
	if len(obj.PublicAddress) == 0 && len(obj.PrivateAddress) > 0 {
		obj.PublicAddress = obj.PrivateAddress
	}
	if len(obj.PrivateAddress) == 0 && len(obj.PublicAddress) > 0 {
		obj.PrivateAddress = obj.PublicAddress
	}
	if len(obj.SSHPrivateKeyFile) == 0 && len(obj.SSHAgentSocket) == 0 {
		obj.SSHAgentSocket = "env:SSH_AUTH_SOCK"
	}
	if obj.SSHUsername == "" {
		obj.SSHUsername = "root"
	}
	if obj.SSHPort == 0 {
		obj.SSHPort = 22
	}
	if obj.BastionPort == 0 {
		obj.BastionPort = 22
	}
	if obj.BastionUser == "" {
		obj.BastionUser = obj.SSHUsername
	}
}
