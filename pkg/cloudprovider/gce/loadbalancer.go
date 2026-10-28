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

// The kube-apiserver load balancer is a backend service-based regional
// external passthrough Network Load Balancer:
//
//	address <name> (regional, EXTERNAL)
//	  └─ forwarding rule <name> (EXTERNAL, TCP 6443)
//	       └─ backend service <name> (regional, EXTERNAL, TCP) ── health check <name> (regional, TCP 6443)
//	            └─ zonal network endpoint group <name>-<zone> (GCE_VM_IP), one per control plane zone
//	                 └─ one endpoint per control plane VM
//
// plus the firewall rules <name>-6443 (clients) and <name>-6443-hc (health
// check probes), both targeting the control plane network tag.
//
// Zonal GCE_VM_IP network endpoint groups are used instead of unmanaged
// instance groups because a VM can only be in a single load-balanced instance
// group, while it can be an endpoint of several network endpoint groups. The
// GCE cloud controller manager uses instance groups for LoadBalancer Services.
//
// See:
//   - https://docs.cloud.google.com/load-balancing/docs/network/networklb-backend-service
//   - https://docs.cloud.google.com/load-balancing/docs/negs/zonal-neg-concepts
//   - https://docs.cloud.google.com/load-balancing/docs/health-check-concepts

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/compute/v1"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	gcetypes "k8c.io/machine-controller/sdk/cloudprovider/gce"
	"k8c.io/machine-controller/sdk/jsonutil"
	"k8c.io/machine-controller/sdk/providerconfig"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	apiFirewallSuffix         = "-6443"
	healthCheckFirewallSuffix = "-6443-hc"
	negTypeGCEVMIP            = "GCE_VM_IP"
	balancingModeConnection   = "CONNECTION"
	schemeExternal            = "EXTERNAL"
	protocolTCP               = "TCP"
)

// healthCheckSourceRanges are the health check probe ranges of regional
// backend service-based external passthrough Network Load Balancers. The docs
// list 209.85.152.0/22 only "for target pools", so it isn't allowed, see
// https://docs.cloud.google.com/load-balancing/docs/health-check-concepts
var healthCheckSourceRanges = []string{"35.191.0.0/16", "209.85.204.0/22"}

// lbConfig is the resolved load balancer configuration.
type lbConfig struct {
	ClusterName string
	ProjectID   string
	// ServiceAccountProjectID is the project of the service account, the
	// machine-controller driver creates the VMs there unless the nodeSet
	// sets a project.
	ServiceAccountProjectID string
	Region                  string
	Name                    string
	// Network of the control plane VMs and the firewall rules, as
	// projects/<project>/global/networks/<name>, set by resolveNetwork.
	Network             string
	AllowedSourceRanges []string
}

// owner is the description set on every load balancer resource KubeOne
// creates. Resources with another description are never changed or deleted.
func (c *lbConfig) owner() string {
	return OwnerDescription(c.ClusterName)
}

// OwnerDescription is the description of the load balancer resources and
// firewall rules KubeOne creates for the cluster.
func OwnerDescription(clusterName string) string {
	return "kubeone:" + clusterName
}

func (c *lbConfig) regionalPath(collection, name string) string {
	return fmt.Sprintf("projects/%s/regions/%s/%s/%s", c.ProjectID, c.Region, collection, name)
}

// firewallProject is the project of the network, the firewall rules of a
// Shared VPC network are in its host project.
func (c *lbConfig) firewallProject() string {
	parts := strings.Split(resourcePath(c.Network), "/")
	if len(parts) >= 2 && parts[0] == "projects" {
		return parts[1]
	}

	return c.ProjectID
}

func (c *lbConfig) negName(zone string) string {
	return c.Name + "-" + zone
}

func resolveLB(s *state.State) (*lbConfig, error) {
	lb := s.Cluster.CloudProvider.GCE.ControlPlane.LoadBalancer

	cfg := &lbConfig{
		ClusterName:         s.Cluster.Name,
		ProjectID:           lb.ProjectID,
		Region:              lb.Region,
		Name:                lb.Name,
		Network:             lb.Network,
		AllowedSourceRanges: lb.AllowedSourceRanges,
	}

	serviceAccount, err := serviceAccountJSON(s)
	if err != nil {
		return nil, err
	}

	if cfg.ServiceAccountProjectID, err = serviceAccountProjectID(serviceAccount); err != nil {
		return nil, err
	}

	if cfg.ProjectID == "" {
		cfg.ProjectID = cfg.ServiceAccountProjectID
	}

	switch {
	case cfg.ProjectID == "":
		return nil, fail.ConfigValidation(errors.New("no GCE project: set .cloudProvider.gce.controlPlane.loadBalancer.projectID or use a service account with project_id"))
	case cfg.Region == "":
		return nil, fail.ConfigValidation(errors.New("no GCE region: set .cloudProvider.gce.controlPlane.loadBalancer.region or a zone in the control plane nodeSets"))
	case cfg.Name == "":
		return nil, fail.ConfigValidation(errors.New(".cloudProvider.gce.controlPlane.loadBalancer.name is empty"))
	}

	return cfg, nil
}

