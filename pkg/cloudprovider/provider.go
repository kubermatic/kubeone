/*
Copyright 2026 The KubeOne Authors.

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

package cloudprovider

import (
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/provisioner"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
)

var controlPlaneProviders []ControlPlaneCloudProvider

func ControlPlaneProviders() []ControlPlaneCloudProvider {
	return controlPlaneProviders
}

func Register(p ControlPlaneCloudProvider) {
	controlPlaneProviders = append(controlPlaneProviders, p)
}

type ControlPlaneCloudProvider interface {
	Name() string

	Enabled(*state.State) bool

	MatchesConfig(*kubeoneapi.KubeOneCluster) bool

	GenerateMachines(clusterName string, nodes []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error)

	EnsureVM(*state.State, clusterv1alpha1.Machine) error

	LookupVMs(*state.State) error
}

// CleanupVMs generates the control plane machines for the given provider and
// destroys their instances at the cloud provider.
func CleanupVMs(p ControlPlaneCloudProvider, s *state.State) error {
	if prep, ok := p.(CleanupPreparer); ok {
		if err := prep.PrepareCleanup(s); err != nil {
			return err
		}
	}

	machines, err := p.GenerateMachines(
		s.Cluster.Name,
		s.Cluster.ControlPlane.NodeSets,
		s.Cluster.Versions.Kubernetes,
	)
	if err != nil {
		return err
	}

	return provisioner.CleanupMachines(s.Context, machines, s.Logger)
}

type CleanupPreparer interface {
	PrepareCleanup(*state.State) error
}

type LoadBalancerProvider interface {
	HasLoadBalancer(*state.State) bool

	EnsureLoadBalancer(*state.State) error

	LookupLoadBalancer(*state.State) error

	CleanupLoadBalancer(*state.State) error
}

func HostConfigsFromMachines(machines []provisioner.Machine, nodeSets []kubeoneapi.NodeSet) []kubeoneapi.HostConfig {
	var hosts []kubeoneapi.HostConfig
	idx := 0

	for _, nodeSet := range nodeSets {
		sshUsername := nodeSet.SSH.Username
		if sshUsername == "" {
			sshUsername = "root"
		}

		for range nodeSet.Replicas {
			if idx >= len(machines) {
				break
			}

			m := machines[idx]
			host := kubeoneapi.HostConfig{
				PublicAddress:        m.PublicAddress,
				PrivateAddress:       m.PrivateAddress,
				Hostname:             m.Hostname,
				SSHUsername:          sshUsername,
				SSHPort:              nodeSet.SSH.Port,
				SSHPrivateKeyFile:    nodeSet.SSH.PrivateKeyFile,
				SSHCertFile:          nodeSet.SSH.CertFile,
				SSHHostPublicKey:     nodeSet.SSH.HostPublicKey,
				SSHAgentSocket:       nodeSet.SSH.AgentSocket,
				Bastion:              nodeSet.SSH.Bastion,
				BastionPort:          nodeSet.SSH.BastionPort,
				BastionUser:          nodeSet.SSH.BastionUser,
				BastionHostPublicKey: nodeSet.SSH.BastionHostPublicKey,
				OperatingSystem:      nodeSet.OperatingSystem,
				Labels:               nodeSet.NodeSettings.Labels,
				Annotations:          nodeSet.NodeSettings.Annotations,
				Taints:               nodeSet.NodeSettings.Taints,
			}

			hosts = append(hosts, host)
			idx++
		}
	}

	return hosts
}
