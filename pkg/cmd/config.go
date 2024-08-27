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
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v2"

	"k8c.io/kubeone/pkg/apis/kubeone/config"
	kubeonev1beta3 "k8c.io/kubeone/pkg/apis/kubeone/v1beta3"
	"k8c.io/kubeone/pkg/containerruntime"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/templates/machinecontroller"
	"k8c.io/kubeone/pkg/yamled"

	kyaml "sigs.k8s.io/yaml"
)

const (
	// defaultCloudProviderName is cloud provider to build the example configuration file for
	defaultCloudProviderName = "aws"
)

//go:embed example-manifest.tmpl.yaml
var exampleManifest string

type printOpts struct {
	FullConfig bool `longflag:"full" shortflag:"f"`

	ClusterName       string `longflag:"cluster-name" shortflag:"n"`
	KubernetesVersion string `longflag:"kubernetes-version" shortflag:"k"`

	CloudProviderName     string `longflag:"provider" shortflag:"p"`
	CloudProviderExternal bool
	CloudProviderCloudCfg string

	ControlPlaneHosts string `longflag:"control-plane-hosts"`

	APIEndpointHost             string   `longflag:"api-endpoint-host"`
	APIEndpointPort             int      `longflag:"api-endpoint-port"`
	APIEndpointAlternativeNames []string `longflag:"api-endpoint-alternative-names"`

	PodSubnet     string `longflag:"pod-subnet"`
	ServiceSubnet string `longflag:"service-subnet"`
	ServiceDNS    string `longflag:"service-dns"`
	NodePortRange string `longflag:"node-port-range"`

	HTTPProxy  string `longflag:"proxy-http"`
	HTTPSProxy string `longflag:"proxy-https"`
	NoProxy    string `longflag:"proxy-no-proxy"`

	EnablePodNodeSelector     bool `longflag:"enable-pod-node-selector"`
	EnablePodSecurityPolicy   bool `longflag:"enable-pod-security-policy"` // TODO: remove in future release
	EnableStaticAuditLog      bool `longflag:"enable-static-audit-log"`
	EnableDynamicAuditLog     bool `longflag:"enable-dynamic-audit-log"`
	EnableMetricsServer       bool `longflag:"enable-metrics-server"`
	EnableOpenIDConnect       bool `longflag:"enable-openid-connect"`
	EnableEncryptionProviders bool `longflag:"enable-encryption-providers"`

	DeployMachineController bool `longflag:"deploy-machine-controller"`

	ContainerLogMaxSize  string `longflag:"container-log-max-size"`
	ContainerLogMaxFiles int32  `longflag:"container-log-max-files"`
}

// configCmd setups the config command
func configCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Commands for working with the KubeOneCluster configuration manifests",
	}

	cmd.AddCommand(configPrintCmd())
	cmd.AddCommand(configDumpCmd(rootFlags))
	cmd.AddCommand(configMigrateCmd(rootFlags))
	cmd.AddCommand(configMachinedeploymentsCmd(rootFlags))
	cmd.AddCommand(configImagesCmd(rootFlags))

	return cmd
}

