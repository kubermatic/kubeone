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
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/compute/v1"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	gcetypes "k8c.io/machine-controller/sdk/cloudprovider/gce"
	"k8c.io/machine-controller/sdk/providerconfig"
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
// backend service-based external passthrough Network Load Balancers.
var healthCheckSourceRanges = []string{"35.191.0.0/16", "209.85.204.0/22"}

// lbConfig is the resolved load balancer configuration.
type lbConfig struct {
	ClusterName         string
	ProjectID           string
	Region              string
	Name                string
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

	if cfg.ProjectID == "" {
		serviceAccount, err := serviceAccountJSON(s)
		if err != nil {
			return nil, err
		}

		if cfg.ProjectID, err = serviceAccountProjectID(serviceAccount); err != nil {
			return nil, err
		}
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

	return cleanupLoadBalancer(s.Context, svc, cfg, s.Cluster.APIEndpoint.Host, s.Logger)
}

// ensureLoadBalancerEndpoint adds the VM of the machine as endpoint to the
// network endpoint group of its zone, creating the group and adding it to the
// backend service as needed. Nothing is done if KubeOne doesn't manage the
// load balancer, i.e. the user provides the kube-apiserver endpoint.
func ensureLoadBalancerEndpoint(s *state.State, machine clusterv1alpha1.Machine) error {
	cfg, err := resolveLB(s)
	if err != nil {
		return err
	}

	svc, err := newComputeService(s)
	if err != nil {
		return err
	}

	bs, err := svc.RegionBackendServices.Get(cfg.ProjectID, cfg.Region, cfg.Name).Context(s.Context).Do()
	switch {
	case isNotFound(err):
		s.Logger.Debugf("no backend service %q, the kube-apiserver load balancer is not managed by KubeOne", cfg.Name)

		return nil
	case err != nil:
		return fail.Cloud(err, "gce", "getting backend service %q", cfg.Name)
	case bs.Description != cfg.owner():
		s.Logger.Debugf("backend service %q is not owned by KubeOne, not adding %q", cfg.Name, machine.Name)

		return nil
	}

	zone, err := machineZone(machine, cfg.ProjectID)
	if err != nil {
		return err
	}

	return ensureEndpoint(s.Context, svc, cfg, bs, zone, machine.Name)
}

// machineZone returns the zone of the machine and checks that it's in the
// project of the load balancer, network endpoint groups can't reference VMs
// of other projects.
func machineZone(machine clusterv1alpha1.Machine, projectID string) (string, error) {
	pconfig, err := providerconfig.GetConfig(machine.Spec.ProviderSpec)
	if err != nil {
		return "", fail.Config(err, "reading provider config")
	}

	gceConfig, err := gcetypes.GetConfig(*pconfig)
	if err != nil {
		return "", fail.Config(err, "decode gce config")
	}

	if p := gceConfig.ProjectID.Value; p != "" && p != projectID {
		return "", fail.ConfigValidation(fmt.Errorf("control plane machine %q is in project %q, but the load balancer is in project %q", machine.Name, p, projectID))
	}

	if gceConfig.Zone.Value == "" {
		return "", fail.ConfigValidation(fmt.Errorf("control plane machine %q has no zone", machine.Name))
	}

	return gceConfig.Zone.Value, nil
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
	existing, err := svc.Firewalls.Get(cfg.ProjectID, want.Name).Context(ctx).Do()
	switch {
	case isNotFound(err):
		op, insertErr := svc.Firewalls.Insert(cfg.ProjectID, want).Context(ctx).Do()
		if insertErr != nil {
			return fail.Cloud(insertErr, "gce", "creating firewall rule %q", want.Name)
		}

		return waitForOperation(ctx, svc, cfg.ProjectID, op)
	case err != nil:
		return fail.Cloud(err, "gce", "getting firewall rule %q", want.Name)
	}

	if err = checkOwner(cfg, "firewall rule", want.Name, existing.Description); err != nil {
		return err
	}

	if sameSet(existing.SourceRanges, want.SourceRanges) {
		return nil
	}

	op, err := svc.Firewalls.Patch(cfg.ProjectID, want.Name, &compute.Firewall{SourceRanges: want.SourceRanges}).Context(ctx).Do()
	if err != nil {
		return fail.Cloud(err, "gce", "updating source ranges of firewall rule %q", want.Name)
	}

	return waitForOperation(ctx, svc, cfg.ProjectID, op)
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

	attached := false
	err = svc.NetworkEndpointGroups.ListNetworkEndpoints(cfg.ProjectID, zone, negName, &compute.NetworkEndpointGroupsListEndpointsRequest{}).Pages(ctx,
		func(page *compute.NetworkEndpointGroupsListNetworkEndpoints) error {
			for _, item := range page.Items {
				if item.NetworkEndpoint != nil && path.Base(item.NetworkEndpoint.Instance) == instanceName {
					attached = true
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
				fw, getErr := svc.Firewalls.Get(cfg.ProjectID, name).Context(ctx).Do()
				if getErr != nil {
					return "", getErr
				}

				return fw.Description, nil
			},
			func() (*compute.Operation, error) {
				return svc.Firewalls.Delete(cfg.ProjectID, name).Context(ctx).Do()
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

func sameSet(a, b []string) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)

	return slices.Equal(slices.Compact(a), slices.Compact(b))
}
