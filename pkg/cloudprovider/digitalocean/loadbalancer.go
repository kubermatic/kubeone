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

package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/digitalocean/godo"
	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	kubeAPIPort = 6443
	lbStatusUp  = "active"
	lbStatusErr = "errored"
	lbPageSize  = 200
)

var (
	// lbPollInterval and lbPollTimeout control how long to wait for a
	// load balancer to become active. Variables to allow overriding in tests.
	lbPollInterval = 10 * time.Second
	lbPollTimeout  = 10 * time.Minute

	// newClients builds the load balancers and VPCs clients from one godo
	// client. Variable to allow injecting fakes in tests.
	newClients = func(token string) (godo.LoadBalancersService, vpcGetter) {
		client := godo.NewFromToken(token)

		return client.LoadBalancers, client.VPCs
	}

	errLBNotFound = errors.New("load balancer not found")
)

// vpcGetter is the part of godo.VPCsService used to check load balancer VPCs.
type vpcGetter interface {
	Get(ctx context.Context, id string) (*godo.VPC, *godo.Response, error)
}

func lbSpec(cluster *kubeoneapi.KubeOneCluster) kubeoneapi.DigitalOceanLoadBalancer {
	if cluster.CloudProvider.DigitalOcean == nil || cluster.CloudProvider.DigitalOcean.ControlPlane == nil {
		return kubeoneapi.DigitalOceanLoadBalancer{}
	}

	return cluster.CloudProvider.DigitalOcean.ControlPlane.LoadBalancer
}

func (p *provider) HasLoadBalancer(s *state.State) bool {
	return p.Enabled(s)
}

func (p *provider) EnsureLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	client, vpcs, err := lbClient(s)
	if err != nil {
		return err
	}

	lb, err := ensureLoadBalancer(s.Context, client, vpcs, s.Cluster, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.APIEndpoint.Host = lb.IP
	s.Cluster.APIEndpoint.Port = kubeAPIPort

	return nil
}

func (p *provider) LookupLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	client, _, err := lbClient(s)
	if err != nil {
		return err
	}

	lb, err := lookupLoadBalancer(s.Context, client, s.Cluster)
	if err != nil {
		return err
	}

	if lb.IP == "" {
		return fail.Cloud(fmt.Errorf("load balancer %q has no IP address assigned yet", lb.Name), "digitalocean", "looking up load balancer")
	}

	s.Logger.Debugf("found load balancer %q with id: %s", lb.Name, lb.ID)
	if lbSpec(s.Cluster).ID == "" && !ownedByKubeOne(lb, s.Cluster.Name) {
		s.Logger.Warnf("load balancer %q targets %s instead of droplets tagged %q, kube-apiserver traffic might not reach the control plane", lb.Name, describeBackends(lb), ClusterTag(s.Cluster.Name))
	}
	s.Cluster.APIEndpoint.Host = lb.IP
	s.Cluster.APIEndpoint.Port = kubeAPIPort

	return nil
}

func (p *provider) CleanupLoadBalancer(s *state.State) error {
	spec := lbSpec(s.Cluster)
	if spec.ID != "" {
		s.Logger.Infof("load balancer %q is not managed by KubeOne, skipping deletion", spec.ID)

		return nil
	}

	client, _, err := lbClient(s)
	if err != nil {
		return err
	}

	return cleanupLoadBalancer(s.Context, client, s.Cluster, s.Logger)
}

func lbClient(s *state.State) (godo.LoadBalancersService, vpcGetter, error) {
	token, err := apiToken(s)
	if err != nil {
		return nil, nil, err
	}

	lbs, vpcs := newClients(token)

	return lbs, vpcs, nil
}