func (p *Provider) EnsureLoadBalancer(s *state.State) error {
	cfg, err := resolveLB(s)
	if err != nil {
		return err
	}

	svc, err := newComputeService(s)
	if err != nil {
		return err
	}

	// Fail before anything is created if a VM would be created in another
	// project than the load balancer
	machines, err := p.GenerateMachines(s.Cluster.Name, s.Cluster.ControlPlane.NodeSets, s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}
	for _, machine := range machines {
		if _, err = machineZone(machine, cfg); err != nil {
			return err
		}
	}

	if err = resolveNetwork(s.Context, svc, cfg, s.Cluster.ControlPlane.NodeSets); err != nil {
		return err
	}

	if !slices.Contains(cfg.AllowedSourceRanges, "0.0.0.0/0") {
		s.Logger.Warnf("allowedSourceRanges %v must include the public IPs (or Cloud NAT IPs) of all control plane and worker nodes, "+
			"they reach kube-apiserver through the load balancer from them; otherwise control plane nodes can't join and workers can't register",
			cfg.AllowedSourceRanges)
	}

	// The control plane VMs need the firewall rules even when the user brings
	// their own kube-apiserver endpoint.
	if err = ensureFirewalls(s.Context, svc, cfg); err != nil {
		return err
	}

	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	ip, err := ensureLoadBalancer(s.Context, svc, cfg)
	if err != nil {
		return err
	}

	s.Cluster.APIEndpoint.Host = ip
	s.Cluster.APIEndpoint.Port = apiServerPort

	return nil
}

func (p *Provider) LookupLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	cfg, err := resolveLB(s)
	if err != nil {
		return err
	}

	svc, err := newComputeService(s)
	if err != nil {
		return err
	}

	fr, err := svc.ForwardingRules.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(s.Context).Do()
	switch {
	case isNotFound(err) && s.DestroyControlPlaneMachines:
		// Let reset clean up a partially created or deleted load balancer
		s.Logger.Warnf("forwarding rule %q not found, the kube-apiserver load balancer is incomplete or already deleted", cfg.Name)

		return nil
	case err != nil:
		return fail.Cloud(err, "gce", "looking up forwarding rule %q", cfg.Name)
	}

	s.Logger.Debugf("found forwarding rule %q with address: %s", cfg.Name, fr.IPAddress)
	s.Cluster.APIEndpoint.Host = fr.IPAddress
	s.Cluster.APIEndpoint.Port = apiServerPort

	return nil
}

// CleanupLoadBalancer deletes the load balancer resources and the firewall
// rules created by EnsureLoadBalancer and EnsureVM. Only resources owned by
// the cluster are deleted, see lbConfig.owner.
func (p *Provider) CleanupLoadBalancer(s *state.State) error {
	cfg, err := resolveLB(s)
	if err != nil {
		return err
	}

	svc, err := newComputeService(s)
	if err != nil {
		return err
	}

	// The firewall rules of a Shared VPC network are in its host project.
	// A misconfiguration or a deleted subnetwork must not block the cleanup.
	if err = resolveNetwork(s.Context, svc, cfg, s.Cluster.ControlPlane.NodeSets); err != nil {
		s.Logger.Warnf("resolving the network of the control plane failed, looking for the firewall rules in project %q: %v", cfg.firewallProject(), err)
	}

	return cleanupLoadBalancer(s.Context, svc, cfg, s.Cluster.APIEndpoint.Host, s.Logger)
}

