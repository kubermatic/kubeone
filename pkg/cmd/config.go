/*
Copyright 2019 The KubeOne Authors.

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

package cmd

import (
	"bytes"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v2"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	kubeonev1alpha1 "github.com/kubermatic/kubeone/pkg/apis/kubeone/v1alpha1"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/util/yamled"

	kyaml "sigs.k8s.io/yaml"
)

const (
	// defaultKubernetesVersion is default Kubernetes version for the example configuration file
	defaultKubernetesVersion = "1.14.1"
	// defaultCloudProviderName is cloud provider to build the example configuration file for
	defaultCloudProviderName = "aws"
)

type printOptions struct {
	globalOptions
	ClusterName       string
	KubernetesVersion string
	CloudProviderName string

	Hosts string

	APIEndpointHost string
	APIEndpointPort int

	PodSubnet     string
	ServiceSubnet string
	ServiceDNS    string
	NodePortRange string

	HTTPProxy  string
	HTTPSProxy string
	NoProxy    string

	EnablePodSecurityPolicy bool
	EnableDynamicAuditLog   bool
	EnableMetricsServer     bool
	EnableOpenIDConnect     bool

	DeployMachineController bool
}

type migrateOptions struct {
	globalOptions
	Manifest string
}

// configCmd setups the config command
func configCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Commands for working with the KubeOneCluster configuration manifests",
	}

	cmd.AddCommand(printCmd(rootFlags))
	cmd.AddCommand(migrateCmd(rootFlags))

	return cmd
}

// printCmd setups the print command
func printCmd(_ *pflag.FlagSet) *cobra.Command {
	pOpts := &printOptions{}
	cmd := &cobra.Command{
		Use:   "print",
		Short: "Print an example configuration manifest",
		Long: `
Print an example configuration manifest. Using the appropriate flags you can customize the configuration manifest.
For the full reference of the configuration manifest check the config.yaml.dist manifest
https://github.com/kubermatic/kubeone/blob/master/config.yaml.dist
`,
		Args:    cobra.ExactArgs(0),
		Example: `kubeone config print --provider digitalocean --kubernetes-version 1.14.1 --cluster-name example`,
		RunE: func(_ *cobra.Command, args []string) error {
			return runPrint(pOpts)
		},
	}

	// General
	cmd.Flags().StringVarP(&pOpts.ClusterName, "cluster-name", "n", "", "cluster name")
	cmd.Flags().StringVarP(&pOpts.KubernetesVersion, "kubernetes-version", "k", defaultKubernetesVersion, "Kubernetes version")
	cmd.Flags().StringVarP(&pOpts.CloudProviderName, "provider", "p", defaultCloudProviderName, "cloud provider name (aws, digitalocean, gce, hetzner, packet, openstack, none)")

	// Hosts
	cmd.Flags().StringVarP(&pOpts.Hosts, "hosts", "", "", "hosts in format of comma-separated key:value list, example: publicAddress:192.168.0.100,privateAddress:192.168.1.100,sshUsername:ubuntu,sshPort:22. Use quoted string of space separated values for multiple hosts")

	// API endpoint
	cmd.Flags().StringVarP(&pOpts.APIEndpointHost, "api-endpoint-host", "", "", "API endpoint hostname or address")
	cmd.Flags().IntVarP(&pOpts.APIEndpointPort, "api-endpoint-port", "", 0, "API endpoint port")

	// Cluster networking
	cmd.Flags().StringVarP(&pOpts.PodSubnet, "pod-subnet", "", "", "Subnet to be used for pods networking")
	cmd.Flags().StringVarP(&pOpts.ServiceSubnet, "service-subnet", "", "", "Subnet to be used for Services")
	cmd.Flags().StringVarP(&pOpts.ServiceDNS, "service-dns", "", "", "Domain name to be used for Services")
	cmd.Flags().StringVarP(&pOpts.NodePortRange, "node-port-range", "", "", "Port range to be used for NodePort")

	// Proxy
	cmd.Flags().StringVarP(&pOpts.HTTPProxy, "proxy-http", "", "", "HTTP proxy to be used for provisioning and Docker")
	cmd.Flags().StringVarP(&pOpts.HTTPSProxy, "proxy-https", "", "", "HTTPs proxy to be used for provisioning and Docker")
	cmd.Flags().StringVarP(&pOpts.NoProxy, "proxy-no-proxy", "", "", "No Proxy to be used for provisioning and Docker")

	// Features
	cmd.Flags().BoolVarP(&pOpts.EnablePodSecurityPolicy, "enable-pod-security-policy", "", false, "enable PodSecurityPolicy")
	cmd.Flags().BoolVarP(&pOpts.EnableDynamicAuditLog, "enable-dynamic-audit-log", "", false, "enable DynamicAuditLog")
	cmd.Flags().BoolVarP(&pOpts.EnableMetricsServer, "enable-metrics-server", "", true, "enable metrics-server")
	cmd.Flags().BoolVarP(&pOpts.EnableOpenIDConnect, "enable-openid-connect", "", false, "enable OpenID Connect authentication")

	// MachineController
	cmd.Flags().BoolVarP(&pOpts.DeployMachineController, "deploy-machine-controller", "", true, "deploy kubermatic machine-controller")

	return cmd
}

// migrateCmd setups the migrate command
func migrateCmd(_ *pflag.FlagSet) *cobra.Command {
	mOpts := &migrateOptions{}
	cmd := &cobra.Command{
		Use:   "migrate <cluster-manifest>",
		Short: "Migrate the pre-v0.6.0 configuration manifest to the KubeOneCluster manifest",
		Long: `
Migrate the pre-v0.6.0 KubeOne configuration manifest to the KubeOneCluster manifest used as of v0.6.0.
The new manifest is printed on the standard output.
`,
		Args:    cobra.ExactArgs(1),
		Example: `kubeone config migrate mycluster.yaml`,
		RunE: func(_ *cobra.Command, args []string) error {
			mOpts.Manifest = args[0]
			if mOpts.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runMigrate(mOpts)
		},
	}

	return cmd
}

// runPrint prints an example configuration file
func runPrint(printOptions *printOptions) error {
	cfg := &yamled.Document{}

	// API data
	cfg.Set(yamled.Path{"apiVersion"}, "kubeone.io/v1alpha1")
	cfg.Set(yamled.Path{"kind"}, "KubeOneCluster")

	// Cluster name
	if printOptions.ClusterName != "" {
		cfg.Set(yamled.Path{"name"}, printOptions.ClusterName)
	}

	// Version
	cfg.Set(yamled.Path{"versions", "kubernetes"}, printOptions.KubernetesVersion)

	// Provider
	p := kubeoneapi.CloudProviderName(printOptions.CloudProviderName)
	cfg.Set(yamled.Path{"cloudProvider", "name"}, p)
	switch p {
	case kubeoneapi.CloudProviderNameDigitalOcean, kubeoneapi.CloudProviderNamePacket, kubeoneapi.CloudProviderNameHetzner:
		cfg.Set(yamled.Path{"cloudProvider", "external"}, true)
	case kubeoneapi.CloudProviderNameOpenStack:
		cfg.Set(yamled.Path{"cloudProvider", "cloudConfig"}, "")
	}

	// Hosts
	if len(printOptions.Hosts) != 0 {
		parseHosts(cfg, printOptions.Hosts)
	}

	// API endpoint
	if len(printOptions.APIEndpointHost) != 0 {
		cfg.Set(yamled.Path{"apiEndpoint", "host"}, printOptions.APIEndpointHost)
	}
	if printOptions.APIEndpointPort != 0 {
		cfg.Set(yamled.Path{"apiEndpoint", "port"}, printOptions.APIEndpointPort)
	}

	// Cluster networking
	if len(printOptions.PodSubnet) != 0 {
		cfg.Set(yamled.Path{"clusterNetwork", "podSubnet"}, printOptions.PodSubnet)
	}
	if len(printOptions.ServiceSubnet) != 0 {
		cfg.Set(yamled.Path{"clusterNetwork", "serviceSubnet"}, printOptions.ServiceSubnet)
	}
	if len(printOptions.ServiceDNS) != 0 {
		cfg.Set(yamled.Path{"clusterNetwork", "serviceDomainName"}, printOptions.ServiceDNS)
	}
	if len(printOptions.NodePortRange) != 0 {
		cfg.Set(yamled.Path{"clusterNetwork", "nodePortRange"}, printOptions.NodePortRange)
	}

	// Proxy
	if len(printOptions.HTTPProxy) != 0 {
		cfg.Set(yamled.Path{"proxy", "http"}, printOptions.HTTPProxy)
	}
	if len(printOptions.HTTPSProxy) != 0 {
		cfg.Set(yamled.Path{"proxy", "https"}, printOptions.HTTPSProxy)
	}
	if len(printOptions.NoProxy) != 0 {
		cfg.Set(yamled.Path{"proxy", "noProxy"}, printOptions.NoProxy)
	}

	// Features
	if printOptions.EnablePodSecurityPolicy {
		cfg.Set(yamled.Path{"features", "podSecurityPolicy", "enable"}, printOptions.EnablePodSecurityPolicy)
	}
	if printOptions.EnableDynamicAuditLog {
		cfg.Set(yamled.Path{"features", "dynamicAuditLog", "enable"}, printOptions.EnableDynamicAuditLog)
	}
	if !printOptions.EnableMetricsServer {
		cfg.Set(yamled.Path{"features", "metricsServer", "enable"}, printOptions.EnableMetricsServer)
	}
	if printOptions.EnableOpenIDConnect {
		cfg.Set(yamled.Path{"features", "openidConnect", "enable"}, printOptions.EnableOpenIDConnect)

		cfg.Set(yamled.Path{"features", "openidConnect", "config", "issuerUrl"}, "")
		cfg.Set(yamled.Path{"features", "openidConnect", "config", "clientId"}, "")
		cfg.Set(yamled.Path{"features", "openidConnect", "config", "usernameClaim"}, "")
		cfg.Set(yamled.Path{"features", "openidConnect", "config", "usernamePrefix"}, "")
		cfg.Set(yamled.Path{"features", "openidConnect", "config", "groupsClaim"}, "")
		cfg.Set(yamled.Path{"features", "openidConnect", "config", "groupsPrefix"}, "")
		cfg.Set(yamled.Path{"features", "openidConnect", "config", "signingAlgs"}, "")
		cfg.Set(yamled.Path{"features", "openidConnect", "config", "requiredClaim"}, "")
		cfg.Set(yamled.Path{"features", "openidConnect", "config", "caFile"}, "")
	}

	// machine-controller
	if !printOptions.DeployMachineController {
		cfg.Set(yamled.Path{"machineController", "deploy"}, printOptions.DeployMachineController)
	}

	// Print the manifest
	err := validateAndPrintConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to validate and print config")
	}

	return nil
}

func parseHosts(cfg *yamled.Document, hostList string) error {
	hosts := strings.Split(hostList, " ")
	for i, host := range hosts {
		fields := strings.Split(host, ",")
		h := make(map[string]interface{})

		for _, field := range fields {
			val := strings.Split(field, ":")
			if len(val) != 2 {
				return errors.New("incorrect format of host variable")
			}

			if val[0] == "sshPort" {
				portInt, err := strconv.Atoi(val[1])
				if err != nil {
					return errors.Wrap(err, "unable to convert ssh port to integer")
				}
				h[val[0]] = portInt
				continue
			}
			h[val[0]] = val[1]
		}

		cfg.Set(yamled.Path{"hosts", i}, h)
	}

	return nil
}

// runMigrate migrates the pre-v0.6.0 KubeOne API manifest to the KubeOneCluster manifest used as of v0.6.0
func runMigrate(migrateOptions *migrateOptions) error {
	// Convert old config yaml to new config yaml
	newConfigYAML, err := config.MigrateToKubeOneClusterAPI(migrateOptions.Manifest)
	if err != nil {
		return errors.Wrap(err, "unable to migrate the provided configuration")
	}

	err = validateAndPrintConfig(newConfigYAML)
	if err != nil {
		return errors.Wrap(err, "unable to validate and print config")
	}

	return nil
}

func validateAndPrintConfig(cfgYaml interface{}) error {
	// Validate new config by unmarshaling
	var buffer bytes.Buffer
	err := yaml.NewEncoder(&buffer).Encode(cfgYaml)
	if err != nil {
		return errors.Wrap(err, "failed to encode new config as YAML")
	}

	cfg := &kubeonev1alpha1.KubeOneCluster{}
	err = kyaml.UnmarshalStrict(buffer.Bytes(), &cfg)
	if err != nil {
		return errors.Wrap(err, "failed to decode new config")
	}

	// Print new config yaml
	err = yaml.NewEncoder(os.Stdout).Encode(cfgYaml)
	if err != nil {
		return errors.Wrap(err, "failed to encode new config as YAML")
	}

	return nil
}