func ensureLoadBalancer(ctx context.Context, client godo.LoadBalancersService, vpcs vpcGetter, cluster *kubeoneapi.KubeOneCluster, logger logrus.FieldLogger) (*godo.LoadBalancer, error) {
	spec := lbSpec(cluster)

	// Load balancers created by KubeOne are always in the default VPC
	if spec.ID != "" {
		return ensureExistingLoadBalancer(ctx, client, vpcs, cluster, logger)
	}

	lb, err := findLoadBalancerByName(ctx, client, spec.Name, spec.Region, cluster.Name)
	switch {
	case err == nil:
		logger.Debugf("load balancer %q already exists with id: %s", lb.Name, lb.ID)
		if !ownedByKubeOne(lb, cluster.Name) {
			return nil, fail.Cloud(
				fmt.Errorf(
					"load balancer %q (id: %s) already exists but targets %s instead of droplets tagged %q, "+
						"it was not created by KubeOne for this cluster or was modified; "+
						"set .cloudProvider.digitalocean.controlPlane.loadBalancer.name to a different name, "+
						"or restore the load balancer to target droplets tagged %q",
					lb.Name, lb.ID, describeBackends(lb), ClusterTag(cluster.Name), ClusterTag(cluster.Name),
				),
				"digitalocean",
				"ensuring load balancer",
			)
		}

		// DigitalOcean rejects updates while the load balancer is provisioning
		if lb, err = waitForLoadBalancer(ctx, client, lb, false, logger); err != nil {
			return nil, err
		}

		// Restore a missing kube-apiserver rule and the health check
		return reconcileLoadBalancer(ctx, client, lb, "", logger)
	case errors.Is(err, errLBNotFound):
		logger.Infof("creating load balancer %q", spec.Name)
		lb, _, err = client.Create(ctx, createRequest(cluster))
		if err != nil {
			return nil, fail.Cloud(err, "digitalocean", "creating load balancer")
		}

		// A new load balancer might not be visible right away
		return waitForLoadBalancer(ctx, client, lb, true, logger)
	default:
		return nil, err
	}
}

func ensureExistingLoadBalancer(ctx context.Context, client godo.LoadBalancersService, vpcs vpcGetter, cluster *kubeoneapi.KubeOneCluster, logger logrus.FieldLogger) (*godo.LoadBalancer, error) {
	spec := lbSpec(cluster)
	backendTag := ClusterTag(cluster.Name)

	lb, err := getLoadBalancer(ctx, client, spec.ID)
	if err != nil {
		return nil, err
	}

	// DigitalOcean rejects updates while the load balancer is provisioning.
	// Waiting first also means the checks below see its final configuration.
	if lb, err = waitForLoadBalancer(ctx, client, lb, false, logger); err != nil {
		return nil, err
	}

	// Global load balancers only forward HTTP/HTTPS on port 80 or 443 with a
	// single rule, kube-apiserver needs TCP passthrough on 6443 so TLS and
	// client certificate authentication end at the apiserver
	if lb.Type == godo.LoadBalancerTypeGlobal {
		return nil, fail.Cloud(
			fmt.Errorf("load balancer %q is a global load balancer, which only forwards HTTP/HTTPS; kube-apiserver requires a regional load balancer forwarding TCP on port %d", lb.Name, kubeAPIPort),
			"digitalocean",
			"configuring load balancer %q", spec.ID,
		)
	}
	if err = checkRegion(lb, spec.Region); err != nil {
		return nil, fail.Cloud(err, "digitalocean", "configuring load balancer %q", spec.ID)
	}
	if err = checkDefaultVPC(ctx, vpcs, lb); err != nil {
		return nil, fail.Cloud(err, "digitalocean", "configuring load balancer %q", spec.ID)
	}

	setTag := ""
	switch {
	case lb.Tag == backendTag:
	case lb.Tag == "" && len(lb.DropletIDs) == 0:
		setTag = backendTag
	default:
		return nil, fail.Cloud(
			fmt.Errorf("load balancer %q already has backends (tag %q, %d droplets), it must target droplets tagged %q or have no backends", lb.Name, lb.Tag, len(lb.DropletIDs), backendTag),
			"digitalocean",
			"configuring load balancer %q", spec.ID,
		)
	}

	return reconcileLoadBalancer(ctx, client, lb, setTag, logger)
}