// EnsureLoadBalancerMembers adds all control plane VMs to the load balancer,
// see cloudprovider.LoadBalancerMemberProvider. It runs after the followers
// joined the cluster, EnsureVM adds only the leader.
func (p *Provider) EnsureLoadBalancerMembers(s *state.State) error {
	machines, err := p.GenerateMachines(s.Cluster.Name, s.Cluster.ControlPlane.NodeSets, s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	return ensureLoadBalancerEndpoints(s, machines)
}

// ensureLoadBalancerEndpoints adds the VMs of the machines as endpoints to the
// network endpoint groups of their zones, creating the groups and adding them
// to the backend service as needed. Nothing is done if KubeOne doesn't manage
// the load balancer, i.e. the user provides the kube-apiserver endpoint.
//
// A VM must only be added once it serves kube-apiserver: the guest
// environment of every VM behind a passthrough load balancer routes the
// load balancer address to the VM itself, so a VM added before it joined the
// cluster would reach itself instead of the control plane and never join.
func ensureLoadBalancerEndpoints(s *state.State, machines []clusterv1alpha1.Machine) error {
	cfg, svc, bs, err := managedBackendService(s)
	if err != nil || bs == nil {
		return err
	}

	for _, machine := range machines {
		zone, err := machineZone(machine, cfg)
		if err != nil {
			return err
		}

		if err = ensureEndpoint(s.Context, svc, cfg, bs, zone, machine.Name); err != nil {
			return err
		}
	}

	return nil
}

// ensureLeaderEndpoint adds the VM of the leader machine to the load balancer
// while the cluster is created, so that it can initialize the cluster and the
// followers can join through the load balancer.
func ensureLeaderEndpoint(s *state.State, machine clusterv1alpha1.Machine) error {
	cfg, svc, bs, err := managedBackendService(s)
	if err != nil || bs == nil {
		return err
	}

	return ensureLeaderEndpointWith(s.Context, svc, cfg, bs, machine, s.Logger)
}

// ensureLeaderEndpointWith adds the leader VM unless another member of the
// load balancer is healthy. In that case a cluster is serving through the load
// balancer and the VM joins it, e.g. the replacement of a lost VM, so it's
// added by EnsureLoadBalancerMembers after it joined. Health is the deciding
// signal: endpoints of VMs that were reset but kept stay in the load balancer,
// but are unhealthy.
func ensureLeaderEndpointWith(ctx context.Context, svc *compute.Service, cfg *lbConfig, bs *compute.BackendService, machine clusterv1alpha1.Machine, logger logrus.FieldLogger) error {
	healthy, err := healthyMembers(ctx, svc, cfg, bs)
	if err != nil {
		return err
	}
	delete(healthy, machine.Name)

	if len(healthy) > 0 {
		logger.Debugf("load balancer %q has healthy members %v, %q is added after it joined the cluster", cfg.Name, slices.Sorted(maps.Keys(healthy)), machine.Name)

		return nil
	}

	zone, err := machineZone(machine, cfg)
	if err != nil {
		return err
	}

	return ensureEndpoint(ctx, svc, cfg, bs, zone, machine.Name)
}

// RemoveJoiningLoadBalancerMembers ensures the leader is a member of the load
// balancer and removes the joining hosts from it, see
// cloudprovider.LoadBalancerMemberProvider. The hostnames are the instance
// names, see hostConfigsFromMachines.
func (p *Provider) RemoveJoiningLoadBalancerMembers(s *state.State, joining []kubeoneapi.HostConfig) error {
	cfg, svc, bs, err := managedBackendService(s)
	if err != nil || bs == nil {
		return err
	}

	machines, err := p.GenerateMachines(s.Cluster.Name, s.Cluster.ControlPlane.NodeSets, s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	// The elected leader serves kube-apiserver at this point: on a new cluster
	// it's the first nodeSet machine EnsureVM added, on a repaired cluster a
	// healthy node, while the first nodeSet machine might be a new VM joining
	leader, err := s.Cluster.Leader()
	if err != nil {
		return fail.Config(err, "finding the control plane leader")
	}

	var joiningNames []string
	for _, host := range joining {
		joiningNames = append(joiningNames, host.Hostname)
	}

	return removeJoiningEndpoints(s.Context, svc, cfg, bs, machines, leader.Hostname, joiningNames, s.Logger)
}

// removeJoiningEndpoints adds the leader VM, which serves kube-apiserver at
// this point, and removes the endpoints of the joining VMs.
func removeJoiningEndpoints(ctx context.Context, svc *compute.Service, cfg *lbConfig, bs *compute.BackendService, machines []clusterv1alpha1.Machine, leaderName string, joining []string, logger logrus.FieldLogger) error {
	for _, machine := range machines {
		// A joining host is never added, it's added after it joined
		if machine.Name != leaderName || slices.Contains(joining, leaderName) {
			continue
		}

		zone, err := machineZone(machine, cfg)
		if err != nil {
			return err
		}

		if err = ensureEndpoint(ctx, svc, cfg, bs, zone, machine.Name); err != nil {
			return err
		}
	}

	if len(joining) == 0 {
		return nil
	}

	// Adding the leader might have added its network endpoint group
	bs, err := svc.RegionBackendServices.Get(cfg.ProjectID, cfg.Region, bs.Name).Context(ctx).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "getting backend service %q", cfg.Name)
	}

	for _, neg := range backendNEGs(bs) {
		var detach []*compute.NetworkEndpoint

		err := svc.NetworkEndpointGroups.ListNetworkEndpoints(neg.project, neg.zone, neg.name, &compute.NetworkEndpointGroupsListEndpointsRequest{}).Pages(ctx,
			func(page *compute.NetworkEndpointGroupsListNetworkEndpoints) error {
				for _, item := range page.Items {
					if item.NetworkEndpoint != nil && slices.Contains(joining, path.Base(item.NetworkEndpoint.Instance)) {
						detach = append(detach, item.NetworkEndpoint)
					}
				}

				return nil
			})
		switch {
		case isNotFound(err):
			continue
		case err != nil:
			return fail.Cloud(err, "gce", "listing endpoints of network endpoint group %q", neg.name)
		case len(detach) == 0:
			continue
		}

		for _, ep := range detach {
			logger.Infof("removing %q from the load balancer until it joined the cluster", path.Base(ep.Instance))
		}

		op, err := svc.NetworkEndpointGroups.DetachNetworkEndpoints(neg.project, neg.zone, neg.name, &compute.NetworkEndpointGroupsDetachEndpointsRequest{
			NetworkEndpoints: detach,
		}).Context(ctx).Do()
		if err != nil {
			return fail.Cloud(err, "gce", "removing endpoints from network endpoint group %q", neg.name)
		}

		if err = waitForOperation(ctx, svc, neg.project, op); err != nil {
			return err
		}
	}

	return nil
}

