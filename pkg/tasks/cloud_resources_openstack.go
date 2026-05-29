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
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	"github.com/gophercloud/gophercloud/pagination"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/provisioner"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	openstacktypes "k8c.io/machine-controller/sdk/cloudprovider/openstack"
	"k8c.io/machine-controller/sdk/jsonutil"
	"k8c.io/machine-controller/sdk/providerconfig"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func isOpenstackControlPlaneEnabled(s *state.State) bool {
	return s.Cluster.CloudProvider.Openstack != nil &&
		s.Cluster.CloudProvider.Openstack.ControlPlane != nil &&
		len(s.Cluster.ControlPlane.NodeSets) > 0
}

func lookupOpenstackVMs(s *state.State) error {
	capimachines, err := generateOpenstackControlPlaneMachines(
		s.Cluster.Name,
		s.Cluster.ControlPlane.NodeSets,
		s.Cluster.Versions.Kubernetes,
	)
	if err != nil {
		return err
	}

	provMachines, err := provisioner.FindMachines(s.Context, capimachines, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, hostConfigsFromMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

	return nil
}

func ensureOpenstackControlPlaneVM(s *state.State, capimachine clusterv1alpha1.Machine) error {
	provMachines, err := provisioner.FindOrCreateMachines(s.Context, []clusterv1alpha1.Machine{capimachine}, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, hostConfigsFromMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

	return nil
}

func generateOpenstackControlPlaneTasks(capimachines []clusterv1alpha1.Machine) Tasks {
	tasks := Tasks{}

	for _, machine := range capimachines {
		tasks = append(tasks,
			Task{
				Description: fmt.Sprintf("Ensure OpenStack control-plane %q VM", machine.Name),
				Operation:   fmt.Sprintf("ensure OpenStack control-plane %q VM", machine.Name),
				Predicate:   isOpenstackControlPlaneEnabled,
				Fn: func(s *state.State) error {
					return ensureOpenstackControlPlaneVM(s, machine)
				},
			},
		)
	}

	return tasks
}

func ensureOpenstackLBMembers(s *state.State) error {
	osCP := s.Cluster.CloudProvider.Openstack.ControlPlane

	lbClient, err := openstackLBClient(s)
	if err != nil {
		return err
	}

	poolID := osCP.LoadBalancer.PoolID
	if poolID == "" {
		// Discover pool from LB
		discoveredPoolID, oserr := discoverOpenstackLBPool(lbClient, osCP.LoadBalancer.Name)
		if oserr != nil {
			return oserr
		}
		poolID = discoveredPoolID
	}

	// List existing members
	existingMembers := map[string]bool{}
	err = pools.ListMembers(lbClient, poolID, pools.ListMembersOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		members, oserr := pools.ExtractMembers(page)
		if oserr != nil {
			return false, oserr
		}
		for _, m := range members {
			existingMembers[m.Address] = true
		}

		return true, nil
	})
	if err != nil {
		return fail.Cloud(err, "openstack", "listing LB pool members")
	}

	// Add missing members
	for _, host := range s.Cluster.ControlPlane.Hosts {
		addr := host.PrivateAddress
		if addr == "" {
			addr = host.PublicAddress
		}
		if addr == "" || existingMembers[addr] {
			continue
		}

		s.Logger.Infof("Adding control plane node %s (%s) to LB pool", host.Hostname, addr)
		protocolPort := 6443
		_, err := pools.CreateMember(lbClient, poolID, pools.CreateMemberOpts{
			Address:      addr,
			ProtocolPort: protocolPort,
			Name:         host.Hostname,
		}).Extract()
		if err != nil {
			return fail.Cloud(err, "openstack", "adding member %s to LB pool", addr)
		}
	}

	return nil
}

func openstackLBClient(s *state.State) (*gophercloud.ServiceClient, error) {
	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return nil, err
	}

	opts := gophercloud.AuthOptions{
		IdentityEndpoint: providerCreds["OS_AUTH_URL"],
		Username:         providerCreds["OS_USERNAME"],
		Password:         providerCreds["OS_PASSWORD"],
		DomainName:       providerCreds["OS_DOMAIN_NAME"],
		TenantName:       providerCreds["OS_TENANT_NAME"],
		TenantID:         providerCreds["OS_TENANT_ID"],
	}

	// Support application credentials
	if appCredID := providerCreds["OS_APPLICATION_CREDENTIAL_ID"]; appCredID != "" {
		opts.ApplicationCredentialID = appCredID
		opts.ApplicationCredentialSecret = providerCreds["OS_APPLICATION_CREDENTIAL_SECRET"]
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, fail.Cloud(err, "openstack", "authenticating")
	}

	region := providerCreds["OS_REGION_NAME"]
	lbClient, err := openstack.NewLoadBalancerV2(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		return nil, fail.Cloud(err, "openstack", "creating load balancer client")
	}

	return lbClient, nil
}