// reconcileLoadBalancer makes the active load balancer forward TCP 6443 to the
// control plane droplets and health check them on that port, setting setTag
// as backend selector if it's not empty. All changes are applied in
// a single Update, DigitalOcean rejects changes while a previous one is still
// being applied.
func reconcileLoadBalancer(ctx context.Context, client godo.LoadBalancersService, lb *godo.LoadBalancer, setTag string, logger logrus.FieldLogger) (*godo.LoadBalancer, error) {
	hasRule, err := hasKubeAPIForwardingRule(lb)
	if err != nil {
		return nil, fail.Cloud(err, "digitalocean", "configuring load balancer %q", lb.ID)
	}

	needsHealthCheck := !isKubeAPIHealthCheck(lb.HealthCheck)

	if setTag == "" && !needsHealthCheck && hasRule {
		return lb, nil
	}

	req := lb.AsRequest()
	// AsRequest copies both size fields, but they're mutually exclusive
	if req.SizeUnit > 0 {
		req.SizeSlug = ""
	}
	if setTag != "" {
		logger.Infof("configuring load balancer %q to target droplets tagged %q", lb.Name, setTag)
		req.Tag = setTag
	}
	// AsRequest also copies the droplets DigitalOcean resolved from the tag,
	// but tag and droplet_ids are mutually exclusive
	req.DropletIDs = nil
	if needsHealthCheck {
		logger.Infof("configuring load balancer %q to health check kube-apiserver on port %d", lb.Name, kubeAPIPort)
		req.HealthCheck = kubeAPIHealthCheck()
	}
	if !hasRule {
		logger.Infof("adding kube-apiserver forwarding rule to load balancer %q", lb.Name)
		req.ForwardingRules = append(req.ForwardingRules, kubeAPIForwardingRule())
	}

	// godo returns a nil load balancer on errors
	updated, _, err := client.Update(ctx, lb.ID, req)
	if err != nil {
		return nil, fail.Cloud(err, "digitalocean", "updating load balancer %q", lb.ID)
	}

	return waitForLoadBalancer(ctx, client, updated, false, logger)
}

// checkRegion fails if the load balancer isn't in the nodeSets region.
// Droplets are always created in the default VPC of that region, a load
// balancer in another region can't reach them.
func checkRegion(lb *godo.LoadBalancer, region string) error {
	if region == "" || lb.Region == nil || lb.Region.Slug == region {
		return nil
	}

	return fmt.Errorf("load balancer %q is in region %q, but the control plane nodeSets use region %q", lb.Name, lb.Region.Slug, region)
}

// checkDefaultVPC fails if the load balancer isn't in the default VPC of its
// region. machine-controller always creates droplets in the default VPC, and a
// load balancer only reaches droplets in its own VPC.
func checkDefaultVPC(ctx context.Context, vpcs vpcGetter, lb *godo.LoadBalancer) error {
	// Load balancers created before VPCs existed use the default network
	if lb.VPCUUID == "" {
		return nil
	}

	vpc, _, err := vpcs.Get(ctx, lb.VPCUUID)
	if err != nil {
		return fmt.Errorf("getting VPC %q of load balancer %q: %w", lb.VPCUUID, lb.Name, err)
	}

	if !vpc.Default {
		return fmt.Errorf("load balancer %q is in VPC %q (%s), but control plane droplets are created in the default VPC of region %q; use a load balancer in the default VPC",
			lb.Name, vpc.Name, vpc.ID, vpc.RegionSlug)
	}

	return nil
}

func lookupLoadBalancer(ctx context.Context, client godo.LoadBalancersService, cluster *kubeoneapi.KubeOneCluster) (*godo.LoadBalancer, error) {
	spec := lbSpec(cluster)
	if spec.ID != "" {
		return getLoadBalancer(ctx, client, spec.ID)
	}

	lb, err := findLoadBalancerByName(ctx, client, spec.Name, spec.Region, cluster.Name)
	if errors.Is(err, errLBNotFound) {
		return nil, fail.Cloud(fmt.Errorf("no load balancer found with name %q in region %q", spec.Name, spec.Region), "digitalocean", "looking up load balancer")
	}

	return lb, err
}

