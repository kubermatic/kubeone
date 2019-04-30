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

package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/apis/kubeone/v1alpha1"
	kubeonev1alpha1 "github.com/kubermatic/kubeone/pkg/apis/kubeone/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MigrateToKubeOneClusterAPI migrates the old API to the new KubeOneCluster API
func MigrateToKubeOneClusterAPI(oldConfigPath string) (*kubeonev1alpha1.KubeOneCluster, error) {
	oldConfig, err := loadClusterConfig(oldConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse the old config")
	}

	// Initialize the KubeOneCluster structure
	newConfig := &kubeonev1alpha1.KubeOneCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeOneCluster",
			APIVersion: "v1alpha1",
		},
		Name: oldConfig.Name,
	}

	// Append hosts
	newConfig.Hosts = []kubeonev1alpha1.HostConfig{}
	for _, oldHost := range oldConfig.Hosts {
		newHost := kubeonev1alpha1.HostConfig{
			PublicAddress:     oldHost.PublicAddress,
			PrivateAddress:    oldHost.PrivateAddress,
			SSHPort:           oldHost.SSHPort,
			SSHUsername:       oldHost.SSHUsername,
			SSHPrivateKeyFile: oldHost.SSHPrivateKeyFile,
			SSHAgentSocket:    oldHost.SSHAgentSocket,
		}
		newConfig.Hosts = append(newConfig.Hosts, newHost)
	}

	// Create the initial API endpoint from the old API server configuration
	newConfig.APIEndpoints = []kubeonev1alpha1.APIEndpoint{
		{
			Host: oldConfig.APIServer.Address,
		},
	}

	// Populate the cloud provider settings
	newConfig.CloudProvider = kubeonev1alpha1.CloudProviderSpec{
		Name:        kubeonev1alpha1.CloudProviderName(oldConfig.Provider.Name),
		External:    oldConfig.Provider.External,
		CloudConfig: oldConfig.Provider.CloudConfig,
	}

	// Populate the Kubernetes version
	newConfig.Versions = kubeonev1alpha1.VersionConfig{
		Kubernetes: oldConfig.Versions.Kubernetes,
	}

	// Populate the ClusterNetwork structure
	newConfig.ClusterNetwork = kubeonev1alpha1.ClusterNetworkConfig{
		PodSubnet:     oldConfig.Network.PodSubnetVal,
		ServiceSubnet: oldConfig.Network.ServiceSubnetVal,
		NodePortRange: oldConfig.Network.NodePortRangeVal,
	}

	// Populate the proxy configuration
	newConfig.Proxy = kubeonev1alpha1.ProxyConfig{
		HTTP:    oldConfig.Proxy.HTTPProxy,
		HTTPS:   oldConfig.Proxy.HTTPSProxy,
		NoProxy: oldConfig.Proxy.NoProxy,
	}

	// Populate the workers information
	newConfig.Workers = []v1alpha1.WorkerConfig{}
	for _, oldWorker := range oldConfig.Workers {
		oldCloudProviderSpec, err := json.Marshal(oldWorker.Config.CloudProviderSpec)
		if err != nil {
			return nil, errors.Errorf("unable to parse workers.Config.CloudProviderSpec for worker: %s: %v", oldWorker.Name, err)
		}
		oldOperatingSystemSpec, err := json.Marshal(oldWorker.Config.OperatingSystemSpec)
		if err != nil {
			return nil, errors.Errorf("unable to parse workers.Config.OperatingSystemSpec for worker: %s: %v", oldWorker.Name, err)
		}

		newWorker := kubeonev1alpha1.WorkerConfig{
			Name:     oldWorker.Name,
			Replicas: oldWorker.Replicas,
			Config: v1alpha1.ProviderSpec{
				CloudProviderSpec:   oldCloudProviderSpec,
				Labels:              oldWorker.Config.Labels,
				SSHPublicKeys:       oldWorker.Config.SSHPublicKeys,
				OperatingSystem:     oldWorker.Config.OperatingSystem,
				OperatingSystemSpec: oldOperatingSystemSpec,
			},
		}
		newConfig.Workers = append(newConfig.Workers, newWorker)
	}

	// Populate the machine-controller configuration
	var deployMachineController bool
	if oldConfig.MachineController.Deploy == nil || (oldConfig.MachineController.Deploy != nil && *oldConfig.MachineController.Deploy) {
		deployMachineController = true
	} else {
		deployMachineController = false
	}
	newConfig.MachineController = &kubeonev1alpha1.MachineControllerConfig{
		Deploy:   deployMachineController,
		Provider: kubeonev1alpha1.CloudProviderName(oldConfig.MachineController.Provider),
	}

	// Populate the features configuration
	if oldConfig.Features.PodSecurityPolicy.Enable != nil {
		newConfig.Features.PodSecurityPolicy = &v1alpha1.PodSecurityPolicy{
			Enable: *oldConfig.Features.PodSecurityPolicy.Enable,
		}
	}
	if oldConfig.Features.DynamicAuditLog.Enable != nil {
		newConfig.Features.DynamicAuditLog = &v1alpha1.DynamicAuditLog{
			Enable: *oldConfig.Features.DynamicAuditLog.Enable,
		}
	}
	if oldConfig.Features.MetricsServer.Enable != nil {
		newConfig.Features.MetricsServer = &v1alpha1.MetricsServer{
			Enable: *oldConfig.Features.MetricsServer.Enable,
		}
	}
	if oldConfig.Features.OpenIDConnect.Enable {
		newConfig.Features.OpenIDConnect = &v1alpha1.OpenIDConnect{
			Enable: *oldConfig.Features.MetricsServer.Enable,
			Config: v1alpha1.OpenIDConnectConfig{
				IssuerURL:      oldConfig.Features.OpenIDConnect.Config.IssuerURL,
				ClientID:       oldConfig.Features.OpenIDConnect.Config.ClientID,
				UsernameClaim:  oldConfig.Features.OpenIDConnect.Config.UsernameClaim,
				UsernamePrefix: oldConfig.Features.OpenIDConnect.Config.UsernamePrefix,
				GroupsClaim:    oldConfig.Features.OpenIDConnect.Config.GroupsClaim,
				GroupsPrefix:   oldConfig.Features.OpenIDConnect.Config.GroupsPrefix,
				RequiredClaim:  oldConfig.Features.OpenIDConnect.Config.RequiredClaim,
				SigningAlgs:    oldConfig.Features.OpenIDConnect.Config.SigningAlgs,
				CAFile:         oldConfig.Features.OpenIDConnect.Config.CAFile,
			},
		}
	}

	return newConfig, nil
}

func loadClusterConfig(oldConfigPath string) (*Cluster, error) {
	content, err := ioutil.ReadFile(oldConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	cluster := Cluster{}
	if err := yaml.Unmarshal(content, &cluster); err != nil {
		return nil, errors.Wrap(err, "failed to decode file as JSON")
	}

	return &cluster, nil
}
