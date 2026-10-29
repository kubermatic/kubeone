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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"k8c.io/kubeone/pkg/credentials"
)

const (
	// gceWorkersLabel is the label applied to worker instances by the
	// managed_cp_gce*.yaml manifests, used by the janitor.
	gceWorkersLabel = "kubeone_workers"

	// gceClusterNameLabel is set to the cluster name on the control plane
	// instances by KubeOne. It's not taken from the KubeOne code, so that a
	// change of the marker is caught instead of followed.
	gceClusterNameLabel = "kubeone_cluster_name"
)

// gceOwnerDescription is the description KubeOne sets on the load balancer
// resources and firewall rules of the cluster.
func gceOwnerDescription(clusterName string) string {
	return "kubeone:" + clusterName
}

func newGCECloud(ctx context.Context, clusterName string) (*gceCloud, error) {
	serviceAccount := os.Getenv(credentials.GoogleServiceAccountKey)
	if serviceAccount == "" {
		return nil, fmt.Errorf("%s is not set", credentials.GoogleServiceAccountKey)
	}

	var key struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal([]byte(serviceAccount), &key); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", credentials.GoogleServiceAccountKey, err)
	}
	if key.ProjectID == "" {
		return nil, fmt.Errorf("%s has no project_id", credentials.GoogleServiceAccountKey)
	}

	svc, err := compute.NewService(ctx, option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(serviceAccount)))
	if err != nil {
		return nil, fmt.Errorf("creating compute client: %w", err)
	}

	return &gceCloud{
		svc:         svc,
		project:     key.ProjectID,
		clusterName: clusterName,
	}, nil
}

// gceCloud finds the resources of the cluster by the markers the GCE
// provider sets: the cluster label on control plane instances and the owner
// description on load balancer resources and firewall rules. Lists are
// filtered client-side, so only exact matches are ever deleted.
type gceCloud struct {
	svc         *compute.Service
	project     string
	clusterName string
}

func (g *gceCloud) ControlPlaneVMs(ctx context.Context) (int, error) {
	instances, err := g.instancesByLabel(ctx, gceClusterNameLabel)

	return len(instances), err
}