// configPrintCmd setups the print command
func configPrintCmd() *cobra.Command {
	opts := &printOpts{}
	cmd := &cobra.Command{
		Use:   "print",
		Short: "Print an example configuration manifest",
		Long: heredoc.Doc(`
			Print an example configuration manifest. Using the appropriate flags you can
			customize the configuration manifest. For the full reference of the
			configuration manifest, run the print command with --full flag.
		`),
		Args:          cobra.ExactArgs(0),
		Example:       fmt.Sprintf("kubeone config print --provider digitalocean --kubernetes-version %s --cluster-name example", defaultKubeVersion),
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runPrint(opts)
		},
	}

	// General
	cmd.Flags().BoolVarP(
		&opts.FullConfig,
		longFlagName(opts, "FullConfig"),
		shortFlagName(opts, "FullConfig"),
		false,
		"show full manifest")

	cmd.Flags().StringVarP(
		&opts.ClusterName,
		longFlagName(opts, "ClusterName"),
		shortFlagName(opts, "ClusterName"),
		"demo-cluster",
		"cluster name")

	cmd.Flags().StringVarP(
		&opts.KubernetesVersion,
		longFlagName(opts, "KubernetesVersion"),
		shortFlagName(opts, "KubernetesVersion"),
		defaultKubeVersion,
		"Kubernetes version")

	cmd.Flags().StringVarP(
		&opts.CloudProviderName,
		longFlagName(opts, "CloudProviderName"),
		shortFlagName(opts, "CloudProviderName"),
		defaultCloudProviderName,
		"cloud provider name (aws, digitalocean, gce, hetzner, equinixmetal, openstack, vsphere, none)")

	// Hosts
	cmd.Flags().StringVar(&opts.ControlPlaneHosts, longFlagName(opts, "ControlPlaneHosts"), "", "control plane hosts in format of comma-separated key:value list, example: publicAddress:192.168.0.100,privateAddress:192.168.1.100,sshUsername:ubuntu,sshPort:22. Use quoted string of space separated values for multiple hosts")

	// API endpoint
	cmd.Flags().StringVar(&opts.APIEndpointHost, longFlagName(opts, "APIEndpointHost"), "", "API endpoint hostname or address")
	cmd.Flags().IntVar(&opts.APIEndpointPort, longFlagName(opts, "APIEndpointPort"), 6443, "API endpoint port")
	cmd.Flags().StringSliceVar(&opts.APIEndpointAlternativeNames, longFlagName(opts, "APIEndpointAlternativeNames"), []string{}, "Comma separated list of API endpoint alternative names, example: host.com,192.16.0.100")

	// Cluster networking
	cmd.Flags().StringVar(&opts.PodSubnet, longFlagName(opts, "PodSubnet"), "", "Subnet to be used for pods networking")
	cmd.Flags().StringVar(&opts.ServiceSubnet, longFlagName(opts, "ServiceSubnet"), "", "Subnet to be used for Services")
	cmd.Flags().StringVar(&opts.ServiceDNS, longFlagName(opts, "ServiceDNS"), "", "Domain name to be used for Services")
	cmd.Flags().StringVar(&opts.NodePortRange, longFlagName(opts, "NodePortRange"), "", "Port range to be used for NodePort")

	// Proxy
	cmd.Flags().StringVar(&opts.HTTPProxy, longFlagName(opts, "HTTPProxy"), "", "HTTP proxy to be used for provisioning and Docker")
	cmd.Flags().StringVar(&opts.HTTPSProxy, longFlagName(opts, "HTTPSProxy"), "", "HTTPs proxy to be used for provisioning and Docker")
	cmd.Flags().StringVar(&opts.NoProxy, longFlagName(opts, "NoProxy"), "", "No Proxy to be used for provisioning and Docker")

	// Features
	cmd.Flags().BoolVar(&opts.EnablePodNodeSelector, longFlagName(opts, "EnablePodNodeSelector"), false, "enable PodNodeSelector admission plugin")
	cmd.Flags().BoolVar(&opts.EnablePodSecurityPolicy, longFlagName(opts, "EnablePodSecurityPolicy"), false, "enable PodSecurityPolicy. NO-OP: this feature is removed")
	cmd.Flags().BoolVar(&opts.EnableStaticAuditLog, longFlagName(opts, "EnableStaticAuditLog"), false, "enable StaticAuditLog")
	cmd.Flags().BoolVar(&opts.EnableDynamicAuditLog, longFlagName(opts, "EnableDynamicAuditLog"), false, "enable DynamicAuditLog")
	cmd.Flags().BoolVar(&opts.EnableMetricsServer, longFlagName(opts, "EnableMetricsServer"), true, "enable metrics-server")
	cmd.Flags().BoolVar(&opts.EnableOpenIDConnect, longFlagName(opts, "EnableOpenIDConnect"), false, "enable OpenID Connect authentication")
	cmd.Flags().BoolVar(&opts.EnableEncryptionProviders, longFlagName(opts, "EnableEncryptionProviders"), false, "enable Encryption Providers")

	// MachineController
	cmd.Flags().BoolVar(&opts.DeployMachineController, longFlagName(opts, "DeployMachineController"), true, "deploy kubermatic machine-controller")

	// LoggingConfig
	cmd.Flags().StringVar(
		&opts.ContainerLogMaxSize,
		longFlagName(opts, "ContainerLogMaxSize"),
		containerruntime.DefaultContainerLogMaxSize,
		"ContainerLogMaxSize")

	cmd.Flags().Int32Var(
		&opts.ContainerLogMaxFiles,
		longFlagName(opts, "ContainerLogMaxFiles"),
		containerruntime.DefaultContainerLogMaxFiles,
		"ContainerLogMaxFiles")

	return cmd
}

// configMigrateCmd setups the migrate command
func configMigrateCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the v1beta2 KubeOneCluster manifest to the v1beta3 version",
		Long: `
Migrate the v1beta2 KubeOneCluster manifest to the v1beta3 version.
The v1beta2 version of the KubeOneCluster manifest is deprecated and will be
removed in one of the next versions.
The new manifest is printed on the standard output.
`,
		Args:          cobra.ExactArgs(0),
		Example:       `kubeone config migrate --manifest mycluster.yaml`,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			return runMigrate(gopts)
		},
	}

	return cmd
}

