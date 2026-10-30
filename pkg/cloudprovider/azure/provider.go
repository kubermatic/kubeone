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

package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/cloudprovider"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/provisioner"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	azuretypes "k8c.io/machine-controller/sdk/cloudprovider/azure"
	"k8c.io/machine-controller/sdk/jsonutil"
	"k8c.io/machine-controller/sdk/providerconfig"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	apiserverPort        = 6443
	frontendIPConfigName = "frontend"
	backendPoolName      = "backend"
	probeName            = "apiserver-probe"
	ruleName             = "apiserver-rule"
)

// cleanupFinalizers are the finalizers the machine-controller Azure driver
// adds to a Machine while creating its resources. Its Cleanup deletes only
// the resources whose finalizer the Machine carries, and the Machines KubeOne
// generates for reset don't remember the ones added on creation, so they're
// set upfront. Each deletion looks resources up by the machine UID tag, so
// the ones that don't exist are skipped.
var cleanupFinalizers = []string{
	"kubermatic.io/cleanup-azure-vm",
	"kubermatic.io/cleanup-azure-disks",
	"kubermatic.io/cleanup-azure-nic",
	"kubermatic.io/cleanup-azure-public-ip",
}

var (
	_ cloudprovider.ControlPlaneCloudProvider = &Provider{}
	_ cloudprovider.LoadBalancerProvider      = &Provider{}
	_ cloudprovider.CleanupPreparer           = &Provider{}
)

func init() {
	cloudprovider.Register(&Provider{})
}

type Provider struct{}

func (p *Provider) Name() string { return "Azure" }

func (p *Provider) Enabled(s *state.State) bool {
	return s.Cluster.CloudProvider.Azure != nil &&
		s.Cluster.CloudProvider.Azure.ControlPlane != nil &&
		len(s.Cluster.ControlPlane.NodeSets) > 0
}

func (p *Provider) MatchesConfig(cluster *kubeoneapi.KubeOneCluster) bool {
	return cluster.CloudProvider.Azure != nil
}

func (p *Provider) HasLoadBalancer(s *state.State) bool {
	return p.Enabled(s)
}

func (p *Provider) GenerateMachines(clusterName string, nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	return generateAzureControlPlaneMachines(clusterName, nodeSet, kubeletVersion)
}

func (p *Provider) EnsureVM(s *state.State, nodeSet kubeoneapi.NodeSet, capimachine clusterv1alpha1.Machine) error {
	if err := prepareAzureEnv(s); err != nil {
		return err
	}

	provMachines, err := provisioner.FindOrCreateMachines(s.Context, []clusterv1alpha1.Machine{capimachine}, s.Logger)
	if err != nil {
		return err
	}

	if err := p.registerNICToBackendPool(s, capimachine.Name); err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, cloudprovider.HostConfigsFromMachines(provMachines, []kubeoneapi.NodeSet{nodeSet})...)

	return nil
}

func (p *Provider) LookupVMs(s *state.State) error {
	if err := prepareAzureEnv(s); err != nil {
		return err
	}

	capimachines, err := p.GenerateMachines(
		s.Cluster.Name,
		s.Cluster.ControlPlane.NodeSets,
		s.Cluster.Versions.Kubernetes,
	)
	if err != nil {
		return err
	}

	provMachines, err := provisioner.FindMachines(s.Context, capimachines, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, cloudprovider.HostConfigsFromMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

	return nil
}

// PrepareCleanup satisfies cloudprovider.CleanupPreparer. CleanupVMs deletes
// the VMs through the machine-controller driver without going through
// EnsureVM or LookupVMs, so the credentials have to be exported here too.
func (p *Provider) PrepareCleanup(s *state.State) error {
	return prepareAzureEnv(s)
}

func (p *Provider) EnsureLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	cfg, err := resolveAzureLB(s)
	if err != nil {
		return err
	}

	lbClient, ipClient, _, err := azureNetworkClients(s)
	if err != nil {
		return err
	}

	publicIP, err := ensureAzurePublicIP(s.Context, ipClient, cfg)
	if err != nil {
		return err
	}

	if err = ensureAzureLoadBalancer(s.Context, lbClient, cfg, publicIP); err != nil {
		return err
	}

	host, err := azureLoadBalancerEndpoint(s.Context, ipClient, cfg)
	if err != nil {
		return err
	}

	s.Cluster.APIEndpoint.Host = host
	s.Cluster.APIEndpoint.Port = apiserverPort

	return nil
}

