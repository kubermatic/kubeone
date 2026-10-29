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

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"k8c.io/kubeone/pkg/cloudprovider/hetzner"
	"k8c.io/kubeone/pkg/credentials"

	"k8s.io/apimachinery/pkg/labels"
)

// hetznerClusterLabel is applied to all the servers (control plane and
// workers) of the cluster: KubeOne sets it on the control plane servers, the
// managed_cp_hetzner.yaml manifest on the workers. The firewall of
// examples/terraform/hetzner-managed-control-plane applies to it as well.
const hetznerClusterLabel = "kubeone_cluster_name"

func newHetznerCloud(clusterName string) (*hetznerCloud, error) {
	token := os.Getenv(credentials.HetznerTokenKey)
	if token == "" {
		return nil, fmt.Errorf("%s is not set", credentials.HetznerTokenKey)
	}

	return &hetznerCloud{
		client:      hcloud.NewClient(hcloud.WithToken(token)),
		clusterName: clusterName,
	}, nil
}

type hetznerCloud struct {
	client      *hcloud.Client
	clusterName string
}

func (hz *hetznerCloud) ControlPlaneVMs(ctx context.Context) (int, error) {
	servers, err := hz.servers(ctx, hetzner.ControlPlaneLabels(hz.clusterName))

	return len(servers), err
}

func (hz *hetznerCloud) LoadBalancers(ctx context.Context) ([]string, error) {
	lb, err := hz.loadBalancer(ctx)
	if err != nil || lb == nil {
		return nil, err
	}

	return []string{lb.PublicNet.IPv4.IP.String()}, nil
}

func (hz *hetznerCloud) Janitor(ctx context.Context) error {
	var errs []error

	servers, err := hz.servers(ctx, map[string]string{hetznerClusterLabel: hz.clusterName})
	if err != nil {
		errs = append(errs, err)
	}

	// The load balancers are deleted first, the ones created by the CCM are
	// found by their targets, which are gone once the servers are deleted
	lbs, err := hz.loadBalancers(ctx, servers)
	if err != nil {
		errs = append(errs, err)
	}

	for _, lb := range lbs {
		if _, err = hz.client.LoadBalancer.Delete(ctx, lb); err != nil {
			errs = append(errs, fmt.Errorf("deleting load balancer %q: %w", lb.Name, err))
		}
	}

	// the servers block terraform destroy of the network and the firewall
	// until they're gone
	var actions []*hcloud.Action
	for _, server := range servers {
		result, _, delErr := hz.client.Server.DeleteWithResult(ctx, server)
		if delErr != nil {
			errs = append(errs, fmt.Errorf("deleting server %q: %w", server.Name, delErr))

			continue
		}
		actions = append(actions, result.Action)
	}

	if err = hz.client.Action.WaitFor(ctx, actions...); err != nil {
		errs = append(errs, fmt.Errorf("waiting for servers to be deleted: %w", err))
	}

	return errors.Join(errs...)
}

func (hz *hetznerCloud) servers(ctx context.Context, selector map[string]string) ([]*hcloud.Server, error) {
	servers, err := hz.client.Server.AllWithOpts(ctx, hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: labels.SelectorFromSet(selector).String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}

	return servers, nil
}

// loadBalancer returns the load balancer created by KubeOne, nil if it doesn't
// exist. KubeOne looks it up by its name, the default one is used by the
// managed_cp_hetzner.yaml manifest.
func (hz *hetznerCloud) loadBalancer(ctx context.Context) (*hcloud.LoadBalancer, error) {
	name := hz.clusterName + "-kubeapi"

	lb, _, err := hz.client.LoadBalancer.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("getting load balancer %q: %w", name, err)
	}

	return lb, nil
}

// loadBalancers returns the load balancer created by KubeOne and the ones
// created by the CCM for the cluster, i.e. that target any of the servers.
func (hz *hetznerCloud) loadBalancers(ctx context.Context, servers []*hcloud.Server) ([]*hcloud.LoadBalancer, error) {
	clusterServers := map[int64]bool{}
	for _, server := range servers {
		clusterServers[server.ID] = true
	}

	all, err := hz.client.LoadBalancer.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing load balancers: %w", err)
	}

	owned, err := hz.loadBalancer(ctx)
	if err != nil {
		return nil, err
	}

	var lbs []*hcloud.LoadBalancer
	for _, lb := range all {
		if owned != nil && lb.ID == owned.ID {
			lbs = append(lbs, lb)

			continue
		}

		for _, target := range lb.Targets {
			if target.Type == hcloud.LoadBalancerTargetTypeServer && target.Server != nil && target.Server.Server != nil && clusterServers[target.Server.Server.ID] {
				lbs = append(lbs, lb)

				break
			}
		}
	}

	return lbs, nil
}