// managedBackendService returns the backend service of the load balancer, or
// nil if KubeOne doesn't manage the load balancer.
func managedBackendService(s *state.State) (*lbConfig, *compute.Service, *compute.BackendService, error) {
	cfg, err := resolveLB(s)
	if err != nil {
		return nil, nil, nil, err
	}

	svc, err := newComputeService(s)
	if err != nil {
		return nil, nil, nil, err
	}

	bs, err := svc.RegionBackendServices.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(s.Context).Do()
	switch {
	case isNotFound(err):
		s.Logger.Debugf("no backend service %q, the kube-apiserver load balancer is not managed by KubeOne", cfg.Name)

		return cfg, svc, nil, nil
	case err != nil:
		return nil, nil, nil, fail.Cloud(err, "gce", "getting backend service %q", cfg.Name)
	case bs.Description != cfg.owner():
		s.Logger.Debugf("backend service %q is not owned by KubeOne, not adding control plane VMs", cfg.Name)

		return cfg, svc, nil, nil
	}

	return cfg, svc, bs, nil
}

type negRef struct {
	project, zone, name string
}

// backendNEGs returns the network endpoint groups of the backend service.
func backendNEGs(bs *compute.BackendService) []negRef {
	var negs []negRef

	for _, backend := range bs.Backends {
		// projects/<project>/zones/<zone>/networkEndpointGroups/<name>
		parts := strings.Split(resourcePath(backend.Group), "/")
		if len(parts) == 6 && parts[0] == "projects" && parts[2] == "zones" && parts[4] == "networkEndpointGroups" {
			negs = append(negs, negRef{project: parts[1], zone: parts[3], name: parts[5]})
		}
	}

	return negs
}

// healthyMembers returns the instance names of the healthy members of the
// load balancer, as seen by its health check.
func healthyMembers(ctx context.Context, svc *compute.Service, cfg *lbConfig, bs *compute.BackendService) (map[string]bool, error) {
	healthy := map[string]bool{}

	for _, backend := range bs.Backends {
		health, err := svc.RegionBackendServices.GetHealth(cfg.ProjectID, cfg.Region, bs.Name, &compute.ResourceGroupReference{
			Group: backend.Group,
		}).Context(ctx).Do()
		if err != nil {
			return nil, fail.Cloud(err, "gce", "getting health of backend service %q", bs.Name)
		}

		for _, status := range health.HealthStatus {
			if status.HealthState == "HEALTHY" {
				healthy[path.Base(status.Instance)] = true
			}
		}
	}

	return healthy, nil
}

// resolveNetwork sets the network of the control plane VMs, where the
// firewall rules are created, from the nodeSets: their network, or the
// network of their subnetwork. All nodeSets must be in the same network, and
// a configured loadBalancer.network must be that network. Networks are
// compared as projects/<project>/global/networks/<name>, the network of a
// Shared VPC is in its host project.
func resolveNetwork(ctx context.Context, svc *compute.Service, cfg *lbConfig, nodeSets []kubeoneapi.NodeSet) error {
	networks := map[string][]string{}
	for _, ns := range nodeSets {
		network, err := nodeSetNetwork(ctx, svc, cfg, ns)
		if err != nil {
			return err
		}
		networks[network] = append(networks[network], ns.Name)
	}

	if len(networks) > 1 {
		return fail.ConfigValidation(fmt.Errorf("all control plane nodeSets must be in the same network, got %v", networks))
	}

	resolved := networkFullPath(cfg.ProjectID, "default")
	for network := range networks {
		resolved = network
	}

	if cfg.Network != "" {
		if configured := networkFullPath(cfg.ProjectID, cfg.Network); configured != resolved {
			return fail.ConfigValidation(fmt.Errorf(".cloudProvider.gce.controlPlane.loadBalancer.network %q must be the network %q of the control plane nodeSets", configured, resolved))
		}
	}

	cfg.Network = resolved

	return nil
}

// networkFullPath returns the network name or (partial) URL as
// projects/<project>/global/networks/<name>, names are in the given project.
func networkFullPath(project, network string) string {
	network = resourcePath(network)

	switch {
	case strings.HasPrefix(network, "projects/"):
		return network
	case strings.Contains(network, "/"):
		return fmt.Sprintf("projects/%s/%s", project, network)
	default:
		return fmt.Sprintf("projects/%s/global/networks/%s", project, network)
	}
}

// nodeSetNetwork returns the network of the nodeSet VMs as projects/<p>/global/networks/<n>.
// Without network and subnetwork the VMs are in the default network.
func nodeSetNetwork(ctx context.Context, svc *compute.Service, cfg *lbConfig, ns kubeoneapi.NodeSet) (string, error) {
	var spec gcetypes.RawConfig
	if err := jsonutil.StrictUnmarshal(ns.CloudProviderSpec, &spec); err != nil {
		return "", fail.Config(err, "decode gce config")
	}

	switch {
	case spec.Network.Value != "":
		return networkFullPath(cfg.ProjectID, spec.Network.Value), nil
	case spec.Subnetwork.Value == "":
		return networkFullPath(cfg.ProjectID, "default"), nil
	}

	// The subnetwork is a name or a (partial) URL, possibly of a Shared VPC host project
	subnetwork := spec.Subnetwork.Value
	project, region := cfg.ProjectID, cfg.Region
	parts := strings.Split(resourcePath(subnetwork), "/")
	for i := 0; i+1 < len(parts); i++ {
		switch parts[i] {
		case "projects":
			project = parts[i+1]
		case "regions":
			region = parts[i+1]
		}
	}

	sn, err := svc.Subnetworks.Get(project, region, path.Base(subnetwork)).Context(ctx).Do()
	if err != nil {
		return "", fail.Cloud(err, "gce", "getting subnetwork %q", subnetwork)
	}

	return networkFullPath(cfg.ProjectID, sn.Network), nil
}