func (p *Provider) LookupLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	cfg, err := resolveAzureLB(s)
	if err != nil {
		return err
	}

	_, ipClient, _, err := azureNetworkClients(s)
	if err != nil {
		return err
	}

	host, err := azureLoadBalancerEndpoint(s.Context, ipClient, cfg)
	if err != nil {
		return err
	}

	s.Cluster.APIEndpoint.Host = host
	s.Cluster.APIEndpoint.Port = apiserverPort

	return nil
}

func (p *Provider) CleanupLoadBalancer(s *state.State) error {
	cfg, err := resolveAzureLB(s)
	if err != nil {
		return err
	}

	lbClient, ipClient, _, err := azureNetworkClients(s)
	if err != nil {
		return err
	}

	if err := deleteAzureLoadBalancer(s.Context, lbClient, cfg, s.Logger); err != nil {
		return err
	}

	if err := deleteAzurePublicIP(s.Context, ipClient, cfg, s.Logger); err != nil {
		return err
	}

	return nil
}

func (p *Provider) registerNICToBackendPool(s *state.State, machineName string) error {
	cfg, err := resolveAzureLB(s)
	if err != nil {
		return err
	}

	_, _, ifClient, err := azureNetworkClients(s)
	if err != nil {
		return err
	}

	nicName := machineName + "-netiface"

	return addNICToBackendPool(s.Context, ifClient, cfg.ResourceGroup, nicName, cfg.backendPoolID(), s.Logger)
}

type azureLBConfig struct {
	SubscriptionID string
	ResourceGroup  string
	Location       string
	Name           string
	SkuName        armnetwork.LoadBalancerSKUName
	PublicIPName   string
	lbID           string
}

func (c *azureLBConfig) backendPoolID() string {
	return c.lbID + "/backendAddressPools/" + backendPoolName
}

func (c *azureLBConfig) publicIPSkuName() armnetwork.PublicIPAddressSKUName {
	if c.SkuName == armnetwork.LoadBalancerSKUNameBasic {
		return armnetwork.PublicIPAddressSKUNameBasic
	}

	return armnetwork.PublicIPAddressSKUNameStandard
}

func resolveAzureLB(s *state.State) (*azureLBConfig, error) {
	azCP := s.Cluster.CloudProvider.Azure.ControlPlane
	lb := azCP.LoadBalancer

	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return nil, err
	}

	subscriptionID := providerCreds[credentials.AzureSubscriptionIDMC]
	resourceGroup := lb.ResourceGroup
	location := lb.Location

	if (resourceGroup == "" || location == "") && len(s.Cluster.ControlPlane.NodeSets) > 0 {
		var raw azuretypes.RawConfig
		if err := jsonutil.StrictUnmarshal(s.Cluster.ControlPlane.NodeSets[0].CloudProviderSpec, &raw); err != nil {
			return nil, fail.Config(err, "decode azure control plane config")
		}
		if resourceGroup == "" {
			resourceGroup = raw.ResourceGroup.Value
		}
		if location == "" {
			location = raw.Location.Value
		}
	}

	if resourceGroup == "" {
		return nil, fail.Config(errors.New("resourceGroup is empty"), "azure control plane loadBalancer.resourceGroup")
	}
	if location == "" {
		return nil, fail.Config(errors.New("location is empty"), "azure control plane loadBalancer.location")
	}

	skuName := armnetwork.LoadBalancerSKUNameStandard
	if lb.Sku == string(armnetwork.LoadBalancerSKUNameBasic) {
		skuName = armnetwork.LoadBalancerSKUNameBasic
	}

	return &azureLBConfig{
		SubscriptionID: subscriptionID,
		ResourceGroup:  resourceGroup,
		Location:       location,
		Name:           lb.Name,
		SkuName:        skuName,
		PublicIPName:   lb.PublicIPName,
		lbID: fmt.Sprintf(
			"/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s",
			subscriptionID,
			resourceGroup,
			lb.Name,
		),
	}, nil
}

