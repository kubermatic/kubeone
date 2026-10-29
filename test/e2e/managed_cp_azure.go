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
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"     //nolint:staticcheck
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"     //nolint:staticcheck
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2021-04-01/resources" //nolint:staticcheck
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"

	"k8c.io/kubeone/pkg/credentials"
)

const (
	// tags applied by KubeOne to the control plane VMs, see
	// pkg/cloudprovider/azure. They're not taken from the KubeOne code, so
	// that a change of the markers is caught instead of followed.
	azureClusterNameTag = "kubeone_cluster_name"
	azureRoleTag        = "kubeone_role"
	azureControlPlane   = "control-plane"
)

// azureLoadBalancerName is the default name of the load balancer KubeOne
// creates, its public IP is named "<load balancer>-pubip".
func azureLoadBalancerName(clusterName string) string {
	return clusterName + "-kubeapi"
}

func newAzureCloud(clusterName string, tfOutput map[string]any) (*azureCloud, error) {
	networking, _ := tfOutput["networking"].(map[string]any)

	resourceGroup, _ := networking["resourceGroup"].(string)
	if resourceGroup == "" {
		return nil, errors.New("resourceGroup is missing from the networking terraform output")
	}

	env := map[string]string{}
	for _, key := range []string{
		credentials.AzureClientID,
		credentials.AzureClientSecret,
		credentials.AzureTenantID,
		credentials.AzureSubscriptionID,
	} {
		env[key] = os.Getenv(key)
		if env[key] == "" {
			return nil, fmt.Errorf("%s is not set", key)
		}
	}

	authorizer, err := auth.NewClientCredentialsConfig(
		env[credentials.AzureClientID],
		env[credentials.AzureClientSecret],
		env[credentials.AzureTenantID],
	).Authorizer()
	if err != nil {
		return nil, fmt.Errorf("creating azure authorizer: %w", err)
	}

	subscriptionID := env[credentials.AzureSubscriptionID]

	vms := compute.NewVirtualMachinesClient(subscriptionID)
	vms.Authorizer = authorizer

	lbs := network.NewLoadBalancersClient(subscriptionID)
	lbs.Authorizer = authorizer

	ips := network.NewPublicIPAddressesClient(subscriptionID)
	ips.Authorizer = authorizer

	groups := resources.NewGroupsClient(subscriptionID)
	groups.Authorizer = authorizer

	return &azureCloud{
		vms:           vms,
		lbs:           lbs,
		ips:           ips,
		groups:        groups,
		resourceGroup: resourceGroup,
		clusterName:   clusterName,
	}, nil
}

// azureCloud looks up the resources of the cluster in the resource group
// created by examples/terraform/azure-managed-control-plane. The resource
// group is dedicated to the cluster, the janitor deletes it as a whole.
type azureCloud struct {
	vms           compute.VirtualMachinesClient
	lbs           network.LoadBalancersClient
	ips           network.PublicIPAddressesClient
	groups        resources.GroupsClient
	resourceGroup string
	clusterName   string
}

func (a *azureCloud) ControlPlaneVMs(ctx context.Context) (int, error) {
	page, err := a.vms.List(ctx, a.resourceGroup, "")
	if azureNotFound(err) {
		return 0, nil
	}

	count := 0
	for ; err == nil && page.NotDone(); err = page.NextWithContext(ctx) {
		for _, vm := range page.Values() {
			if azureTag(vm.Tags, azureClusterNameTag) == a.clusterName && azureTag(vm.Tags, azureRoleTag) == azureControlPlane {
				count++
			}
		}
	}
	if err != nil {
		return 0, fmt.Errorf("listing virtual machines in resource group %q: %w", a.resourceGroup, err)
	}

	return count, nil
}

func (a *azureCloud) LoadBalancers(ctx context.Context) ([]string, error) {
	lb, err := a.lbs.Get(ctx, a.resourceGroup, azureLoadBalancerName(a.clusterName), "")
	if azureNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting load balancer %q: %w", azureLoadBalancerName(a.clusterName), err)
	}

	var addresses []string
	if lb.FrontendIPConfigurations == nil {
		return addresses, nil
	}

	for _, frontend := range *lb.FrontendIPConfigurations {
		if frontend.PublicIPAddress == nil || frontend.PublicIPAddress.ID == nil {
			continue
		}

		// the frontend references the public IP by ID only
		id := *frontend.PublicIPAddress.ID
		ip, err := a.ips.Get(ctx, a.resourceGroup, id[strings.LastIndex(id, "/")+1:], "")
		if err != nil {
			return nil, fmt.Errorf("getting public IP %q: %w", id, err)
		}
		if ip.IPAddress != nil {
			addresses = append(addresses, *ip.IPAddress)
		}
	}

	return addresses, nil
}

// LeftoverResources returns the public IP of the load balancer, it's a
// separate resource that outlives the load balancer if not deleted.
func (a *azureCloud) LeftoverResources(ctx context.Context) ([]string, error) {
	name := azureLoadBalancerName(a.clusterName) + "-pubip"

	_, err := a.ips.Get(ctx, a.resourceGroup, name, "")
	if azureNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting public IP %q: %w", name, err)
	}

	return []string{"public IP " + name}, nil
}

// Janitor deletes the whole resource group, including the VMs, their disks,
// NICs and public IPs, and the load balancers created by KubeOne and by the
// CCM. terraform destroy then finds nothing left to delete.
func (a *azureCloud) Janitor(ctx context.Context) error {
	future, err := a.groups.Delete(ctx, a.resourceGroup, "")
	if azureNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting resource group %q: %w", a.resourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, a.groups.Client); err != nil {
		return fmt.Errorf("waiting for deletion of resource group %q: %w", a.resourceGroup, err)
	}

	return nil
}

func azureTag(tags map[string]*string, key string) string {
	if v := tags[key]; v != nil {
		return *v
	}

	return ""
}

func azureNotFound(err error) bool {
	var detailed autorest.DetailedError
	if errors.As(err, &detailed) {
		return detailed.Response != nil && detailed.Response.StatusCode == http.StatusNotFound
	}

	return false
}
