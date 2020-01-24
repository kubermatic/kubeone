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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
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
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_KubeOneCluster(obj *KubeOneCluster) {
	SetDefaults_Hosts(obj)
	SetDefaults_APIEndpoints(obj)
	SetDefaults_ClusterNetwork(obj)
	SetDefaults_MachineController(obj)
	SetDefaults_SystemPackages(obj)
	SetDefaults_Features(obj)
}

func SetDefaults_Hosts(obj *KubeOneCluster) {
	// No hosts, so skip defaulting
	if len(obj.Hosts) == 0 {
		return
	}

	// Set first host to be the leader
	obj.Hosts[0].IsLeader = true

	// Define a unique ID for each host
	for idx := range obj.Hosts {
		obj.Hosts[idx].ID = idx
		defaultHostConfig(&obj.Hosts[idx])
	}
}

func SetDefaults_APIEndpoints(obj *KubeOneCluster) {
	// If no API endpoint is provided, assume the public address is an endpoint
	if len(obj.APIEndpoint.Host) == 0 {
		if len(obj.Hosts) == 0 {
			// No hosts, so can't default to the first one
			return
		}
		obj.APIEndpoint.Host = obj.Hosts[0].PublicAddress
	}
	if obj.APIEndpoint.Port == 0 {
		obj.APIEndpoint.Port = 6443
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
	if obj.ClusterNetwork.CNI == nil {
		obj.ClusterNetwork.CNI = &CNI{
			Provider: CNIProviderCanal,
		}
	}
}

func SetDefaults_MachineController(obj *KubeOneCluster) {
	if obj.MachineController == nil {
		obj.MachineController = &MachineControllerConfig{
			Deploy: true,
		}
	}

	if obj.MachineController.Provider == "" {
		obj.MachineController.Provider = obj.CloudProvider.Name
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
	if obj.Features.MetricsServer == nil {
		obj.Features.MetricsServer = &MetricsServer{
			Enable: true,
		}
	}
	if obj.Features.StaticAuditLog != nil && obj.Features.StaticAuditLog.Enable {
		defaultStaticAuditLogConfig(&obj.Features.StaticAuditLog.Config)
	}
	if obj.Features.Backup != nil && obj.Features.Backup.Enable {
		defaultBackupConfig(&obj.Features.Backup.Config)
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

func defaultBackupConfig(obj *BackupConfig) {
	if len(obj.ResticPassword) == 0 {
		obj.ResticPassword = rand.String(10)
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