func azureNetworkClients(s *state.State) (*armnetwork.LoadBalancersClient, *armnetwork.PublicIPAddressesClient, *armnetwork.InterfacesClient, error) {
	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return nil, nil, nil, err
	}

	clientID := providerCreds[credentials.AzureClientIDMC]
	clientSecret := providerCreds[credentials.AzureClientSecretMC]
	tenantID := providerCreds[credentials.AzureTenantIDMC]
	subscriptionID := providerCreds[credentials.AzureSubscriptionIDMC]

	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		return nil, nil, nil, fail.Runtime(err, "creating azure credential")
	}

	factory, err := armnetwork.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, nil, nil, fail.Runtime(err, "creating azure network clients")
	}

	return factory.NewLoadBalancersClient(), factory.NewPublicIPAddressesClient(), factory.NewInterfacesClient(), nil
}

func prepareAzureEnv(s *state.State) error {
	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return err
	}

	for _, key := range []string{
		credentials.AzureClientIDMC,
		credentials.AzureClientSecretMC,
		credentials.AzureTenantIDMC,
		credentials.AzureSubscriptionIDMC,
	} {
		if err := os.Setenv(key, providerCreds[key]); err != nil {
			return fail.Runtime(err, "setting %s", key)
		}
	}

	return nil
}

func ensureAzurePublicIP(ctx context.Context, ipClient *armnetwork.PublicIPAddressesClient, cfg *azureLBConfig) (*armnetwork.PublicIPAddress, error) {
	existing, err := ipClient.Get(ctx, cfg.ResourceGroup, cfg.PublicIPName, nil)
	if err == nil {
		return &existing.PublicIPAddress, nil
	}
	if !azureErrorNotFound(err) {
		return nil, fail.Cloud(err, "azure", "getting public IP %q", cfg.PublicIPName)
	}

	params := armnetwork.PublicIPAddress{
		Name:     new(cfg.PublicIPName),
		Location: new(cfg.Location),
		SKU: &armnetwork.PublicIPAddressSKU{
			Name: to.Ptr(cfg.publicIPSkuName()),
		},
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodStatic),
			PublicIPAddressVersion:   to.Ptr(armnetwork.IPVersionIPv4),
		},
	}

	poller, err := ipClient.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, cfg.PublicIPName, params, nil)
	if err != nil {
		return nil, fail.Cloud(err, "azure", "creating public IP %q", cfg.PublicIPName)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fail.Cloud(err, "azure", "waiting for public IP %q", cfg.PublicIPName)
	}

	return &result.PublicIPAddress, nil
}

func ensureAzureLoadBalancer(ctx context.Context, lbClient *armnetwork.LoadBalancersClient, cfg *azureLBConfig, publicIP *armnetwork.PublicIPAddress) error {
	_, err := lbClient.Get(ctx, cfg.ResourceGroup, cfg.Name, nil)
	if err == nil {
		return nil
	}
	if !azureErrorNotFound(err) {
		return fail.Cloud(err, "azure", "getting load balancer %q", cfg.Name)
	}

	frontendID := cfg.lbID + "/frontendIPConfigurations/" + frontendIPConfigName

	params := armnetwork.LoadBalancer{
		Name:     new(cfg.Name),
		Location: new(cfg.Location),
		SKU: &armnetwork.LoadBalancerSKU{
			Name: new(cfg.SkuName),
		},
		Properties: &armnetwork.LoadBalancerPropertiesFormat{
			FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
				{
					Name: new(frontendIPConfigName),
					Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
						PublicIPAddress: &armnetwork.PublicIPAddress{
							ID: publicIP.ID,
						},
					},
				},
			},
			BackendAddressPools: []*armnetwork.BackendAddressPool{
				{
					Name: new(backendPoolName),
				},
			},
			Probes: []*armnetwork.Probe{
				{
					Name: new(probeName),
					Properties: &armnetwork.ProbePropertiesFormat{
						Protocol:          to.Ptr(armnetwork.ProbeProtocolTCP),
						Port:              to.Ptr[int32](apiserverPort),
						IntervalInSeconds: to.Ptr[int32](5),
						NumberOfProbes:    to.Ptr[int32](2),
					},
				},
			},
			LoadBalancingRules: []*armnetwork.LoadBalancingRule{
				{
					Name: new(ruleName),
					Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
						FrontendIPConfiguration: &armnetwork.SubResource{ID: new(frontendID)},
						BackendAddressPool:      &armnetwork.SubResource{ID: new(cfg.backendPoolID())},
						Probe:                   &armnetwork.SubResource{ID: new(cfg.lbID + "/probes/" + probeName)},
						Protocol:                to.Ptr(armnetwork.TransportProtocolTCP),
						FrontendPort:            to.Ptr[int32](apiserverPort),
						BackendPort:             to.Ptr[int32](apiserverPort),
					},
				},
			},
		},
	}

	poller, err := lbClient.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, cfg.Name, params, nil)
	if err != nil {
		return fail.Cloud(err, "azure", "creating load balancer %q", cfg.Name)
	}
	if _, err := poller.PollUntilDone(ctx, nil); err != nil {
		return fail.Cloud(err, "azure", "waiting for load balancer %q", cfg.Name)
	}

	return nil
}