func cleanupLoadBalancer(ctx context.Context, client godo.LoadBalancersService, cluster *kubeoneapi.KubeOneCluster, logger logrus.FieldLogger) error {
	spec := lbSpec(cluster)
	name := spec.Name

	lb, err := findLoadBalancerByName(ctx, client, name, spec.Region, cluster.Name)
	if errors.Is(err, errLBNotFound) {
		logger.Debugf("no load balancer %q found, skipping deletion", name)

		return nil
	}
	if err != nil {
		return err
	}

	if !ownedByKubeOne(lb, cluster.Name) {
		logger.Warnf("load balancer %q doesn't target droplets tagged %q, it's not owned by KubeOne, skipping deletion", name, ClusterTag(cluster.Name))

		return nil
	}

	// LookupLoadBalancer sets .apiEndpoint.host to the load balancer IP. A
	// different host was provided by the user, in which case KubeOne doesn't
	// manage any load balancer.
	if host := cluster.APIEndpoint.Host; host != "" && host != lb.IP {
		logger.Infof("apiEndpoint.host %q is not load balancer %q, skipping deletion", host, name)

		return nil
	}

	logger.Debugf("deleting load balancer %q with id: %s", name, lb.ID)

	resp, err := client.Delete(ctx, lb.ID)
	if err != nil && !isNotFound(resp) {
		return fail.Cloud(err, "digitalocean", "deleting load balancer")
	}

	return nil
}

func createRequest(cluster *kubeoneapi.KubeOneCluster) *godo.LoadBalancerRequest {
	spec := lbSpec(cluster)

	return &godo.LoadBalancerRequest{
		Name:            spec.Name,
		Region:          spec.Region,
		SizeUnit:        spec.SizeUnit,
		Tag:             ClusterTag(cluster.Name),
		ForwardingRules: []godo.ForwardingRule{kubeAPIForwardingRule()},
		HealthCheck:     kubeAPIHealthCheck(),
	}
}

func kubeAPIHealthCheck() *godo.HealthCheck {
	return &godo.HealthCheck{
		Protocol:               "tcp",
		Port:                   kubeAPIPort,
		CheckIntervalSeconds:   3,
		ResponseTimeoutSeconds: 3,
		HealthyThreshold:       2,
		UnhealthyThreshold:     3,
	}
}

func isKubeAPIHealthCheck(hc *godo.HealthCheck) bool {
	return hc != nil && hc.Protocol == "tcp" && hc.Port == kubeAPIPort
}

func kubeAPIForwardingRule() godo.ForwardingRule {
	return godo.ForwardingRule{
		EntryProtocol:  "tcp",
		EntryPort:      kubeAPIPort,
		TargetProtocol: "tcp",
		TargetPort:     kubeAPIPort,
	}
}

// hasKubeAPIForwardingRule reports whether the load balancer forwards TCP 6443
// to the droplets. A rule on entry port 6443 that forwards elsewhere is an
// error, as DigitalOcean doesn't allow a second rule on the same entry port.
func hasKubeAPIForwardingRule(lb *godo.LoadBalancer) (bool, error) {
	want := kubeAPIForwardingRule()

	for _, rule := range lb.ForwardingRules {
		if rule.EntryPort != kubeAPIPort {
			continue
		}
		if rule.EntryProtocol != want.EntryProtocol || rule.TargetProtocol != want.TargetProtocol || rule.TargetPort != want.TargetPort {
			return false, fmt.Errorf("load balancer %q forwards %s/%d to %s/%d, kube-apiserver requires %s/%d to %s/%d",
				lb.Name, rule.EntryProtocol, rule.EntryPort, rule.TargetProtocol, rule.TargetPort,
				want.EntryProtocol, want.EntryPort, want.TargetProtocol, want.TargetPort)
		}

		return true, nil
	}

	return false, nil
}

// waitForLoadBalancer waits until the load balancer is active and has an IP
// address assigned, DigitalOcean assigns it only after provisioning. If
// justCreated is true, the load balancer not being found is retried, as it
// might not be visible yet.
func waitForLoadBalancer(ctx context.Context, client godo.LoadBalancersService, lb *godo.LoadBalancer, justCreated bool, logger logrus.FieldLogger) (*godo.LoadBalancer, error) {
	if lb.Status == lbStatusUp && lb.IP != "" {
		return lb, nil
	}

	logger.Infof("waiting for load balancer %q to become active...", lb.Name)

	var current *godo.LoadBalancer
	err := wait.PollUntilContextTimeout(ctx, lbPollInterval, lbPollTimeout, true, func(ctx context.Context) (bool, error) {
		got, resp, err := client.Get(ctx, lb.ID)
		if err != nil {
			if isPermanentError(resp, justCreated) {
				return false, err
			}
			logger.Debugf("failed to get load balancer %q: %v", lb.ID, err)

			return false, nil
		}
		current = got

		if got.Status == lbStatusErr {
			return false, fmt.Errorf("load balancer is in %q state", lbStatusErr)
		}

		return got.Status == lbStatusUp && got.IP != "", nil
	})
	if err != nil {
		return nil, fail.Cloud(fmt.Errorf("load balancer %q did not become active: %w", lb.Name, err), "digitalocean", "waiting for load balancer")
	}

	return current, nil
}

