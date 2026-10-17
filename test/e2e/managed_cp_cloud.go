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

	"github.com/digitalocean/godo"

	"k8c.io/kubeone/pkg/cloudprovider/digitalocean"
	"k8c.io/kubeone/pkg/credentials"
)

// managedControlPlaneCloud inspects and cleans up the cloud resources created
// by KubeOne for a managed control plane, independently of KubeOne.
type managedControlPlaneCloud interface {
	// ControlPlaneVMs returns the number of control plane VMs of the cluster
	ControlPlaneVMs(ctx context.Context) (int, error)
	// LoadBalancers returns the IP addresses of the load balancers created by KubeOne
	LoadBalancers(ctx context.Context) ([]string, error)
	// Janitor deletes all leftover resources of the cluster
	Janitor(ctx context.Context) error
}

func newManagedControlPlaneCloud(provider, clusterName string) (managedControlPlaneCloud, error) {
	if provider != "digitalocean" {
		return nil, fmt.Errorf("managed control plane is not supported for provider %q", provider)
	}

	token := os.Getenv(credentials.DigitalOceanTokenKey)
	if token == "" {
		return nil, fmt.Errorf("%s is not set", credentials.DigitalOceanTokenKey)
	}

	return &digitaloceanCloud{
		client:      godo.NewFromToken(token),
		clusterName: clusterName,
	}, nil
}

// digitaloceanWorkersTag is the tag applied to worker droplets by the
// managed_cp_digitalocean*.yaml manifests, used by the janitor.
func digitaloceanWorkersTag(clusterName string) string {
	return "kubeone-workers:" + clusterName
}

type digitaloceanCloud struct {
	client      *godo.Client
	clusterName string
}

func (do *digitaloceanCloud) ControlPlaneVMs(ctx context.Context) (int, error) {
	droplets, err := do.dropletsByTag(ctx, digitalocean.OwnedTag(do.clusterName))

	return len(droplets), err
}

func (do *digitaloceanCloud) LoadBalancers(ctx context.Context) ([]string, error) {
	lbs, err := do.ownedLoadBalancers(ctx)
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, lb := range lbs {
		ips = append(ips, lb.IP)
	}

	return ips, nil
}

func (do *digitaloceanCloud) Janitor(ctx context.Context) error {
	var errs []error

	for _, tag := range []string{digitalocean.OwnedTag(do.clusterName), digitaloceanWorkersTag(do.clusterName)} {
		droplets, err := do.dropletsByTag(ctx, tag)
		if err != nil {
			errs = append(errs, err)

			continue
		}
		if len(droplets) == 0 {
			continue
		}

		if _, err = do.client.Droplets.DeleteByTag(ctx, tag); err != nil {
			errs = append(errs, fmt.Errorf("deleting droplets tagged %q: %w", tag, err))
		}
	}

	lbs, err := do.ownedLoadBalancers(ctx)
	if err != nil {
		errs = append(errs, err)
	}

	for _, lb := range lbs {
		if _, err = do.client.LoadBalancers.Delete(ctx, lb.ID); err != nil {
			errs = append(errs, fmt.Errorf("deleting load balancer %q: %w", lb.Name, err))
		}
	}

	return errors.Join(errs...)
}

func (do *digitaloceanCloud) dropletsByTag(ctx context.Context, tag string) ([]godo.Droplet, error) {
	var (
		all  []godo.Droplet
		opts = &godo.ListOptions{Page: 1, PerPage: 200}
	)

	for {
		droplets, resp, err := do.client.Droplets.ListByTag(ctx, tag, opts)
		if err != nil {
			return nil, fmt.Errorf("listing droplets tagged %q: %w", tag, err)
		}
		all = append(all, droplets...)

		if resp == nil || resp.Links == nil || resp.Links.IsLastPage() {
			return all, nil
		}
		opts.Page++
	}
}

func (do *digitaloceanCloud) ownedLoadBalancers(ctx context.Context) ([]godo.LoadBalancer, error) {
	var (
		owned []godo.LoadBalancer
		// DigitalOcean doesn't persist tags on load balancers, the ones
		// created by KubeOne select their backends by the cluster tag
		tag  = digitalocean.ClusterTag(do.clusterName)
		opts = &godo.ListOptions{Page: 1, PerPage: 200}
	)

	for {
		lbs, resp, err := do.client.LoadBalancers.List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("listing load balancers: %w", err)
		}

		for _, lb := range lbs {
			if lb.Tag == tag {
				owned = append(owned, lb)
			}
		}

		if resp == nil || resp.Links == nil || resp.Links.IsLastPage() {
			return owned, nil
		}
		opts.Page++
	}
}
