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
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network" //nolint:staticcheck
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
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

var (
	_ cloudprovider.ControlPlaneCloudProvider = &Provider{}
	_ cloudprovider.LoadBalancerProvider      = &Provider{}
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

func (p *Provider) EnsureVM(s *state.State, capimachine clusterv1alpha1.Machine) error {
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

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, cloudprovider.HostConfigsFromMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

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
	SkuName        network.LoadBalancerSkuName
	PublicIPName   string
	lbID           string
}

func (c *azureLBConfig) backendPoolID() string {
	return c.lbID + "/backendAddressPools/" + backendPoolName
}

func (c *azureLBConfig) publicIPSkuName() network.PublicIPAddressSkuName {
	if c.SkuName == network.LoadBalancerSkuNameBasic {
		return network.PublicIPAddressSkuNameBasic
	}

	return network.PublicIPAddressSkuNameStandard
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

	skuName := network.LoadBalancerSkuNameStandard
	if lb.Sku == string(network.LoadBalancerSkuNameBasic) {
		skuName = network.LoadBalancerSkuNameBasic
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

func azureNetworkClients(s *state.State) (*network.LoadBalancersClient, *network.PublicIPAddressesClient, *network.InterfacesClient, error) {
	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return nil, nil, nil, err
	}

	clientID := providerCreds[credentials.AzureClientIDMC]
	clientSecret := providerCreds[credentials.AzureClientSecretMC]
	tenantID := providerCreds[credentials.AzureTenantIDMC]
	subscriptionID := providerCreds[credentials.AzureSubscriptionIDMC]

	authorizer, err := auth.NewClientCredentialsConfig(clientID, clientSecret, tenantID).Authorizer()
	if err != nil {
		return nil, nil, nil, fail.Runtime(err, "creating azure authorizer")
	}

	lbClient := network.NewLoadBalancersClient(subscriptionID)
	lbClient.Authorizer = authorizer

	ipClient := network.NewPublicIPAddressesClient(subscriptionID)
	ipClient.Authorizer = authorizer

	ifClient := network.NewInterfacesClient(subscriptionID)
	ifClient.Authorizer = authorizer

	return &lbClient, &ipClient, &ifClient, nil
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

func ensureAzurePublicIP(ctx context.Context, ipClient *network.PublicIPAddressesClient, cfg *azureLBConfig) (*network.PublicIPAddress, error) {
	existing, err := ipClient.Get(ctx, cfg.ResourceGroup, cfg.PublicIPName, "")
	if err == nil {
		return &existing, nil
	}
	if !azureErrorNotFound(err) {
		return nil, fail.Cloud(err, "azure", "getting public IP %q", cfg.PublicIPName)
	}

	params := network.PublicIPAddress{
		Name:     to.StringPtr(cfg.PublicIPName),
		Location: to.StringPtr(cfg.Location),
		Sku: &network.PublicIPAddressSku{
			Name: cfg.publicIPSkuName(),
		},
		PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: network.Static,
			PublicIPAddressVersion:   network.IPv4,
		},
	}

	future, err := ipClient.CreateOrUpdate(ctx, cfg.ResourceGroup, cfg.PublicIPName, params)
	if err != nil {
		return nil, fail.Cloud(err, "azure", "creating public IP %q", cfg.PublicIPName)
	}
	if err = future.WaitForCompletionRef(ctx, ipClient.Client); err != nil {
		return nil, fail.Cloud(err, "azure", "waiting for public IP %q", cfg.PublicIPName)
	}

	result, err := future.Result(*ipClient)
	if err != nil {
		return nil, fail.Cloud(err, "azure", "creating public IP %q", cfg.PublicIPName)
	}

	return &result, nil
}

func ensureAzureLoadBalancer(ctx context.Context, lbClient *network.LoadBalancersClient, cfg *azureLBConfig, publicIP *network.PublicIPAddress) error {
	_, err := lbClient.Get(ctx, cfg.ResourceGroup, cfg.Name, "")
	if err == nil {
		return nil
	}
	if !azureErrorNotFound(err) {
		return fail.Cloud(err, "azure", "getting load balancer %q", cfg.Name)
	}

	frontendID := cfg.lbID + "/frontendIPConfigurations/" + frontendIPConfigName

	params := network.LoadBalancer{
		Name:     to.StringPtr(cfg.Name),
		Location: to.StringPtr(cfg.Location),
		Sku: &network.LoadBalancerSku{
			Name: cfg.SkuName,
		},
		LoadBalancerPropertiesFormat: &network.LoadBalancerPropertiesFormat{
			FrontendIPConfigurations: &[]network.FrontendIPConfiguration{
				{
					Name: to.StringPtr(frontendIPConfigName),
					FrontendIPConfigurationPropertiesFormat: &network.FrontendIPConfigurationPropertiesFormat{
						PublicIPAddress: &network.PublicIPAddress{
							ID: publicIP.ID,
						},
					},
				},
			},
			BackendAddressPools: &[]network.BackendAddressPool{
				{
					Name: to.StringPtr(backendPoolName),
				},
			},
			Probes: &[]network.Probe{
				{
					Name: to.StringPtr(probeName),
					ProbePropertiesFormat: &network.ProbePropertiesFormat{
						Protocol:          network.ProbeProtocolTCP,
						Port:              to.Int32Ptr(apiserverPort),
						IntervalInSeconds: to.Int32Ptr(5),
						NumberOfProbes:    to.Int32Ptr(2),
					},
				},
			},
			LoadBalancingRules: &[]network.LoadBalancingRule{
				{
					Name: to.StringPtr(ruleName),
					LoadBalancingRulePropertiesFormat: &network.LoadBalancingRulePropertiesFormat{
						FrontendIPConfiguration: &network.SubResource{ID: to.StringPtr(frontendID)},
						BackendAddressPool:      &network.SubResource{ID: to.StringPtr(cfg.backendPoolID())},
						Probe:                   &network.SubResource{ID: to.StringPtr(cfg.lbID + "/probes/" + probeName)},
						Protocol:                network.TransportProtocolTCP,
						FrontendPort:            to.Int32Ptr(apiserverPort),
						BackendPort:             to.Int32Ptr(apiserverPort),
					},
				},
			},
		},
	}

	future, err := lbClient.CreateOrUpdate(ctx, cfg.ResourceGroup, cfg.Name, params)
	if err != nil {
		return fail.Cloud(err, "azure", "creating load balancer %q", cfg.Name)
	}
	if err := future.WaitForCompletionRef(ctx, lbClient.Client); err != nil {
		return fail.Cloud(err, "azure", "waiting for load balancer %q", cfg.Name)
	}
	if _, err := future.Result(*lbClient); err != nil {
		return fail.Cloud(err, "azure", "creating load balancer %q", cfg.Name)
	}

	return nil
}

func azureLoadBalancerEndpoint(ctx context.Context, ipClient *network.PublicIPAddressesClient, cfg *azureLBConfig) (string, error) {
	publicIP, err := ipClient.Get(ctx, cfg.ResourceGroup, cfg.PublicIPName, "")
	if err != nil {
		return "", fail.Cloud(err, "azure", "getting public IP %q", cfg.PublicIPName)
	}

	if publicIP.IPAddress == nil || *publicIP.IPAddress == "" {
		return "", fail.Cloud(errors.New("public IP address is empty"), "azure", "resolving load balancer endpoint")
	}

	return *publicIP.IPAddress, nil
}

func addNICToBackendPool(
	ctx context.Context,
	ifClient *network.InterfacesClient,
	resourceGroup, nicName, backendPoolID string,
	logger logrus.FieldLogger,
) error {
	nic, err := ifClient.Get(ctx, resourceGroup, nicName, "")
	if err != nil {
		return fail.Cloud(err, "azure", "getting network interface %q", nicName)
	}

	if nic.IPConfigurations == nil {
		return fail.Cloud(errors.New("no IP configurations found"), "azure", "getting network interface %q", nicName)
	}

	for i := range *nic.IPConfigurations {
		ipConfig := &(*nic.IPConfigurations)[i]
		if ipConfig.Primary == nil || !*ipConfig.Primary {
			continue
		}

		if ipConfig.LoadBalancerBackendAddressPools != nil {
			for _, pool := range *ipConfig.LoadBalancerBackendAddressPools {
				if pool.ID != nil && *pool.ID == backendPoolID {
					logger.Debugf("network interface %q already associated with backend pool", nicName)

					return nil
				}
			}
		}

		pool := network.BackendAddressPool{ID: to.StringPtr(backendPoolID)}
		if ipConfig.LoadBalancerBackendAddressPools == nil {
			ipConfig.LoadBalancerBackendAddressPools = &[]network.BackendAddressPool{pool}
		} else {
			*ipConfig.LoadBalancerBackendAddressPools = append(*ipConfig.LoadBalancerBackendAddressPools, pool)
		}
	}

	future, err := ifClient.CreateOrUpdate(ctx, resourceGroup, nicName, nic)
	if err != nil {
		return fail.Cloud(err, "azure", "associating network interface %q with backend pool", nicName)
	}
	if err := future.WaitForCompletionRef(ctx, ifClient.Client); err != nil {
		return fail.Cloud(err, "azure", "waiting for network interface %q association", nicName)
	}
	if _, err := future.Result(*ifClient); err != nil {
		return fail.Cloud(err, "azure", "associating network interface %q with backend pool", nicName)
	}

	return nil
}

func azureErrorNotFound(err error) bool {
	var detailed autorest.DetailedError
	if errors.As(err, &detailed) {
		return detailed.Response != nil && detailed.Response.StatusCode == http.StatusNotFound
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

			name := fmt.Sprintf("%s-%s-%d", clusterName, node.Name, idx)
			machines = append(machines, clusterv1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					UID:  types.UID(name),
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