func getLoadBalancer(ctx context.Context, client godo.LoadBalancersService, id string) (*godo.LoadBalancer, error) {
	lb, resp, err := client.Get(ctx, id)
	if err != nil {
		if isNotFound(resp) {
			return nil, fail.Cloud(fmt.Errorf("no load balancer found with id: %s", id), "digitalocean", "looking up load balancer")
		}

		return nil, fail.Cloud(err, "digitalocean", "getting load balancer %q", id)
	}

	return lb, nil
}

// findLoadBalancerByName lists all load balancers and returns the one with the
// given name in the given region (any region if empty). The same name can be
// used in different regions (e.g. by clusters with the same name), and
// DigitalOcean doesn't enforce unique names, so if several load balancers
// match, the one owned by the cluster is preferred. The DigitalOcean API
// doesn't support filtering by name.
func findLoadBalancerByName(ctx context.Context, client godo.LoadBalancersService, name, region, clusterName string) (*godo.LoadBalancer, error) {
	opts := &godo.ListOptions{Page: 1, PerPage: lbPageSize}

	var firstMatch *godo.LoadBalancer
	for {
		lbs, resp, err := client.List(ctx, opts)
		if err != nil {
			return nil, fail.Cloud(err, "digitalocean", "listing load balancers")
		}

		for i := range lbs {
			if lbs[i].Name != name || (region != "" && (lbs[i].Region == nil || lbs[i].Region.Slug != region)) {
				continue
			}
			if ownedByKubeOne(&lbs[i], clusterName) {
				return &lbs[i], nil
			}
			if firstMatch == nil {
				firstMatch = &lbs[i]
			}
		}

		if resp == nil || resp.Links == nil || resp.Links.IsLastPage() {
			if firstMatch == nil {
				return nil, errLBNotFound
			}

			return firstMatch, nil
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, fail.Cloud(err, "digitalocean", "paginating load balancers")
		}
		opts.Page = page + 1
	}
}

func isNotFound(resp *godo.Response) bool {
	return resp != nil && resp.Response != nil && resp.StatusCode == http.StatusNotFound
}

// isPermanentError reports whether retrying the request can't succeed: a bad
// request, missing permissions, or the load balancer not existing (unless
// retryNotFound is true, it might not be visible yet right after it's
// created). Other errors (e.g. 408, 409, 429, 5xx) might be temporary.
func isPermanentError(resp *godo.Response, retryNotFound bool) bool {
	if resp == nil || resp.Response == nil {
		return false
	}

	switch resp.StatusCode {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden:
		return true
	case http.StatusNotFound:
		return !retryNotFound
	default:
		return false
	}
}

// describeBackends describes how the load balancer selects its backends.
func describeBackends(lb *godo.LoadBalancer) string {
	switch {
	case lb.Tag != "":
		return fmt.Sprintf("droplets tagged %q", lb.Tag)
	case len(lb.DropletIDs) > 0:
		return fmt.Sprintf("%d droplets by ID", len(lb.DropletIDs))
	default:
		return "no droplets"
	}
}

// ownedByKubeOne reports whether the load balancer was created by KubeOne.
// DigitalOcean doesn't persist tags on load balancers, so ownership is derived
// from the backend selector tag. KubeOne also sets that selector on a load
// balancer referenced by ID, but such a load balancer is never deleted as long
// as it's referenced by ID.
func ownedByKubeOne(lb *godo.LoadBalancer, clusterName string) bool {
	return lb.Tag == ClusterTag(clusterName)
}