func discoverOpenstackLBPool(lbClient *gophercloud.ServiceClient, lbName string) (string, error) {
	var lbID string

	err := loadbalancers.List(lbClient, loadbalancers.ListOpts{Name: lbName}).EachPage(func(page pagination.Page) (bool, error) {
		lbs, err := loadbalancers.ExtractLoadBalancers(page)
		if err != nil {
			return false, err
		}
		if len(lbs) > 0 {
			lbID = lbs[0].ID
		}

		return true, nil
	})
	if err != nil {
		return "", fail.Cloud(err, "openstack", "listing load balancers")
	}

	if lbID == "" {
		return "", fail.Cloud(fmt.Errorf("no load balancer found with name: %s", lbName), "openstack", "looking up load balancer")
	}

	// Find pool associated with this LB
	var poolID string
	err = pools.List(lbClient, pools.ListOpts{LoadbalancerID: lbID}).EachPage(func(page pagination.Page) (bool, error) {
		poolList, oserr := pools.ExtractPools(page)
		if oserr != nil {
			return false, oserr
		}
		if len(poolList) > 0 {
			poolID = poolList[0].ID
		}

		return true, nil
	})
	if err != nil {
		return "", fail.Cloud(err, "openstack", "listing LB pools")
	}

	if poolID == "" {
		return "", fail.Cloud(fmt.Errorf("no pool found for load balancer: %s", lbName), "openstack", "looking up LB pool")
	}

	return poolID, nil
}

func openstackLabels(clusterName string) map[string]string {
	return map[string]string{
		"kubeone_cluster_name": clusterName,
		"kubeone_role":         "control-plane",
	}
}

func generateOpenstackControlPlaneMachines(clusterName string, nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	var machines []clusterv1alpha1.Machine

	for _, node := range nodeSet {
		timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
		nodeLabels := map[string]string{
			"kubeone_own_since_timestamp": timestamp,
		}
		maps.Copy(nodeLabels, openstackLabels(clusterName))

		if node.NodeSettings.Labels == nil {
			node.NodeSettings.Labels = map[string]string{}
		}
		maps.Copy(node.NodeSettings.Labels, nodeLabels)

		for idx := range node.Replicas {
			osSpecRaw, err := json.Marshal(node.OperatingSystemSpec)
			if err != nil {
				return nil, err
			}

			var osConfig openstacktypes.RawConfig
			if err = jsonutil.StrictUnmarshal(node.CloudProviderSpec, &osConfig); err != nil {
				return nil, fail.Config(err, "decode openstack config")
			}

			if osConfig.Tags == nil {
				osConfig.Tags = map[string]string{}
			}
			maps.Copy(osConfig.Tags, nodeLabels)

			openstackSpec, err := json.Marshal(osConfig)
			if err != nil {
				return nil, fail.Config(err, "marshaling cloud provider spec")
			}

			providerConfig := providerconfig.Config{
				SSHPublicKeys: node.SSH.PublicKeys,
				CloudProvider: providerconfig.CloudProviderOpenstack,
				CloudProviderSpec: runtime.RawExtension{
					Raw: openstackSpec,
				},
				OperatingSystem: providerconfig.OperatingSystem(node.OperatingSystem),
				OperatingSystemSpec: runtime.RawExtension{
					Raw: osSpecRaw,
				},
			}

			providerSpecRaw, err := json.Marshal(providerConfig)
			if err != nil {
				return nil, fail.Cloud(err, "openstack", "json marshaling provider config")
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
