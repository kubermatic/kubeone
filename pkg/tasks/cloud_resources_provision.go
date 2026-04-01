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
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"time"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubeonescheme "k8c.io/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1beta3 "k8c.io/kubeone/pkg/apis/kubeone/v1beta3"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/provisioner"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	hetznertypes "k8c.io/machine-controller/sdk/cloudprovider/hetzner"
	"k8c.io/machine-controller/sdk/jsonutil"
	"k8c.io/machine-controller/sdk/providerconfig"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func WithEnsureControlPlane(t Tasks) Tasks {
	return t.append(
		Task{
			Description: "Ensure Hetzner load balancer",
			Operation:   "ensure Hetzner load balancer",
			Predicate:   isHetznerControlPlaneEnabled,
			Fn:          ensureHetznerLoadBalancer,
		},
		Task{
			Description: "Ensure Hetzner control-plane VMs",
			Operation:   "ensure Hetzner control-plane VMs",
			Predicate:   isHetznerControlPlaneEnabled,
			Fn:          ensureHetznerControlPlaneVMs,
		},
		Task{
			Operation: "defaulting cluster hosts",
			Predicate: func(s *state.State) bool { return len(s.Cluster.ControlPlane.NodeSets) != 0 },
			Fn: func(s *state.State) error {
				v1beta3Cluster := kubeonev1beta3.NewKubeOneCluster()
				if err := kubeonescheme.Scheme.Convert(s.Cluster, v1beta3Cluster, nil); err != nil {
					return fail.Config(err, "converting internal to v1beta3 object")
				}

				// run defauling again, to populate Hosts
				kubeonescheme.Scheme.Default(v1beta3Cluster)

				if err := kubeonescheme.Scheme.Convert(v1beta3Cluster, s.Cluster, nil); err != nil {
					return fail.Config(err, fmt.Sprintf("converting %s to internal object", v1beta3Cluster.GroupVersionKind()))
				}

				return nil
			},
		},
	)
}

func isHetznerControlPlaneEnabled(s *state.State) bool {
	return s.Cluster.CloudProvider.Hetzner != nil && len(s.Cluster.ControlPlane.NodeSets) > 0
}

func ensureHetznerLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return err
	}

	hzclient := hcloud.NewClient(hcloud.WithToken(providerCreds[credentials.HetznerTokenKeyMC]))
	var networkID int64

	networkName := s.Cluster.CloudProvider.Hetzner.NetworkID
	ctx := context.Background()

	networks, _, err := hzclient.Network.List(ctx, hcloud.NetworkListOpts{
		Name: networkName,
	})
	if err != nil {
		return fail.Cloud(err, "hetzner", "listing networks")
	}

	if len(networks) == 0 {
		return fail.Cloud(fmt.Errorf("no network ID found with ID: %s", networkName), "hetzner", "looking up network")
	}

	networkID = networks[0].ID

	clusterLBName := s.Cluster.CloudProvider.Hetzner.ControlPlane.LoadBalancer.Name
	lbs, _, err := hzclient.LoadBalancer.List(ctx, hcloud.LoadBalancerListOpts{
		Name: clusterLBName,
	})
	if err != nil {
		return fail.Cloud(err, "hetzner", "listing loadbalancers")
	}

	var realLB *hcloud.LoadBalancer

	if len(lbs) > 0 {
		s.Logger.Infof("loadbalancer already exists with id: %d", lbs[0].ID)
		realLB = lbs[0]
	} else {
		s.Logger.Infof("no existing loadbalancer found, creating a new one")
		lb, err := createLoadBalancer(
			ctx,
			&hzclient.LoadBalancer,
			s.Cluster,
			networkID,
		)
		if err != nil {
			return fail.Cloud(err, "hetzner", "creating loadbalancer")
		}
		realLB = lb
	}

	s.Cluster.APIEndpoint.Host = realLB.PublicNet.IPv4.IP.String()
	s.Cluster.APIEndpoint.Port = 6443

	return nil
}