// leaderMachineName returns the name of the machine of the leader, the first
// control plane host, which initializes the cluster.
func leaderMachineName(clusterName string, nodeSets []kubeoneapi.NodeSet) string {
	for _, ns := range nodeSets {
		if ns.Replicas > 0 {
			return ns.MachineName(clusterName, 0)
		}
	}

	return ""
}

// machineLocation returns the zone and project of the machine's VM. The
// machine-controller driver creates the VM in the project of the nodeSet, or
// of the service account if the nodeSet sets none.
func machineLocation(machine clusterv1alpha1.Machine, cfg *lbConfig) (zone, project string, err error) {
	pconfig, err := providerconfig.GetConfig(machine.Spec.ProviderSpec)
	if err != nil {
		return "", "", fail.Config(err, "reading provider config")
	}

	gceConfig, err := gcetypes.GetConfig(*pconfig)
	if err != nil {
		return "", "", fail.Config(err, "decode gce config")
	}

	if gceConfig.Zone.Value == "" {
		return "", "", fail.ConfigValidation(fmt.Errorf("control plane machine %q has no zone", machine.Name))
	}

	project = gceConfig.ProjectID.Value
	if project == "" {
		project = cfg.ServiceAccountProjectID
	}

	return gceConfig.Zone.Value, project, nil
}

// machineZone returns the zone of the machine and checks that its VM is in
// the project of the load balancer, network endpoint groups can't reference
// VMs of other projects. It's only used to add VMs to the load balancer, so a
// misconfigured project doesn't block reset.
func machineZone(machine clusterv1alpha1.Machine, cfg *lbConfig) (string, error) {
	zone, project, err := machineLocation(machine, cfg)
	if err != nil {
		return "", err
	}

	if project != cfg.ProjectID {
		const msg = "control plane machine %q is created in project %q (cloudProviderSpec.projectID or the project of the service account), " +
			"but the load balancer is in project %q; set cloudProviderSpec.projectID of the control plane nodeSets to %q"

		return "", fail.ConfigValidation(fmt.Errorf(msg, machine.Name, project, cfg.ProjectID, cfg.ProjectID))
	}

	return zone, nil
}

func ensureFirewalls(ctx context.Context, svc *compute.Service, cfg *lbConfig) error {
	targetTags := []string{controlPlaneTag(cfg.ClusterName)}
	allowed := []*compute.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{strconv.Itoa(apiServerPort)}}}

	sourceRanges := cfg.AllowedSourceRanges
	if len(sourceRanges) == 0 {
		sourceRanges = []string{"0.0.0.0/0"}
	}

	for _, fw := range []*compute.Firewall{
		{
			Name:         cfg.Name + apiFirewallSuffix,
			Description:  cfg.owner(),
			Network:      cfg.Network,
			Direction:    "INGRESS",
			TargetTags:   targetTags,
			SourceRanges: sourceRanges,
			Allowed:      allowed,
		},
		{
			Name:         cfg.Name + healthCheckFirewallSuffix,
			Description:  cfg.owner(),
			Network:      cfg.Network,
			Direction:    "INGRESS",
			TargetTags:   targetTags,
			SourceRanges: healthCheckSourceRanges,
			Allowed:      allowed,
		},
	} {
		if err := ensureFirewall(ctx, svc, cfg, fw); err != nil {
			return err
		}
	}

	return nil
}

// ensureFirewall creates the firewall rule, or updates the source ranges of
// an existing one, e.g. after allowedSourceRanges changed.
func ensureFirewall(ctx context.Context, svc *compute.Service, cfg *lbConfig, want *compute.Firewall) error {
	existing, err := svc.Firewalls.Get(cfg.firewallProject(), want.Name).Context(ctx).Do()
	switch {
	case isNotFound(err):
		op, insertErr := svc.Firewalls.Insert(cfg.firewallProject(), want).Context(ctx).Do()
		if insertErr != nil {
			return fail.Cloud(insertErr, "gce", "creating firewall rule %q", want.Name)
		}

		return waitForOperation(ctx, svc, cfg.firewallProject(), op)
	case err != nil:
		return fail.Cloud(err, "gce", "getting firewall rule %q", want.Name)
	}

	if err = checkOwner(cfg, "firewall rule", want.Name, existing.Description); err != nil {
		return err
	}

	// The network of a firewall rule can't be changed
	if have, wantNetwork := networkFullPath(cfg.ProjectID, existing.Network), networkFullPath(cfg.ProjectID, want.Network); have != wantNetwork {
		return fail.Cloud(
			fmt.Errorf("firewall rule %q is in network %q, but the control plane is in network %q; delete the firewall rule to have it recreated", want.Name, have, wantNetwork),
			"gce", "ensuring firewall rule %q", want.Name,
		)
	}

	if sets.New(existing.SourceRanges...).Equal(sets.New(want.SourceRanges...)) &&
		sets.New(existing.TargetTags...).Equal(sets.New(want.TargetTags...)) {
		return nil
	}

	op, err := svc.Firewalls.Patch(cfg.firewallProject(), want.Name, &compute.Firewall{
		SourceRanges: want.SourceRanges,
		TargetTags:   want.TargetTags,
	}).Context(ctx).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "updating firewall rule %q", want.Name)
	}

	return waitForOperation(ctx, svc, cfg.firewallProject(), op)
}

