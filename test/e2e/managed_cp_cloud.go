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
	"fmt"
)

// managedControlPlaneCloud inspects and cleans up the cloud resources created
// by KubeOne for a managed control plane, independently of KubeOne.
type managedControlPlaneCloud interface {
	// ControlPlaneVMs returns the number of control plane VMs of the cluster
	ControlPlaneVMs(ctx context.Context) (int, error)
	// LoadBalancers returns the addresses (IP or DNS name) of the load
	// balancers created by KubeOne
	LoadBalancers(ctx context.Context) ([]string, error)
	// Janitor deletes all leftover resources of the cluster
	Janitor(ctx context.Context) error
}

// loadBalancerMembersCloud is implemented by clouds where KubeOne adds the
// control plane VMs to the load balancer itself, e.g. GCE where the followers
// are added only after they joined the cluster.
type loadBalancerMembersCloud interface {
	// HealthyLoadBalancerMembers returns the number of control plane VMs that
	// are healthy members of the load balancers created by KubeOne
	HealthyLoadBalancerMembers(ctx context.Context) (int, error)
}

// leftoverResourcesCloud is implemented by clouds where the load balancer
// consists of several resources, it returns the ones of the cluster that
// still exist (e.g. "firewall rule k1-123-kubeapi-6443").
type leftoverResourcesCloud interface {
	LeftoverResources(ctx context.Context) ([]string, error)
}

// newManagedControlPlaneCloud returns the managedControlPlaneCloud of the
// provider. tfOutput holds the output values of the infrastructure Terraform
// configs, nil if the infrastructure has none.
func newManagedControlPlaneCloud(ctx context.Context, provider, clusterName string, tfOutput map[string]any) (managedControlPlaneCloud, error) {
	switch provider {
	case "aws":
		return newAWSCloud(ctx, clusterName, tfOutput)
	case "azure":
		return newAzureCloud(clusterName, tfOutput)
	case "digitalocean":
		return newDigitaloceanCloud(clusterName)
	case "gce":
		return newGCECloud(ctx, clusterName)
	case "hetzner":
		return newHetznerCloud(clusterName)
	default:
		return nil, fmt.Errorf("managed control plane is not supported for provider %q", provider)
	}
}
