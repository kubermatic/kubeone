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

package gce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/cloudprovider"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/provisioner"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	gcetypes "k8c.io/machine-controller/sdk/cloudprovider/gce"
	"k8c.io/machine-controller/sdk/jsonutil"
	"k8c.io/machine-controller/sdk/providerconfig"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const apiServerPort = 6443

var (
	_ cloudprovider.ControlPlaneCloudProvider = &Provider{}
	_ cloudprovider.LoadBalancerProvider      = &Provider{}
)

func init() {
	cloudprovider.Register(&Provider{})
}

type Provider struct{}

func (p *Provider) Name() string { return "GCE" }

func (p *Provider) Enabled(s *state.State) bool {
	gce := s.Cluster.CloudProvider.GCE

	return gce != nil && gce.ControlPlane != nil && len(s.Cluster.ControlPlane.NodeSets) > 0
}

func (p *Provider) MatchesConfig(cluster *kubeoneapi.KubeOneCluster) bool {
	return cluster.CloudProvider.GCE != nil
}

func (p *Provider) HasLoadBalancer(s *state.State) bool {
	return p.Enabled(s)
}

func (p *Provider) GenerateMachines(clusterName string, nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	return generateGCEControlPlaneMachines(clusterName, nodeSet, kubeletVersion)
}

func (p *Provider) EnsureVM(s *state.State, capimachine clusterv1alpha1.Machine) error {
	provMachines, err := provisioner.FindOrCreateMachines(s.Context, []clusterv1alpha1.Machine{capimachine}, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, cloudprovider.HostConfigsFromMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

	return p.ensureTargetPoolMembership(s, capimachine)
}

func (p *Provider) LookupVMs(s *state.State) error {
	capimachines, err := p.GenerateMachines(
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

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, cloudprovider.HostConfigsFromMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

	return nil
}

// ensureTargetPoolMembership adds the just (re)discovered/created VM to the Target Pool.
//
// GCE Target Pools, unlike Hetzner's label-selector load balancer, require
// each member instance to be explicitly registered. EnsureLoadBalancer runs
// before any control-plane VM exists, so membership is synced here instead,
// once per VM, right after it's created/found.
func (p *Provider) ensureTargetPoolMembership(s *state.State, capimachine clusterv1alpha1.Machine) error {
	lb := s.Cluster.CloudProvider.GCE.ControlPlane.LoadBalancer

	pconfig, err := providerconfig.GetConfig(capimachine.Spec.ProviderSpec)
	if err != nil {
		return fail.Config(err, "reading provider config")
	}

	gceConfig, err := gcetypes.GetConfig(*pconfig)
	if err != nil {
		return fail.Config(err, "decode gce config")
	}

	projectID := gceConfig.ProjectID.Value
	if projectID == "" {
		projectID = lb.ProjectID
	}

	svc, err := newComputeService(s)
	if err != nil {
		return err
	}

	instanceSelfLink := fmt.Sprintf("projects/%s/zones/%s/instances/%s", projectID, gceConfig.Zone.Value, capimachine.Name)

	return ensureTargetPoolMember(s.Context, svc, projectID, lb.Region, loadBalancerName(s.Cluster.Name, lb.Name), instanceSelfLink)
}

func (p *Provider) EnsureLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	lb := s.Cluster.CloudProvider.GCE.ControlPlane.LoadBalancer

	svc, err := newComputeService(s)
	if err != nil {
		return err
	}

	name := loadBalancerName(s.Cluster.Name, lb.Name)
	network := lb.Network
	if network == "" {
		network = "global/networks/default"
	}

	if ensureErr := ensureFirewallRule(s.Context, svc, lb.ProjectID, name, network, controlPlaneTag(s.Cluster.Name)); ensureErr != nil {
		return ensureErr
	}

	if ensureErr := ensureTargetPool(s.Context, svc, lb.ProjectID, lb.Region, name); ensureErr != nil {
		return ensureErr
	}

	fr, err := ensureForwardingRule(s.Context, svc, lb.ProjectID, lb.Region, name)
	if err != nil {
		return err
	}

	s.Cluster.APIEndpoint.Host = fr.IPAddress
	s.Cluster.APIEndpoint.Port = apiServerPort

	return nil
}

func (p *Provider) LookupLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	lb := s.Cluster.CloudProvider.GCE.ControlPlane.LoadBalancer

	svc, err := newComputeService(s)
	if err != nil {
		return err
	}

	name := loadBalancerName(s.Cluster.Name, lb.Name)

	fr, err := svc.ForwardingRules.Get(lb.ProjectID, lb.Region, name).Context(s.Context).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "looking up forwarding rule %q", name)
	}

	s.Logger.Debugf("found forwarding rule %q with address: %s", name, fr.IPAddress)
	s.Cluster.APIEndpoint.Host = fr.IPAddress

	return nil
}

func newComputeService(s *state.State) (*compute.Service, error) {
	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return nil, err
	}

	svc, err := compute.NewService(s.Context, option.WithCredentialsJSON([]byte(providerCreds[credentials.GoogleServiceAccountKey])))
	if err != nil {
		return nil, fail.Cloud(err, "gce", "creating compute client")
	}

	return svc, nil
}

func loadBalancerName(clusterName, configuredName string) string {
	if configuredName != "" {
		return configuredName
	}

	return fmt.Sprintf("%s-kubeapi", clusterName)
}

func controlPlaneTag(clusterName string) string {
	return fmt.Sprintf("%s-control-plane", clusterName)
}

func isNotFound(err error) bool {
	var gerr *googleapi.Error

	return errors.As(err, &gerr) && gerr.Code == http.StatusNotFound
}