func (g *gceCloud) LoadBalancers(ctx context.Context) ([]string, error) {
	owner := gceOwnerDescription(g.clusterName)

	var ips []string
	err := g.svc.ForwardingRules.AggregatedList(g.project).Pages(ctx, func(page *compute.ForwardingRuleAggregatedList) error {
		for _, scoped := range page.Items {
			for _, fr := range scoped.ForwardingRules {
				if fr.Description == owner {
					ips = append(ips, fr.IPAddress)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing forwarding rules: %w", err)
	}

	return ips, nil
}

func (g *gceCloud) HealthyLoadBalancerMembers(ctx context.Context) (int, error) {
	owner := gceOwnerDescription(g.clusterName)
	healthy := 0

	var services []*compute.BackendService
	err := g.svc.BackendServices.AggregatedList(g.project).Pages(ctx, func(page *compute.BackendServiceAggregatedList) error {
		for _, scoped := range page.Items {
			for _, bs := range scoped.BackendServices {
				if bs.Description == owner && bs.Region != "" {
					services = append(services, bs)
				}
			}
		}

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("listing backend services: %w", err)
	}

	for _, bs := range services {
		for _, backend := range bs.Backends {
			health, err := g.svc.RegionBackendServices.GetHealth(g.project, path.Base(bs.Region), bs.Name, &compute.ResourceGroupReference{
				Group: backend.Group,
			}).Context(ctx).Do()
			if err != nil {
				return 0, fmt.Errorf("getting health of backend service %q: %w", bs.Name, err)
			}

			for _, status := range health.HealthStatus {
				if status.HealthState == "HEALTHY" {
					healthy++
				}
			}
		}
	}

	return healthy, nil
}

func (g *gceCloud) LeftoverResources(ctx context.Context) ([]string, error) {
	resources, err := g.ownedLoadBalancerResources(ctx)

	var names []string
	for _, r := range resources {
		names = append(names, r.kind+" "+r.name)
	}

	return names, err
}

// Janitor deletes the instances first and then the load balancer resources,
// in reverse order of their dependencies, waiting for each step.
func (g *gceCloud) Janitor(ctx context.Context) error {
	var errs []error

	for _, label := range []string{gceClusterNameLabel, gceWorkersLabel} {
		instances, err := g.instancesByLabel(ctx, label)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		for _, instance := range instances {
			errs = append(errs, g.delete(ctx, "instance", instance.Name, func() (*compute.Operation, error) {
				return g.svc.Instances.Delete(g.project, path.Base(instance.Zone), instance.Name).Context(ctx).Do()
			}))
		}
	}

	resources, err := g.ownedLoadBalancerResources(ctx)
	if err != nil {
		errs = append(errs, err)
	}

	for _, r := range resources {
		errs = append(errs, g.delete(ctx, r.kind, r.name, r.delete))
	}

	return errors.Join(errs...)
}

func (g *gceCloud) instancesByLabel(ctx context.Context, label string) ([]*compute.Instance, error) {
	var instances []*compute.Instance

	err := g.svc.Instances.AggregatedList(g.project).Pages(ctx, func(page *compute.InstanceAggregatedList) error {
		for _, scoped := range page.Items {
			for _, instance := range scoped.Instances {
				if instance.Labels[label] == g.clusterName {
					instances = append(instances, instance)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing instances labeled %s=%s: %w", label, g.clusterName, err)
	}

	return instances, nil
}

type gceResource struct {
	kind   string
	name   string
	delete func() (*compute.Operation, error)
}

// ownedLoadBalancerResources returns the owned resources in the order they
// have to be deleted.
func (g *gceCloud) ownedLoadBalancerResources(ctx context.Context) ([]gceResource, error) {
	owner := gceOwnerDescription(g.clusterName)
	var (
		resources []gceResource
		errs      []error
	)

	err := g.svc.ForwardingRules.AggregatedList(g.project).Pages(ctx, func(page *compute.ForwardingRuleAggregatedList) error {
		for _, scoped := range page.Items {
			for _, fr := range scoped.ForwardingRules {
				if fr.Description == owner {
					region, name := path.Base(fr.Region), fr.Name
					resources = append(resources, gceResource{"forwarding rule", name, func() (*compute.Operation, error) {
						return g.svc.ForwardingRules.Delete(g.project, region, name).Context(ctx).Do()
					}})
				}
			}
		}

		return nil
	})
	errs = append(errs, err)

	err = g.svc.BackendServices.AggregatedList(g.project).Pages(ctx, func(page *compute.BackendServiceAggregatedList) error {
		for _, scoped := range page.Items {
			for _, bs := range scoped.BackendServices {
				if bs.Description == owner && bs.Region != "" {
					region, name := path.Base(bs.Region), bs.Name
					resources = append(resources, gceResource{"backend service", name, func() (*compute.Operation, error) {
						return g.svc.RegionBackendServices.Delete(g.project, region, name).Context(ctx).Do()
					}})
				}
			}
		}

		return nil
	})
	errs = append(errs, err)

	err = g.svc.NetworkEndpointGroups.AggregatedList(g.project).Pages(ctx, func(page *compute.NetworkEndpointGroupAggregatedList) error {
		for _, scoped := range page.Items {
			for _, neg := range scoped.NetworkEndpointGroups {
				if neg.Description == owner && neg.Zone != "" {
					zone, name := path.Base(neg.Zone), neg.Name
					resources = append(resources, gceResource{"network endpoint group", name, func() (*compute.Operation, error) {
						return g.svc.NetworkEndpointGroups.Delete(g.project, zone, name).Context(ctx).Do()
					}})
				}
			}
		}

		return nil
	})
	errs = append(errs, err)

	err = g.svc.HealthChecks.AggregatedList(g.project).Pages(ctx, func(page *compute.HealthChecksAggregatedList) error {
		for _, scoped := range page.Items {
			for _, hc := range scoped.HealthChecks {
				if hc.Description == owner && hc.Region != "" {
					region, name := path.Base(hc.Region), hc.Name
					resources = append(resources, gceResource{"health check", name, func() (*compute.Operation, error) {
						return g.svc.RegionHealthChecks.Delete(g.project, region, name).Context(ctx).Do()
					}})
				}
			}
		}

		return nil
	})
	errs = append(errs, err)

	err = g.svc.Addresses.AggregatedList(g.project).Pages(ctx, func(page *compute.AddressAggregatedList) error {
		for _, scoped := range page.Items {
			for _, addr := range scoped.Addresses {
				if addr.Description == owner && addr.Region != "" {
					region, name := path.Base(addr.Region), addr.Name
					resources = append(resources, gceResource{"address", name, func() (*compute.Operation, error) {
						return g.svc.Addresses.Delete(g.project, region, name).Context(ctx).Do()
					}})
				}
			}
		}

		return nil
	})
	errs = append(errs, err)

	err = g.svc.Firewalls.List(g.project).Pages(ctx, func(page *compute.FirewallList) error {
		for _, fw := range page.Items {
			if fw.Description == owner {
				name := fw.Name
				resources = append(resources, gceResource{"firewall rule", name, func() (*compute.Operation, error) {
					return g.svc.Firewalls.Delete(g.project, name).Context(ctx).Do()
				}})
			}
		}

		return nil
	})
	errs = append(errs, err)

	if err := errors.Join(errs...); err != nil {
		return resources, fmt.Errorf("listing load balancer resources: %w", err)
	}

	return resources, nil
}

// delete runs the delete call and waits for the operation, a resource that
// is already gone is not an error.
func (g *gceCloud) delete(ctx context.Context, kind, name string, del func() (*compute.Operation, error)) error {
	op, err := del()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("deleting %s %q: %w", kind, name, err)
	}

	for op.Status != "DONE" {
		switch {
		case op.Zone != "":
			op, err = g.svc.ZoneOperations.Wait(g.project, path.Base(op.Zone), op.Name).Context(ctx).Do()
		case op.Region != "":
			op, err = g.svc.RegionOperations.Wait(g.project, path.Base(op.Region), op.Name).Context(ctx).Do()
		default:
			op, err = g.svc.GlobalOperations.Wait(g.project, op.Name).Context(ctx).Do()
		}
		if err != nil {
			return fmt.Errorf("waiting for deletion of %s %q: %w", kind, name, err)
		}
	}

	if op.Error != nil && len(op.Error.Errors) > 0 {
		var msgs []string
		for _, e := range op.Error.Errors {
			msgs = append(msgs, e.Code+": "+e.Message)
		}

		return fmt.Errorf("deleting %s %q: %s", kind, name, strings.Join(msgs, "; "))
	}

	return nil
}