func deleteAzureLoadBalancer(
	ctx context.Context,
	lbClient *armnetwork.LoadBalancersClient,
	cfg *azureLBConfig,
	logger logrus.FieldLogger,
) error {
	if _, err := lbClient.Get(ctx, cfg.ResourceGroup, cfg.Name, nil); err != nil {
		if azureErrorNotFound(err) {
			logger.Debugf("no load balancer %q found, skipping deletion", cfg.Name)

			return nil
		}

		return fail.Cloud(err, "azure", "getting load balancer %q", cfg.Name)
	}

	logger.Debugf("deleting load balancer %q", cfg.Name)

	poller, err := lbClient.BeginDelete(ctx, cfg.ResourceGroup, cfg.Name, nil)
	if err != nil {
		return fail.Cloud(err, "azure", "deleting load balancer %q", cfg.Name)
	}
	if _, err := poller.PollUntilDone(ctx, nil); err != nil {
		return fail.Cloud(err, "azure", "waiting for load balancer %q deletion", cfg.Name)
	}

	return nil
}

func deleteAzurePublicIP(
	ctx context.Context,
	ipClient *armnetwork.PublicIPAddressesClient,
	cfg *azureLBConfig,
	logger logrus.FieldLogger,
) error {
	if _, err := ipClient.Get(ctx, cfg.ResourceGroup, cfg.PublicIPName, nil); err != nil {
		if azureErrorNotFound(err) {
			logger.Debugf("no public IP %q found, skipping deletion", cfg.PublicIPName)

			return nil
		}

		return fail.Cloud(err, "azure", "getting public IP %q", cfg.PublicIPName)
	}

	logger.Debugf("deleting public IP %q", cfg.PublicIPName)

	poller, err := ipClient.BeginDelete(ctx, cfg.ResourceGroup, cfg.PublicIPName, nil)
	if err != nil {
		return fail.Cloud(err, "azure", "deleting public IP %q", cfg.PublicIPName)
	}
	if _, err := poller.PollUntilDone(ctx, nil); err != nil {
		return fail.Cloud(err, "azure", "waiting for public IP %q deletion", cfg.PublicIPName)
	}

	return nil
}

func azureLoadBalancerEndpoint(ctx context.Context, ipClient *armnetwork.PublicIPAddressesClient, cfg *azureLBConfig) (string, error) {
	publicIP, err := ipClient.Get(ctx, cfg.ResourceGroup, cfg.PublicIPName, nil)
	if err != nil {
		return "", fail.Cloud(err, "azure", "getting public IP %q", cfg.PublicIPName)
	}

	if publicIP.Properties == nil || publicIP.Properties.IPAddress == nil || *publicIP.Properties.IPAddress == "" {
		return "", fail.Cloud(errors.New("public IP address is empty"), "azure", "resolving load balancer endpoint")
	}

	return *publicIP.Properties.IPAddress, nil
}