// ensureLoadBalancer ensures the address, health check, backend service and
// forwarding rule, and returns the IP address of the load balancer.
func ensureLoadBalancer(ctx context.Context, svc *compute.Service, cfg *lbConfig) (string, error) {
	err := ensureResource(ctx, svc, cfg, "address", cfg.Name,
		func() (string, error) {
			addr, getErr := svc.Addresses.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
			if getErr != nil {
				return "", getErr
			}

			return addr.Description, nil
		},
		func() (*compute.Operation, error) {
			return svc.Addresses.Insert(cfg.ProjectID, cfg.Region, &compute.Address{
				Name:        cfg.Name,
				Description: cfg.owner(),
				AddressType: schemeExternal,
			}).Context(ctx).Do()
		},
	)
	if err != nil {
		return "", err
	}

	err = ensureResource(ctx, svc, cfg, "health check", cfg.Name,
		func() (string, error) {
			hc, getErr := svc.RegionHealthChecks.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
			if getErr != nil {
				return "", getErr
			}

			return hc.Description, nil
		},
		func() (*compute.Operation, error) {
			// Interval, timeout and thresholds are left to the GCE defaults
			return svc.RegionHealthChecks.Insert(cfg.ProjectID, cfg.Region, &compute.HealthCheck{
				Name:        cfg.Name,
				Description: cfg.owner(),
				Type:        protocolTCP,
				TcpHealthCheck: &compute.TCPHealthCheck{
					Port: apiServerPort,
				},
			}).Context(ctx).Do()
		},
	)
	if err != nil {
		return "", err
	}

	err = ensureResource(ctx, svc, cfg, "backend service", cfg.Name,
		func() (string, error) {
			bs, getErr := svc.RegionBackendServices.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
			if getErr != nil {
				return "", getErr
			}

			return bs.Description, nil
		},
		func() (*compute.Operation, error) {
			// Backends are added by ensureEndpoint as the VMs are created
			return svc.RegionBackendServices.Insert(cfg.ProjectID, cfg.Region, &compute.BackendService{
				Name:                cfg.Name,
				Description:         cfg.owner(),
				LoadBalancingScheme: schemeExternal,
				Protocol:            protocolTCP,
				HealthChecks:        []string{cfg.regionalPath("healthChecks", cfg.Name)},
			}).Context(ctx).Do()
		},
	)
	if err != nil {
		return "", err
	}

	addr, err := svc.Addresses.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
	if err != nil {
		return "", fail.Cloud(err, "gce", "getting address %q", cfg.Name)
	}

	err = ensureResource(ctx, svc, cfg, "forwarding rule", cfg.Name,
		func() (string, error) {
			fr, getErr := svc.ForwardingRules.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
			if getErr != nil {
				return "", getErr
			}

			return fr.Description, nil
		},
		func() (*compute.Operation, error) {
			return svc.ForwardingRules.Insert(cfg.ProjectID, cfg.Region, &compute.ForwardingRule{
				Name:                cfg.Name,
				Description:         cfg.owner(),
				LoadBalancingScheme: schemeExternal,
				IPProtocol:          protocolTCP,
				Ports:               []string{strconv.Itoa(apiServerPort)},
				IPAddress:           addr.Address,
				BackendService:      cfg.regionalPath("backendServices", cfg.Name),
			}).Context(ctx).Do()
		},
	)
	if err != nil {
		return "", err
	}

	return addr.Address, nil
}

