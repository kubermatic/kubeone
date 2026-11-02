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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v3"

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

	cred, err := azidentity.NewClientSecretCredential(
		env[credentials.AzureTenantID],
		env[credentials.AzureClientID],
		env[credentials.AzureClientSecret],
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating azure credential: %w", err)
	}

	subscriptionID := env[credentials.AzureSubscriptionID]

	vms, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating azure virtual machines client: %w", err)
	}

	networkClients, err := armnetwork.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating azure network clients: %w", err)
	}

	groups, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating azure resource groups client: %w", err)
	}

	return &azureCloud{
		vms:           vms,
		lbs:           networkClients.NewLoadBalancersClient(),
		ips:           networkClients.NewPublicIPAddressesClient(),
		groups:        groups,
		resourceGroup: resourceGroup,
		clusterName:   clusterName,
	}, nil
}

// azureCloud looks up the resources of the cluster in the resource group
// created by examples/terraform/azure-managed-control-plane. The resource
// group is dedicated to the cluster, the janitor deletes it as a whole.
type azureCloud struct {
	vms           *armcompute.VirtualMachinesClient
	lbs           *armnetwork.LoadBalancersClient
	ips           *armnetwork.PublicIPAddressesClient
	groups        *armresources.ResourceGroupsClient
	resourceGroup string
	clusterName   string
}

func (a *azureCloud) ControlPlaneVMs(ctx context.Context) (int, error) {
	count := 0
	pager := a.vms.NewListPager(a.resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if azureNotFound(err) {
			return 0, nil
		}
		if err != nil {
			return 0, fmt.Errorf("listing virtual machines in resource group %q: %w", a.resourceGroup, err)
		}

		for _, vm := range page.Value {
			if azureTag(vm.Tags, azureClusterNameTag) == a.clusterName && azureTag(vm.Tags, azureRoleTag) == azureControlPlane {
				count++
			}
		}
	}

	return count, nil
}

func (a *azureCloud) LoadBalancers(ctx context.Context) ([]string, error) {
	lb, err := a.lbs.Get(ctx, a.resourceGroup, azureLoadBalancerName(a.clusterName), nil)
	if azureNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting load balancer %q: %w", azureLoadBalancerName(a.clusterName), err)
	}

	var addresses []string
	if lb.Properties == nil {
		return addresses, nil
	}

	for _, frontend := range lb.Properties.FrontendIPConfigurations {
		if frontend == nil || frontend.Properties == nil ||
			frontend.Properties.PublicIPAddress == nil || frontend.Properties.PublicIPAddress.ID == nil {
			continue
		}

		// the frontend references the public IP by ID only
		id := *frontend.Properties.PublicIPAddress.ID
		ip, err := a.ips.Get(ctx, a.resourceGroup, id[strings.LastIndex(id, "/")+1:], nil)
		if err != nil {
			return nil, fmt.Errorf("getting public IP %q: %w", id, err)
		}
		if ip.Properties != nil && ip.Properties.IPAddress != nil {
			addresses = append(addresses, *ip.Properties.IPAddress)
		}
	}

	return addresses, nil
}

// LeftoverResources returns the public IP of the load balancer, it's a
// separate resource that outlives the load balancer if not deleted.
func (a *azureCloud) LeftoverResources(ctx context.Context) ([]string, error) {
	name := azureLoadBalancerName(a.clusterName) + "-pubip"

	_, err := a.ips.Get(ctx, a.resourceGroup, name, nil)
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
	poller, err := a.groups.BeginDelete(ctx, a.resourceGroup, nil)
	if azureNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting resource group %q: %w", a.resourceGroup, err)
	}

	if _, err = poller.PollUntilDone(ctx, nil); err != nil {
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
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		return respErr.StatusCode == http.StatusNotFound
	}

	return false
}