func addNICToBackendPool(
	ctx context.Context,
	ifClient *armnetwork.InterfacesClient,
	resourceGroup, nicName, backendPoolID string,
	logger logrus.FieldLogger,
) error {
	resp, err := ifClient.Get(ctx, resourceGroup, nicName, nil)
	if err != nil {
		return fail.Cloud(err, "azure", "getting network interface %q", nicName)
	}

	nic := resp.Interface
	if nic.Properties == nil || nic.Properties.IPConfigurations == nil {
		return fail.Cloud(errors.New("no IP configurations found"), "azure", "getting network interface %q", nicName)
	}

	for _, ipConfig := range nic.Properties.IPConfigurations {
		if ipConfig == nil || ipConfig.Properties == nil || ipConfig.Properties.Primary == nil || !*ipConfig.Properties.Primary {
			continue
		}

		for _, pool := range ipConfig.Properties.LoadBalancerBackendAddressPools {
			if pool != nil && pool.ID != nil && *pool.ID == backendPoolID {
				logger.Debugf("network interface %q already associated with backend pool", nicName)

				return nil
			}
		}

		ipConfig.Properties.LoadBalancerBackendAddressPools = append(
			ipConfig.Properties.LoadBalancerBackendAddressPools,
			&armnetwork.BackendAddressPool{ID: to.Ptr(backendPoolID)},
		)
	}

	poller, err := ifClient.BeginCreateOrUpdate(ctx, resourceGroup, nicName, nic, nil)
	if err != nil {
		return fail.Cloud(err, "azure", "associating network interface %q with backend pool", nicName)
	}
	if _, err := poller.PollUntilDone(ctx, nil); err != nil {
		return fail.Cloud(err, "azure", "waiting for network interface %q association", nicName)
	}

	return nil
}

func azureErrorNotFound(err error) bool {
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		return respErr.StatusCode == http.StatusNotFound
	}

	return false
}

func azureLabels(clusterName string) map[string]string {
	return map[string]string{
		"kubeone_cluster_name": clusterName,
		"kubeone_role":         "control-plane",
	}
}

func generateAzureControlPlaneMachines(clusterName string, nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	var machines []clusterv1alpha1.Machine

	for _, node := range nodeSet {
		timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
		nodeLabels := map[string]string{
			"kubeone_own_since_timestamp": timestamp,
		}
		maps.Copy(nodeLabels, azureLabels(clusterName))

		if node.NodeSettings.Labels == nil {
			node.NodeSettings.Labels = map[string]string{}
		}
		maps.Copy(node.NodeSettings.Labels, nodeLabels)

		for idx := range node.Replicas {
			osSpecRaw, err := json.Marshal(node.OperatingSystemSpec)
			if err != nil {
				return nil, err
			}

			var azureConfig azuretypes.RawConfig
			if err = jsonutil.StrictUnmarshal(node.CloudProviderSpec, &azureConfig); err != nil {
				return nil, fail.Config(err, "decode azure config")
			}

			if azureConfig.Tags == nil {
				azureConfig.Tags = map[string]string{}
			}
			maps.Copy(azureConfig.Tags, nodeLabels)

			azureSpec, err := json.Marshal(azureConfig)
			if err != nil {
				return nil, fail.Config(err, "marshaling cloud provider spec")
			}

			providerConfig := providerconfig.Config{
				SSHPublicKeys: node.SSH.PublicKeys,
				CloudProvider: providerconfig.CloudProviderAzure,
				CloudProviderSpec: runtime.RawExtension{
					Raw: azureSpec,
				},
				OperatingSystem: providerconfig.OperatingSystem(node.OperatingSystem),
				OperatingSystemSpec: runtime.RawExtension{
					Raw: osSpecRaw,
				},
			}

			providerSpecRaw, err := json.Marshal(providerConfig)
			if err != nil {
				return nil, fail.Cloud(err, "azure", "json marshaling provider config")
			}

			name := node.MachineName(clusterName, idx)
			machines = append(machines, clusterv1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					UID:        types.UID(name),
					Finalizers: slices.Clone(cleanupFinalizers),
				},
				Spec: clusterv1alpha1.MachineSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:        name,
						Labels:      node.NodeSettings.Labels,
						Annotations: node.NodeSettings.Annotations,
					},
					Taints: node.NodeSettings.Taints,
					Versions: clusterv1alpha1.MachineVersionInfo{
						Kubelet: kubeletVersion,
					},
					ProviderSpec: clusterv1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{
							Raw: providerSpecRaw,
						},
					},
				},
			})
		}
	}

	return machines, nil
}
