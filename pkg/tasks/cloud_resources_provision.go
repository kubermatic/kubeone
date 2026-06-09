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

package tasks

import (
	"fmt"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubeonescheme "k8c.io/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1beta3 "k8c.io/kubeone/pkg/apis/kubeone/v1beta3"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/provisioner"
	"k8c.io/kubeone/pkg/state"
)

func WithFindControlPlane(t Tasks) Tasks {
	return t.append(
		Task{
			Description: "Find Hetzner load balancer",
			Predicate:   isHetznerControlPlaneEnabled,
			Fn:          lookupHetznerLoadBalancer,
		},
		Task{
			Description: "Find Hetzner VMs",
			Predicate:   isHetznerControlPlaneEnabled,
			Fn:          lookupHetznerVMs,
		},
		Task{
			Description: "Find OpenStack load balancer",
			Predicate:   isOpenstackControlPlaneEnabled,
			Fn:          lookupOpenstackLoadBalancer,
		},
		Task{
			Description: "Find OpenStack VMs",
			Predicate:   isOpenstackControlPlaneEnabled,
			Fn:          lookupOpenstackVMs,
		},
		Task{
			Operation: "defaulting cluster hosts",
			Predicate: func(s *state.State) bool { return len(s.Cluster.ControlPlane.NodeSets) != 0 },
			Fn:        defaultCluster,
		},
	).append(
		WithHostnameOS(nil)...,
	)
}

func WithEnsureControlPlane(steps Tasks, cluster *kubeoneapi.KubeOneCluster) (Tasks, error) {
	clusterName := cluster.Name
	nodeSet := cluster.ControlPlane.NodeSets
	kubeletVersion := cluster.Versions.Kubernetes

	switch {
	case cluster.CloudProvider.Hetzner != nil:
		hetznerCAPIMachines, err := generateHetznerControlPlaneMachines(clusterName, nodeSet, kubeletVersion)
		if err != nil {
			return nil, err
		}

		steps = steps.
			append(Task{
				Description: "Ensure Hetzner load balancer",
				Predicate:   isHetznerControlPlaneEnabled,
				Fn:          ensureHetznerLoadBalancer,
			}).
			append(generateHetznerControlPlaneTasks(hetznerCAPIMachines)...)
	case cluster.CloudProvider.Openstack != nil:
		openstackCAPIMachines, err := generateOpenstackControlPlaneMachines(clusterName, nodeSet, kubeletVersion)
		if err != nil {
			return nil, err
		}

		steps = steps.
			append(generateOpenstackControlPlaneTasks(openstackCAPIMachines)...).
			append(
				Task{
					Description: "Find OpenStack load balancer",
					Predicate:   isOpenstackControlPlaneEnabled,
					Fn:          lookupOpenstackLoadBalancer,
				},
			).
			append(Task{
				Description: "Register OpenStack LB members",
				Predicate:   isOpenstackControlPlaneEnabled,
				Fn:          ensureOpenstackLBMembers,
			})
	}

	return steps.
		append(Task{
			Operation: "defaulting cluster hosts",
			Predicate: func(s *state.State) bool { return len(s.Cluster.ControlPlane.NodeSets) != 0 },
			Fn:        defaultCluster,
		}), nil
}

func defaultCluster(st *state.State) error {
	v1beta3Cluster := kubeonev1beta3.NewKubeOneCluster()
	if err := kubeonescheme.Scheme.Convert(st.Cluster, v1beta3Cluster, nil); err != nil {
		return fail.Config(err, fmt.Sprintf("converting internal to %s object", v1beta3Cluster.GroupVersionKind()))
	}

	// run defauling again, to populate Hosts
	kubeonescheme.Scheme.Default(v1beta3Cluster)

	if err := kubeonescheme.Scheme.Convert(v1beta3Cluster, st.Cluster, nil); err != nil {
		return fail.Config(err, fmt.Sprintf("converting %s to internal object", v1beta3Cluster.GroupVersionKind()))
	}

	return nil
}

func hostConfigsFromMachines(machines []provisioner.Machine, nodeSets []kubeoneapi.NodeSet) []kubeoneapi.HostConfig {
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