// ensureEndpoint ensures the network endpoint group of the zone, that it's a
// backend of the backend service, and that the VM is one of its endpoints.
func ensureEndpoint(ctx context.Context, svc *compute.Service, cfg *lbConfig, bs *compute.BackendService, zone, instanceName string) error {
	instance, err := svc.Instances.Get(cfg.ProjectID, zone, instanceName).Context(ctx).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "getting instance %q", instanceName)
	}
	if len(instance.NetworkInterfaces) == 0 {
		return fail.Cloud(errors.New("instance has no network interface"), "gce", "getting network of instance %q", instanceName)
	}
	nic := instance.NetworkInterfaces[0]

	negName := cfg.negName(zone)
	err = ensureResource(ctx, svc, cfg, "network endpoint group", negName,
		func() (string, error) {
			neg, getErr := svc.NetworkEndpointGroups.Get(cfg.ProjectID, zone, negName).Context(ctx).Do()
			if getErr != nil {
				return "", getErr
			}

			if resourcePath(neg.Subnetwork) != resourcePath(nic.Subnetwork) {
				return neg.Description, fmt.Errorf("network endpoint group %q is in subnetwork %q, instance %q in %q; all control plane VMs of a zone must be in the same subnetwork",
					negName, resourcePath(neg.Subnetwork), instanceName, resourcePath(nic.Subnetwork))
			}

			return neg.Description, nil
		},
		func() (*compute.Operation, error) {
			return svc.NetworkEndpointGroups.Insert(cfg.ProjectID, zone, &compute.NetworkEndpointGroup{
				Name:                negName,
				Description:         cfg.owner(),
				NetworkEndpointType: negTypeGCEVMIP,
				Network:             nic.Network,
				Subnetwork:          nic.Subnetwork,
			}).Context(ctx).Do()
		},
	)
	if err != nil {
		return err
	}

	if err = ensureBackend(ctx, svc, cfg, bs, zone); err != nil {
		return err
	}

	// An endpoint of a former VM with the same name, e.g. a recreated VM, has
	// another IP address and is replaced
	var stale []*compute.NetworkEndpoint
	attached := false
	err = svc.NetworkEndpointGroups.ListNetworkEndpoints(cfg.ProjectID, zone, negName, &compute.NetworkEndpointGroupsListEndpointsRequest{}).Pages(ctx,
		func(page *compute.NetworkEndpointGroupsListNetworkEndpoints) error {
			for _, item := range page.Items {
				ep := item.NetworkEndpoint
				switch {
				case ep == nil || path.Base(ep.Instance) != instanceName:
				case ep.IpAddress == "" || ep.IpAddress == nic.NetworkIP:
					attached = true
				default:
					stale = append(stale, ep)
				}
			}

			return nil
		})
	if err != nil {
		return fail.Cloud(err, "gce", "listing endpoints of network endpoint group %q", negName)
	}
	if attached {
		return nil
	}

	if len(stale) > 0 {
		op, detachErr := svc.NetworkEndpointGroups.DetachNetworkEndpoints(cfg.ProjectID, zone, negName, &compute.NetworkEndpointGroupsDetachEndpointsRequest{
			NetworkEndpoints: stale,
		}).Context(ctx).Do()
		if detachErr != nil {
			return fail.Cloud(detachErr, "gce", "removing stale endpoints of %q from network endpoint group %q", instanceName, negName)
		}
		if err = waitForOperation(ctx, svc, cfg.ProjectID, op); err != nil {
			return err
		}
	}

	op, err := svc.NetworkEndpointGroups.AttachNetworkEndpoints(cfg.ProjectID, zone, negName, &compute.NetworkEndpointGroupsAttachEndpointsRequest{
		NetworkEndpoints: []*compute.NetworkEndpoint{{
			Instance:  instanceName,
			IpAddress: nic.NetworkIP,
		}},
	}).Context(ctx).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "attaching instance %q to network endpoint group %q", instanceName, negName)
	}

	return waitForOperation(ctx, svc, cfg.ProjectID, op)
}

// ensureBackend adds the network endpoint group of the zone to the backend
// service, unless it's already a backend.
func ensureBackend(ctx context.Context, svc *compute.Service, cfg *lbConfig, bs *compute.BackendService, zone string) error {
	negPath := fmt.Sprintf("projects/%s/zones/%s/networkEndpointGroups/%s", cfg.ProjectID, zone, cfg.negName(zone))

	// The backend service passed in might be stale if another zone was added
	// in the meantime, the fingerprint must match for the patch.
	current, err := svc.RegionBackendServices.Get(cfg.ProjectID, cfg.Region, bs.Name).Context(ctx).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "getting backend service %q", bs.Name)
	}

	for _, b := range current.Backends {
		if resourcePath(b.Group) == negPath {
			return nil
		}
	}

	backends := append(slices.Clone(current.Backends), &compute.Backend{
		Group:         negPath,
		BalancingMode: balancingModeConnection,
	})

	op, err := svc.RegionBackendServices.Patch(cfg.ProjectID, cfg.Region, bs.Name, &compute.BackendService{
		Backends:    backends,
		Fingerprint: current.Fingerprint,
	}).Context(ctx).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "adding network endpoint group %q to backend service %q", cfg.negName(zone), bs.Name)
	}

	return waitForOperation(ctx, svc, cfg.ProjectID, op)
}

// ensureResource creates a resource if get reports it as not found. An
// existing resource must be owned by the cluster, see lbConfig.owner.
func ensureResource(ctx context.Context, svc *compute.Service, cfg *lbConfig, kind, name string, get func() (string, error), create func() (*compute.Operation, error)) error {
	description, err := get()
	switch {
	case isNotFound(err):
		op, createErr := create()
		if createErr != nil {
			return fail.Cloud(createErr, "gce", "creating %s %q", kind, name)
		}

		return waitForOperation(ctx, svc, cfg.ProjectID, op)
	case err != nil:
		return fail.Cloud(err, "gce", "getting %s %q", kind, name)
	}

	return checkOwner(cfg, kind, name, description)
}

func checkOwner(cfg *lbConfig, kind, name, description string) error {
	if description == cfg.owner() {
		return nil
	}

	return fail.Cloud(
		fmt.Errorf("%s %q already exists and was not created by KubeOne for this cluster (description %q, expected %q), choose another .cloudProvider.gce.controlPlane.loadBalancer.name", kind, name, description, cfg.owner()),
		"gce", "ensuring %s %q", kind, name,
	)
}