// configMachinedeploymentsCmd setups the machinedeployments command
func configMachinedeploymentsCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "machinedeployments",
		Short: "Print the manifest for creating MachineDeployments",
		Long: `
Print the manifest for creating MachineDeployment objects.

The manifest contains all MachineDeployments defined in the API/config.
Note that manifest may include already created MachineDeployments.
The manifest is printed on the standard output.
`,
		Args:          cobra.ExactArgs(0),
		Example:       `kubeone config machinedeployments --manifest mycluster.yaml`,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			return runGenerateMachineDeployments(gopts)
		},
	}

	return cmd
}

// runPrint prints an example configuration file
func runPrint(printOptions *printOpts) error {
	if printOptions.FullConfig {
		switch printOptions.CloudProviderName {
		case "digitalocean", "equinixmetal", "hetzner":
			printOptions.CloudProviderExternal = true
		case "openstack":
			printOptions.CloudProviderCloudCfg = "<< cloudConfig is required for OpenStack >>"
		case "vsphere":
			printOptions.CloudProviderCloudCfg = "<< cloudConfig is required for vSphere >>"
		case "azure":
			printOptions.CloudProviderCloudCfg = "<< cloudConfig is required for Azure >>"
		}

		tmpl, err := template.New("example-manifest").Parse(exampleManifest)
		if err != nil {
			return fail.Runtime(err, "parsing example-manifest template")
		}

		var buffer bytes.Buffer
		err = tmpl.Execute(&buffer, printOptions)
		if err != nil {
			return fail.Runtime(err, "executing example-manifest template")
		}

		cfg := kubeonev1beta3.NewKubeOneCluster()
		err = kyaml.UnmarshalStrict(buffer.Bytes(), &cfg)
		if err != nil {
			return fail.Runtime(err, "testing marshal/unmarshal")
		}

		fmt.Println(buffer.String())

		return nil
	}

	return createAndPrintManifest(printOptions)
}