func ensureFirewallRule(ctx context.Context, svc *compute.Service, projectID, name, network, targetTag string) error {
	firewallName := fmt.Sprintf("%s-%d", name, apiServerPort)

	_, err := svc.Firewalls.Get(projectID, firewallName).Context(ctx).Do()
	if err == nil {
		return nil
	}
	if !isNotFound(err) {
		return fail.Cloud(err, "gce", "looking up firewall rule %q", firewallName)
	}

	_, err = svc.Firewalls.Insert(projectID, &compute.Firewall{
		Name:      firewallName,
		Network:   network,
		Direction: "INGRESS",
		TargetTags: []string{
			targetTag,
		},
		SourceRanges: []string{"0.0.0.0/0"},
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      []string{strconv.Itoa(apiServerPort)},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "creating firewall rule %q", firewallName)
	}

	return nil
}

func ensureTargetPool(ctx context.Context, svc *compute.Service, projectID, region, name string) error {
	_, err := svc.TargetPools.Get(projectID, region, name).Context(ctx).Do()
	if err == nil {
		return nil
	}
	if !isNotFound(err) {
		return fail.Cloud(err, "gce", "looking up target pool %q", name)
	}

	_, err = svc.TargetPools.Insert(projectID, region, &compute.TargetPool{
		Name: name,
	}).Context(ctx).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "creating target pool %q", name)
	}

	return nil
}

func ensureTargetPoolMember(ctx context.Context, svc *compute.Service, projectID, region, poolName, instanceSelfLink string) error {
	pool, err := svc.TargetPools.Get(projectID, region, poolName).Context(ctx).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "getting target pool %q", poolName)
	}

	for _, instance := range pool.Instances {
		if instance == instanceSelfLink {
			return nil
		}
	}

	_, err = svc.TargetPools.AddInstance(projectID, region, poolName, &compute.TargetPoolsAddInstanceRequest{
		Instances: []*compute.InstanceReference{
			{Instance: instanceSelfLink},
		},
	}).Context(ctx).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "adding instance %q to target pool %q", instanceSelfLink, poolName)
	}

	return nil
}

func ensureForwardingRule(ctx context.Context, svc *compute.Service, projectID, region, name string) (*compute.ForwardingRule, error) {
	fr, err := svc.ForwardingRules.Get(projectID, region, name).Context(ctx).Do()
	if err == nil {
		return fr, nil
	}
	if !isNotFound(err) {
		return nil, fail.Cloud(err, "gce", "looking up forwarding rule %q", name)
	}

	targetPoolSelfLink := fmt.Sprintf("projects/%s/regions/%s/targetPools/%s", projectID, region, name)

	_, err = svc.ForwardingRules.Insert(projectID, region, &compute.ForwardingRule{
		Name:                name,
		IPProtocol:          "TCP",
		PortRange:           fmt.Sprintf("%d-%d", apiServerPort, apiServerPort),
		Target:              targetPoolSelfLink,
		LoadBalancingScheme: "EXTERNAL",
	}).Context(ctx).Do()
	if err != nil {
		return nil, fail.Cloud(err, "gce", "creating forwarding rule %q", name)
	}

	fr, err = svc.ForwardingRules.Get(projectID, region, name).Context(ctx).Do()
	if err != nil {
		return nil, fail.Cloud(err, "gce", "reading created forwarding rule %q", name)
	}

	return fr, nil
}

func gceLabels(clusterName string) map[string]string {
	return map[string]string{
		"kubeone_cluster_name": clusterName,
		"kubeone_role":         "control-plane",
	}
}

func generateGCEControlPlaneMachines(clusterName string, nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	var machines []clusterv1alpha1.Machine

	for _, node := range nodeSet {
		timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
		nodeLabels := map[string]string{
			"kubeone_own_since_timestamp": timestamp,
		}
		maps.Copy(nodeLabels, gceLabels(clusterName))

		if node.NodeSettings.Labels == nil {
			node.NodeSettings.Labels = map[string]string{}
		}
		maps.Copy(node.NodeSettings.Labels, nodeLabels)

		for idx := range node.Replicas {
			osSpecRaw, err := json.Marshal(node.OperatingSystemSpec)
			if err != nil {
				return nil, err
			}

			var gceConfig gcetypes.RawConfig
			if err = jsonutil.StrictUnmarshal(node.CloudProviderSpec, &gceConfig); err != nil {
				return nil, fail.Config(err, "decode gce config")
			}

			if gceConfig.Labels == nil {
				gceConfig.Labels = map[string]string{}
			}
			maps.Copy(gceConfig.Labels, nodeLabels)

			// Links the VM to the :6443 firewall rule created by EnsureLoadBalancer.
			gceConfig.Tags = append(gceConfig.Tags, controlPlaneTag(clusterName))

			gceSpec, err := json.Marshal(gceConfig)
			if err != nil {
				return nil, fail.Config(err, "marshaling cloud provider spec")
			}

			providerConfig := providerconfig.Config{
				SSHPublicKeys: node.SSH.PublicKeys,
				CloudProvider: providerconfig.CloudProviderGoogle,
				CloudProviderSpec: runtime.RawExtension{
					Raw: gceSpec,
				},
				OperatingSystem: providerconfig.OperatingSystem(node.OperatingSystem),
				OperatingSystemSpec: runtime.RawExtension{
					Raw: osSpecRaw,
				},
			}

			providerSpecRaw, err := json.Marshal(providerConfig)
			if err != nil {
				return nil, fail.Cloud(err, "gce", "json marshaling provider config")
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
