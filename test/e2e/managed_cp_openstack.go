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

package e2e

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	"github.com/gophercloud/gophercloud/pagination"

	"k8c.io/kubeone/pkg/credentials"
)

const (
	// openstackClusterNameTag is set to the cluster name in the metadata of
	// the control plane servers by KubeOne. It's not taken from the KubeOne
	// code, so that a change of the marker is caught instead of followed.
	openstackClusterNameTag = "kubeone_cluster_name"

	// openstackWorkersTag is the metadata applied to worker servers by the
	// managed_cp_openstack*.yaml manifests, used by the janitor.
	openstackWorkersTag = "kubeone_workers"
)

// openstackCloud finds the control plane servers by the metadata KubeOne sets
// and the load balancer by the name the Terraform configs give it. KubeOne
// doesn't create the load balancer, it adds the control plane servers to its
// pool and deletes it on reset with --destroy-control-plane-load-balancer.
type openstackCloud struct {
	compute     *gophercloud.ServiceClient
	lb          *gophercloud.ServiceClient
	clusterName string
	lbName      string
	subnetID    string
}

func newOpenstackCloud(clusterName string, tfOutput map[string]any) (*openstackCloud, error) {
	loadbalancer, _ := tfOutput["loadbalancer"].(map[string]any)
	lbName, _ := loadbalancer["name"].(string)
	if lbName == "" {
		return nil, errors.New("name is missing from the loadbalancer terraform output")
	}

	networking, _ := tfOutput["networking"].(map[string]any)
	subnetID, _ := networking["subnetID"].(string)
	if subnetID == "" {
		return nil, errors.New("subnetID is missing from the networking terraform output")
	}

	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, fmt.Errorf("reading OpenStack credentials: %w", err)
	}

	// the client lives for the whole test, which outlasts the Keystone token,
	// so the cleanups would fail with "Authentication failed" without reauth
	opts.AllowReauth = true

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, fmt.Errorf("authenticating to OpenStack: %w", err)
	}

	endpoint := gophercloud.EndpointOpts{Region: os.Getenv(credentials.OpenStackRegionName)}

	computeClient, err := openstack.NewComputeV2(provider, endpoint)
	if err != nil {
		return nil, fmt.Errorf("creating compute client: %w", err)
	}

	lbClient, err := openstack.NewLoadBalancerV2(provider, endpoint)
	if err != nil {
		return nil, fmt.Errorf("creating load balancer client: %w", err)
	}

	return &openstackCloud{
		compute:     computeClient,
		lb:          lbClient,
		clusterName: clusterName,
		lbName:      lbName,
		subnetID:    subnetID,
	}, nil
}

func (o *openstackCloud) ControlPlaneVMs(context.Context) (int, error) {
	vms, err := o.serversByTag(openstackClusterNameTag)

	return len(vms), err
}

func (o *openstackCloud) LoadBalancers(context.Context) ([]string, error) {
	lbs, err := o.loadBalancers(loadbalancers.ListOpts{Name: o.lbName})
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, lb := range lbs {
		ips = append(ips, lb.VipAddress)
	}

	return ips, nil
}

func (o *openstackCloud) HealthyLoadBalancerMembers(context.Context) (int, error) {
	lbs, err := o.loadBalancers(loadbalancers.ListOpts{Name: o.lbName})
	if err != nil {
		return 0, err
	}

	healthy := 0
	for _, lb := range lbs {
		allPools, err := listAll(pools.List(o.lb, pools.ListOpts{LoadbalancerID: lb.ID}), pools.ExtractPools)
		if err != nil {
			return 0, fmt.Errorf("listing pools of load balancer %q: %w", lb.Name, err)
		}

		for _, pool := range allPools {
			members, err := listAll(pools.ListMembers(o.lb, pool.ID, pools.ListMembersOpts{}), pools.ExtractMembers)
			if err != nil {
				return 0, fmt.Errorf("listing members of pool %q: %w", pool.Name, err)
			}

			for _, member := range members {
				if member.OperatingStatus == "ONLINE" {
					healthy++
				}
			}
		}
	}

	return healthy, nil
}

// Janitor deletes the servers of the cluster and all the load balancers on its
// subnet, which includes the ones created by the OpenStack CCM for Services,
// and waits for them to be gone, as they block deleting the subnet.
func (o *openstackCloud) Janitor(ctx context.Context) error {
	var errs []error

	for _, tag := range []string{openstackClusterNameTag, openstackWorkersTag} {
		vms, err := o.serversByTag(tag)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		for _, vm := range vms {
			if err = servers.Delete(o.compute, vm.ID).ExtractErr(); err != nil && !isNotFound(err) {
				errs = append(errs, fmt.Errorf("deleting server %q: %w", vm.Name, err))
			}
		}
	}

	lbs, err := o.loadBalancers(loadbalancers.ListOpts{VipSubnetID: o.subnetID})
	if err != nil {
		errs = append(errs, err)
	}

	for _, lb := range lbs {
		if lb.ProvisioningStatus == "PENDING_DELETE" {
			continue
		}

		err = loadbalancers.Delete(o.lb, lb.ID, loadbalancers.DeleteOpts{Cascade: true}).ExtractErr()
		if err != nil && !isNotFound(err) {
			errs = append(errs, fmt.Errorf("deleting load balancer %q: %w", lb.Name, err))
		}
	}

	if err = errors.Join(errs...); err != nil {
		return err
	}

	return o.waitForJanitor(ctx)
}

func (o *openstackCloud) waitForJanitor(ctx context.Context) error {
	for {
		var left []string

		for _, tag := range []string{openstackClusterNameTag, openstackWorkersTag} {
			vms, err := o.serversByTag(tag)
			if err != nil {
				return err
			}

			for _, vm := range vms {
				left = append(left, "server "+vm.Name)
			}
		}

		lbs, err := o.loadBalancers(loadbalancers.ListOpts{VipSubnetID: o.subnetID})
		if err != nil {
			return err
		}

		for _, lb := range lbs {
			left = append(left, "load balancer "+lb.Name)
		}

		if len(left) == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for deletion of %v: %w", left, ctx.Err())
		case <-time.After(15 * time.Second):
		}
	}
}

// serversByTag returns the servers with the given metadata set to the cluster
// name. Servers are filtered client-side, so only exact matches are returned.
func (o *openstackCloud) serversByTag(tag string) ([]servers.Server, error) {
	all, err := listAll(servers.List(o.compute, servers.ListOpts{}), servers.ExtractServers)
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}

	var matching []servers.Server
	for _, vm := range all {
		if vm.Metadata[tag] == o.clusterName {
			matching = append(matching, vm)
		}
	}

	return matching, nil
}

func (o *openstackCloud) loadBalancers(opts loadbalancers.ListOpts) ([]loadbalancers.LoadBalancer, error) {
	lbs, err := listAll(loadbalancers.List(o.lb, opts), loadbalancers.ExtractLoadBalancers)
	if err != nil {
		return nil, fmt.Errorf("listing load balancers: %w", err)
	}

	return lbs, nil
}

func listAll[T any](pager pagination.Pager, extract func(pagination.Page) ([]T, error)) ([]T, error) {
	page, err := pager.AllPages()
	if err != nil {
		return nil, err
	}

	return extract(page)
}

func isNotFound(err error) bool {
	var notFound gophercloud.ErrDefault404

	return errors.As(err, &notFound)
}