func createLoadBalancer(
	ctx context.Context,
	client hcloud.ILoadBalancerClient,
	cluster *kubeoneapi.KubeOneCluster,
	networkID int64,
) (*hcloud.LoadBalancer, error) {
	vmsLabelSelector := labels.SelectorFromSet(hetznerLabels(cluster.Name)).String()
	now := time.Now().UTC()
	timestamp := strconv.FormatInt(now.Unix(), 10)
	newLabels := map[string]string{
		"kubeone_own_since_timestamp": timestamp,
	}
	hzlbSpec := cluster.CloudProvider.Hetzner.ControlPlane.LoadBalancer
	labels := make(map[string]string)
	maps.Copy(labels, hzlbSpec.Labels)
	maps.Copy(labels, newLabels)

	createReq := hcloud.LoadBalancerCreateOpts{
		Name:             hzlbSpec.Name,
		LoadBalancerType: &hcloud.LoadBalancerType{Name: hzlbSpec.Type},
		Location:         &hcloud.Location{Name: hzlbSpec.Location},
		Labels:           labels,
		PublicInterface:  hzlbSpec.PublicIP,
		Services: []hcloud.LoadBalancerCreateOptsService{
			{
				Protocol:        hcloud.LoadBalancerServiceProtocolTCP,
				ListenPort:      new(6443),
				DestinationPort: new(6443),
			},
		},
		Targets: []hcloud.LoadBalancerCreateOptsTarget{
			{
				UsePrivateIP: new(true),
				Type:         hcloud.LoadBalancerTargetTypeLabelSelector,
				LabelSelector: hcloud.LoadBalancerCreateOptsTargetLabelSelector{
					Selector: vmsLabelSelector,
				},
			},
		},
		Network: &hcloud.Network{ID: networkID},
	}

	result, _, err := client.Create(ctx, createReq)
	if err != nil {
		return nil, err
	}

	return result.LoadBalancer, nil
}

func ensureHetznerControlPlaneVMs(s *state.State) error {
	capimachines, err := generateHetznerControlPlaneMachines(s.Cluster.Name, s.Cluster.ControlPlane.NodeSets, s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	provMachines, err := provisioner.FindOrCreateMachines(s.Context, capimachines, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = hostConfigsFromMachines(provMachines, s.Cluster.ControlPlane.NodeSets)

	return nil
}

func hetznerLabels(clusterName string) map[string]string {
	return map[string]string{
		"kubeone_cluster_name": clusterName,
		"kubeone_role":         "api",
	}
}

func generateHetznerControlPlaneMachines(clusterName string, nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	var machines []clusterv1alpha1.Machine

	for _, node := range nodeSet {
		timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
		labels := map[string]string{
			"kubeone_own_since_timestamp": timestamp,
			"kubeone_role":                "control-plane",
		}
		maps.Copy(labels, hetznerLabels(clusterName))

		if node.NodeSettings.Labels == nil {
			node.NodeSettings.Labels = map[string]string{}
		}
		maps.Copy(node.NodeSettings.Labels, labels)

		for idx := range node.Replicas {
			osSpecRaw, err := json.Marshal(node.OperatingSystemSpec)
			if err != nil {
				return nil, err
			}

			var hetznerConfig hetznertypes.RawConfig
			if err = jsonutil.StrictUnmarshal(node.CloudProviderSpec, &hetznerConfig); err != nil {
				return nil, fail.Config(err, "decode hetzner config")
			}

			if hetznerConfig.Labels == nil {
				hetznerConfig.Labels = map[string]string{}
			}

			maps.Copy(hetznerConfig.Labels, labels)

			hetznerSpec, err := json.Marshal(hetznerConfig)
			if err != nil {
				return nil, fail.Config(err, "marshaling cloud provider spec")
			}

			providerConfig := providerconfig.Config{
				SSHPublicKeys: node.SSH.PublicKeys,
				CloudProvider: providerconfig.CloudProviderHetzner,
				CloudProviderSpec: runtime.RawExtension{
					Raw: hetznerSpec,
				},
				OperatingSystem: providerconfig.OperatingSystem(node.OperatingSystem),
				OperatingSystemSpec: runtime.RawExtension{
					Raw: osSpecRaw,
				},
			}

			providerSpecRaw, err := json.Marshal(providerConfig)
			if err != nil {
				return nil, fail.Cloud(err, "hetzner", "json marshaling provider config")
			}

			name := fmt.Sprintf("%s-%s-%d", clusterName, node.Name, idx)
			machines = append(machines, clusterv1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					UID:  types.UID(name),
				},
				Spec: clusterv1alpha1.MachineSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:        name,
						Labels:      node.NodeSettings.Labels,
						Annotations: node.NodeSettings.Annotations,
					},
					Taints: node.NodeSettings.Taints,
					Versions: clusterv1alpha1.MachineVersionInfo{
						Kubelet: kubeletVersion,
					},
					ProviderSpec: clusterv1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{
							Raw: providerSpecRaw,
						},
					},
				},
			})
		}
	}

	return machines, nil
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
				IsLeader:             idx == 0,
			}

			hosts = append(hosts, host)
			idx++
		}
	}

	return hosts
}