func createAndPrintManifest(printOptions *printOpts) error {
	cfg := &yamled.Document{}

	// API data
	cfg.Set(yamled.Path{"apiVersion"}, "kubeone.k8c.io/v1beta2")
	cfg.Set(yamled.Path{"kind"}, "KubeOneCluster")

	// Cluster name
	if printOptions.ClusterName != "" {
		cfg.Set(yamled.Path{"name"}, printOptions.ClusterName)
	}

	// Version
	cfg.Set(yamled.Path{"versions", "kubernetes"}, printOptions.KubernetesVersion)

	// Provider
	var providerVal struct{}
	switch printOptions.CloudProviderName {
	case "aws":
		cfg.Set(yamled.Path{"cloudProvider", "aws"}, providerVal)
	case "azure":
		cfg.Set(yamled.Path{"cloudProvider", "azure"}, providerVal)
		cfg.Set(yamled.Path{"cloudProvider", "cloudConfig"}, "<< cloudConfig is required for Azure >>\n")
	case "digitalocean":
		cfg.Set(yamled.Path{"cloudProvider", "digitalocean"}, providerVal)
		cfg.Set(yamled.Path{"cloudProvider", "external"}, true)
	case "gce":
		cfg.Set(yamled.Path{"cloudProvider", "gce"}, providerVal)
	case "hetzner":
		cfg.Set(yamled.Path{"cloudProvider", "hetzner"}, providerVal)
		cfg.Set(yamled.Path{"cloudProvider", "external"}, true)
	case "openstack":
		cfg.Set(yamled.Path{"cloudProvider", "openstack"}, providerVal)
		cfg.Set(yamled.Path{"cloudProvider", "cloudConfig"}, "<< cloudConfig is required for OpenStack >>\n")
	case "equinixmetal":
		cfg.Set(yamled.Path{"cloudProvider", "equinixmetal"}, providerVal)
		cfg.Set(yamled.Path{"cloudProvider", "external"}, true)
	case "vmwareCloudDirector":
		cfg.Set(yamled.Path{"cloudProvider", "vmwareCloudDirector"}, providerVal)
		cfg.Set(yamled.Path{"cloudProvider", "external"}, true)
	case "vsphere":
		cfg.Set(yamled.Path{"cloudProvider", "vsphere"}, providerVal)
		cfg.Set(yamled.Path{"cloudProvider", "cloudConfig"}, "<< cloudConfig is required for vSphere >>\n")
	case "none":
		cfg.Set(yamled.Path{"cloudProvider", "none"}, providerVal)
	}

	// Hosts
	if len(printOptions.ControlPlaneHosts) != 0 {
		if err := parseControlPlaneHosts(cfg, printOptions.ControlPlaneHosts); err != nil {
			return err
		}
	}

	// API endpoint
	if len(printOptions.APIEndpointHost) != 0 {
		cfg.Set(yamled.Path{"apiEndpoint", "host"}, printOptions.APIEndpointHost)
	}
	if printOptions.APIEndpointPort != 0 {
		cfg.Set(yamled.Path{"apiEndpoint", "port"}, printOptions.APIEndpointPort)
	}

	if len(printOptions.APIEndpointAlternativeNames) > 0 {
		cfg.Set(yamled.Path{"apiEndpoint", "alternativeNames"}, printOptions.APIEndpointAlternativeNames)
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
	printFeatures(cfg, printOptions)

	// machine-controller
	if !printOptions.DeployMachineController {
		cfg.Set(yamled.Path{"machineController", "deploy"}, printOptions.DeployMachineController)
	}

	// Logging configuration
	cfg.Set(yamled.Path{"loggingConfig", "containerLogMaxSize"}, printOptions.ContainerLogMaxSize)
	cfg.Set(yamled.Path{"loggingConfig", "containerLogMaxFiles"}, printOptions.ContainerLogMaxFiles)

	// Print the manifest
	return validateAndPrintConfig(cfg)
}

func printFeatures(cfg *yamled.Document, printOptions *printOpts) {
	if printOptions.EnablePodNodeSelector {
		cfg.Set(yamled.Path{"features", "podNodeSelector", "enable"}, printOptions.EnablePodNodeSelector)
		cfg.Set(yamled.Path{"features", "podNodeSelector", "config", "configFilePath"}, "")
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
	if printOptions.EnableEncryptionProviders {
		cfg.Set(yamled.Path{"features", "encryptionProviders", "enable"}, printOptions.EnableEncryptionProviders)
	}
}

func parseControlPlaneHosts(cfg *yamled.Document, hostList string) error {
	hosts := strings.Split(hostList, " ")
	for i, host := range hosts {
		fields := strings.Split(host, ",")
		h := make(map[string]interface{})

		for _, field := range fields {
			val := strings.Split(field, ":")
			if len(val) != 2 {
				return fail.RuntimeError{
					Op:  "parsing host variable",
					Err: errors.New("incorrect format"),
				}
			}

			if val[0] == "sshPort" {
				portInt, err := strconv.Atoi(val[1])
				if err != nil {
					return fail.Runtime(err, "parsing sshPort")
				}

				h[val[0]] = portInt

				continue
			}
			h[val[0]] = val[1]
		}

		cfg.Set(yamled.Path{"controlPlane", "hosts", i}, h)
	}

	return nil
}

// runMigrate migrates the KubeOneCluster manifest from v1alpha1 to v1beta1
func runMigrate(opts *globalOptions) error {
	var (
		tfOutput []byte
		err      error
	)

	if opts.TerraformState != "" {
		tfOutput, err = config.TFOutput(opts.TerraformState)
		if err != nil {
			return err
		}
	}

	v1beta3Manifest, err := config.MigrateV1beta2V1beta3(opts.ManifestFile, tfOutput)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", v1beta3Manifest)

	return nil
}

// runGenerateMachineDeployments generates the MachineDeployments manifest
func runGenerateMachineDeployments(opts *globalOptions) error {
	s, err := opts.BuildState()
	if err != nil {
		return err
	}

	manifest, err := machinecontroller.GenerateMachineDeploymentsManifest(s)
	if err != nil {
		return err
	}

	fmt.Println(manifest)

	return nil
}

func validateAndPrintConfig(cfgYaml interface{}) error {
	// Validate new config by unmarshaling
	var buffer bytes.Buffer

	err := yaml.NewEncoder(&buffer).Encode(cfgYaml)
	if err != nil {
		return fail.Runtime(err, "marshalling new config as YAML")
	}

	cfg := kubeonev1beta3.NewKubeOneCluster()
	if err = kyaml.UnmarshalStrict(buffer.Bytes(), &cfg); err != nil {
		return fail.Runtime(err, "testing marshal/unmarshal")
	}

	// Print new config yaml
	if err = yaml.NewEncoder(os.Stdout).Encode(cfgYaml); err != nil {
		return fail.Runtime(err, "marshalling new config as YAML")
	}

	return nil
}