// cleanupLoadBalancer deletes the owned load balancer resources in reverse
// order of their dependencies, GCE refuses to delete a resource that is still
// used by another one. Missing resources are skipped.
//
// apiHost is the kube-apiserver endpoint. LookupLoadBalancer sets it to the
// load balancer address, so a different address means the user provides the
// endpoint and the load balancer is kept. The firewall rules are created in
// either case and are always deleted.
func cleanupLoadBalancer(ctx context.Context, svc *compute.Service, cfg *lbConfig, apiHost string, logger logrus.FieldLogger) error {
	keepLB := false
	if apiHost != "" {
		addr, err := svc.Addresses.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
		switch {
		case isNotFound(err):
		case err != nil:
			return fail.Cloud(err, "gce", "getting address %q", cfg.Name)
		case addr.Address != apiHost:
			logger.Infof("kube-apiserver endpoint %q is not the load balancer address %q, keeping the load balancer", apiHost, addr.Address)
			keepLB = true
		}
	}

	if !keepLB {
		if err := deleteLoadBalancer(ctx, svc, cfg, logger); err != nil {
			return err
		}
	}

	for _, name := range []string{cfg.Name + apiFirewallSuffix, cfg.Name + healthCheckFirewallSuffix} {
		err := deleteResource(ctx, svc, cfg, logger, "firewall rule", name,
			func() (string, error) {
				fw, getErr := svc.Firewalls.Get(cfg.firewallProject(), name).Context(ctx).Do()
				if getErr != nil {
					return "", getErr
				}

				return fw.Description, nil
			},
			func() (*compute.Operation, error) {
				return svc.Firewalls.Delete(cfg.firewallProject(), name).Context(ctx).Do()
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteLoadBalancer(ctx context.Context, svc *compute.Service, cfg *lbConfig, logger logrus.FieldLogger) error {
	err := deleteResource(ctx, svc, cfg, logger, "forwarding rule", cfg.Name,
		func() (string, error) {
			fr, getErr := svc.ForwardingRules.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
			if getErr != nil {
				return "", getErr
			}

			return fr.Description, nil
		},
		func() (*compute.Operation, error) {
			return svc.ForwardingRules.Delete(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
		},
	)
	if err != nil {
		return err
	}

	err = deleteResource(ctx, svc, cfg, logger, "backend service", cfg.Name,
		func() (string, error) {
			bs, getErr := svc.RegionBackendServices.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
			if getErr != nil {
				return "", getErr
			}

			return bs.Description, nil
		},
		func() (*compute.Operation, error) {
			return svc.RegionBackendServices.Delete(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
		},
	)
	if err != nil {
		return err
	}

	// The zones of the network endpoint groups aren't known without the
	// nodeSets of the time they were created, so all zones are searched.
	var negs []*compute.NetworkEndpointGroup
	err = svc.NetworkEndpointGroups.AggregatedList(cfg.ProjectID).Pages(ctx, func(page *compute.NetworkEndpointGroupAggregatedList) error {
		for _, scoped := range page.Items {
			for _, neg := range scoped.NetworkEndpointGroups {
				if neg.Zone != "" && neg.Name == cfg.negName(path.Base(neg.Zone)) {
					negs = append(negs, neg)
				}
			}
		}

		return nil
	})
	if err != nil {
		return fail.Cloud(err, "gce", "listing network endpoint groups")
	}

	for _, neg := range negs {
		zone := path.Base(neg.Zone)
		err = deleteResource(ctx, svc, cfg, logger, "network endpoint group", neg.Name,
			func() (string, error) {
				return neg.Description, nil
			},
			func() (*compute.Operation, error) {
				return svc.NetworkEndpointGroups.Delete(cfg.ProjectID, zone, neg.Name).Context(ctx).Do()
			},
		)
		if err != nil {
			return err
		}
	}

	err = deleteResource(ctx, svc, cfg, logger, "health check", cfg.Name,
		func() (string, error) {
			hc, getErr := svc.RegionHealthChecks.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
			if getErr != nil {
				return "", getErr
			}

			return hc.Description, nil
		},
		func() (*compute.Operation, error) {
			return svc.RegionHealthChecks.Delete(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
		},
	)
	if err != nil {
		return err
	}

	return deleteResource(ctx, svc, cfg, logger, "address", cfg.Name,
		func() (string, error) {
			addr, getErr := svc.Addresses.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
			if getErr != nil {
				return "", getErr
			}

			return addr.Description, nil
		},
		func() (*compute.Operation, error) {
			return svc.Addresses.Delete(cfg.ProjectID, cfg.Region, cfg.Name).Context(ctx).Do()
		},
	)
}

// deleteResource deletes a resource if it exists and is owned by the
// cluster. Resources of others are logged and kept.
func deleteResource(ctx context.Context, svc *compute.Service, cfg *lbConfig, logger logrus.FieldLogger, kind, name string, get func() (string, error), del func() (*compute.Operation, error)) error {
	description, err := get()
	switch {
	case isNotFound(err):
		logger.Debugf("%s %q not found, nothing to delete", kind, name)

		return nil
	case err != nil:
		return fail.Cloud(err, "gce", "getting %s %q", kind, name)
	case description != cfg.owner():
		logger.Warnf("%s %q was not created by KubeOne for cluster %q (description %q), not deleting it", kind, name, cfg.ClusterName, description)

		return nil
	}

	logger.Infof("deleting %s %q", kind, name)

	op, err := del()
	switch {
	case isNotFound(err):
		return nil
	case err != nil:
		return fail.Cloud(err, "gce", "deleting %s %q", kind, name)
	}

	return waitForOperation(ctx, svc, cfg.ProjectID, op)
}

// resourcePath strips the API prefix from a resource URL, so that full URLs
// ("https://www.googleapis.com/compute/v1/projects/...") and partial URLs
// ("projects/...") can be compared.
func resourcePath(url string) string {
	if idx := strings.Index(url, "projects/"); idx >= 0 {
		return url[idx:]
	}

	return url
}
