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
	"fmt"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
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

	// EnsureVM finds or creates the VM of the given machine, which was
	// generated from the given nodeSet, and adds it to the control plane hosts.
	EnsureVM(*state.State, kubeoneapi.NodeSet, clusterv1alpha1.Machine) error

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

// Preflighter checks the existing cloud resources before the control plane is
// looked up or ensured, e.g. to refuse changes that can't be applied in place.
type Preflighter interface {
	Preflight(*state.State) error
}

type LoadBalancerProvider interface {
	HasLoadBalancer(*state.State) bool

	EnsureLoadBalancer(*state.State) error

	LookupLoadBalancer(*state.State) error

	CleanupLoadBalancer(*state.State) error
}

// LoadBalancerMemberProvider is implemented by providers whose control plane
// VMs must not be members of the load balancer while they join, e.g. because
// the VMs behind the load balancer route its address to themselves, so a
// joining node would reach itself instead of the cluster. Their EnsureVM adds
// at most the leader.
type LoadBalancerMemberProvider interface {
	// RemoveJoiningLoadBalancerMembers runs after the leader initialized the
	// cluster and before the followers join. It ensures the leader is a member
	// of the load balancer and the joining hosts are not, e.g. hosts that
	// were reset but kept their VMs.
	RemoveJoiningLoadBalancerMembers(s *state.State, joining []kubeoneapi.HostConfig) error

	// EnsureLoadBalancerMembers runs on every apply once the control plane
	// serves, after the followers joined, and adds all control plane VMs to
	// the load balancer.
	EnsureLoadBalancerMembers(*state.State) error
}

// MachineNodeSets returns the nodeSet each machine was generated from.
// GenerateMachines emits the replicas of each nodeSet in order, so the i-th
// machine belongs to the i-th replica slot across nodeSets.
func MachineNodeSets(machines []clusterv1alpha1.Machine, nodeSets []kubeoneapi.NodeSet) ([]kubeoneapi.NodeSet, error) {
	var out []kubeoneapi.NodeSet
	for _, nodeSet := range nodeSets {
		for range nodeSet.Replicas {
			out = append(out, nodeSet)
		}
	}

	if len(out) != len(machines) {
		return nil, fail.Config(fmt.Errorf("generated %d machines for %d control plane replicas", len(machines), len(out)), "mapping machines to nodeSets")
	}

	return out, nil
}

// HostConfigsFromMachines maps the given machines to hosts, assuming they are
// ordered as returned by GenerateMachines for the given nodeSets.
func HostConfigsFromMachines(machines []provisioner.Machine, nodeSets []kubeoneapi.NodeSet) []kubeoneapi.HostConfig {
	var hosts []kubeoneapi.HostConfig
	idx := 0

	for _, nodeSet := range nodeSets {
		for range nodeSet.Replicas {
			if idx >= len(machines) {
				break
			}

			hosts = append(hosts, HostConfigFromMachine(machines[idx], nodeSet))
			idx++
		}
	}

	return hosts
}

// HostConfigFromMachine maps a machine to a host with the SSH and node
// settings of the nodeSet the machine was generated from.
func HostConfigFromMachine(m provisioner.Machine, nodeSet kubeoneapi.NodeSet) kubeoneapi.HostConfig {
	sshUsername := nodeSet.SSH.Username
	if sshUsername == "" {
		sshUsername = "root"
	}

	return kubeoneapi.HostConfig{
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
}
