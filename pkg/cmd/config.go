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
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v2"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/apis/kubeone/config"
	"k8c.io/kubeone/pkg/templates/machinecontroller"
	"k8c.io/kubeone/pkg/yamled"

	kyaml "sigs.k8s.io/yaml"
)

const (
	// defaultKubernetesVersion is default Kubernetes version for the example configuration file
	defaultKubernetesVersion = "1.18.2"
	// defaultCloudProviderName is cloud provider to build the example configuration file for
	defaultCloudProviderName = "aws"
)

type printOpts struct {
	FullConfig bool `longflag:"full" shortflag:"f"`

	ClusterName       string `longflag:"cluster-name" shortflag:"n"`
	KubernetesVersion string `longflag:"kubernetes-version" shortflag:"k"`

	CloudProviderName     string `longflag:"provider" shortflag:"p"`
	CloudProviderExternal bool
	CloudProviderCloudCfg string

	ControlPlaneHosts string `longflag:"control-plane-hosts"`

	APIEndpointHost string `longflag:"api-endpoint-host"`
	APIEndpointPort int    `longflag:"api-endpoint-port"`

	PodSubnet     string `longflag:"pod-subnet"`
	ServiceSubnet string `longflag:"service-subnet"`
	ServiceDNS    string `longflag:"service-dns"`
	NodePortRange string `longflag:"node-port-range"`

	HTTPProxy  string `longflag:"proxy-http"`
	HTTPSProxy string `longflag:"proxy-https"`
	NoProxy    string `longflag:"proxy-no-proxy"`

	EnablePodNodeSelector   bool `longflag:"enable-pod-node-selector"`
	EnablePodSecurityPolicy bool `longflag:"enable-pod-security-policy"`
	EnablePodPresets        bool `longflag:"enable-pod-presets"`
	EnableStaticAuditLog    bool `longflag:"enable-static-audit-log"`
	EnableDynamicAuditLog   bool `longflag:"enable-dynamic-audit-log"`
	EnableMetricsServer     bool `longflag:"enable-metrics-server"`
	EnableOpenIDConnect     bool `longflag:"enable-openid-connect"`

	DeployMachineController bool `longflag:"deploy-machine-controller"`
}

// configCmd setups the config command
func configCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Commands for working with the KubeOneCluster configuration manifests",
	}

	cmd.AddCommand(printCmd())
	cmd.AddCommand(migrateCmd(rootFlags))
	cmd.AddCommand(machinedeploymentsCmd(rootFlags))

	return cmd
}

// printCmd setups the print command
func printCmd() *cobra.Command {
	opts := &printOpts{}
	cmd := &cobra.Command{
		Use:   "print",
		Short: "Print an example configuration manifest",
		Long: `
Print an example configuration manifest. Using the appropriate flags you can
customize the configuration manifest. For the full reference of the
configuration manifest, run the print command with --full flag.
`,
		Args:    cobra.ExactArgs(0),
		Example: fmt.Sprintf("kubeone config print --provider digitalocean --kubernetes-version %s --cluster-name example", defaultKubernetesVersion),
		RunE: func(_ *cobra.Command, args []string) error {
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
		defaultKubernetesVersion,
		"Kubernetes version")

	cmd.Flags().StringVarP(
		&opts.CloudProviderName,
		longFlagName(opts, "CloudProviderName"),
		shortFlagName(opts, "CloudProviderName"),
		defaultCloudProviderName,
		"cloud provider name (aws, digitalocean, gce, hetzner, packet, openstack, vsphere, none)")

	// Hosts
	cmd.Flags().StringVar(&opts.ControlPlaneHosts, longFlagName(opts, "ControlPlaneHosts"), "", "control plane hosts in format of comma-separated key:value list, example: publicAddress:192.168.0.100,privateAddress:192.168.1.100,sshUsername:ubuntu,sshPort:22. Use quoted string of space separated values for multiple hosts")

	// API endpoint
	cmd.Flags().StringVar(&opts.APIEndpointHost, longFlagName(opts, "APIEndpointHost"), "", "API endpoint hostname or address")
	cmd.Flags().IntVar(&opts.APIEndpointPort, longFlagName(opts, "APIEndpointPort"), 6443, "API endpoint port")

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
	cmd.Flags().BoolVar(&opts.EnablePodSecurityPolicy, longFlagName(opts, "EnablePodSecurityPolicy"), false, "enable PodSecurityPolicy")
	cmd.Flags().BoolVar(&opts.EnablePodPresets, longFlagName(opts, "EnablePodPresets"), false, "enable PodPresets")
	cmd.Flags().BoolVar(&opts.EnableStaticAuditLog, longFlagName(opts, "EnableStaticAuditLog"), false, "enable StaticAuditLog")
	cmd.Flags().BoolVar(&opts.EnableDynamicAuditLog, longFlagName(opts, "EnableDynamicAuditLog"), false, "enable DynamicAuditLog")
	cmd.Flags().BoolVar(&opts.EnableMetricsServer, longFlagName(opts, "EnableMetricsServer"), true, "enable metrics-server")
	cmd.Flags().BoolVar(&opts.EnableOpenIDConnect, longFlagName(opts, "EnableOpenIDConnect"), false, "enable OpenID Connect authentication")

	// MachineController
	cmd.Flags().BoolVar(&opts.DeployMachineController, longFlagName(opts, "DeployMachineController"), true, "deploy kubermatic machine-controller")

	return cmd
}

// migrateCmd setups the migrate command
func migrateCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the v1alpha1 KubeOneCluster manifest to the v1beta1 version",
		Long: `
Migrate the v1alpha1 KubeOneCluster manifest to the v1beta1 version.
The v1alpha1 version of the KubeOneCluster manifest is deprecated and will be
removed in one of the next versions.
The new manifest is printed on the standard output.
`,
		Args:    cobra.ExactArgs(0),
		Example: `kubeone config migrate --manifest mycluster.yaml`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			return runMigrate(gopts)
		},
	}

	return cmd
}

// machinedeploymentsCmd setups the machinedeployments command
func machinedeploymentsCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "machinedeployments",
		Short: "Print the manifest for creating MachineDeployments",
		Long: `
Print the manifest for creating MachineDeployment objects.

The manifest contains all MachineDeployments defined in the API/config.
Note that manifest may include already created MachineDeplyoments.
The manifest is printed on the standard output.
`,
		Args:    cobra.ExactArgs(0),
		Example: `kubeone config machinedeplyoments --manifest mycluster.yaml`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
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
		case "digitalocean", "packet", "hetzner":
			printOptions.CloudProviderExternal = true
		case "openstack":
			printOptions.CloudProviderCloudCfg = "<< cloudConfig is required for OpenStack >>"
		}

		tmpl, err := template.New("example-manifest").Parse(exampleManifest)
		if err != nil {
			return errors.Wrap(err, "unable to parse the example manifest template")
		}

		var buffer bytes.Buffer
		err = tmpl.Execute(&buffer, printOptions)
		if err != nil {
			return errors.Wrap(err, "unable to run the example manifest template")
		}

		cfg := &kubeoneapi.KubeOneCluster{}
		err = kyaml.UnmarshalStrict(buffer.Bytes(), &cfg)
		if err != nil {
			return errors.Wrap(err, "failed to decode new config")
		}

		fmt.Println(buffer.String())

		return nil
	}

	err := createAndPrintManifest(printOptions)
	if err != nil {
		return errors.Wrap(err, "unable to create example manifest")
	}

	return nil
}

func createAndPrintManifest(printOptions *printOpts) error {
	cfg := &yamled.Document{}

	// API data
	cfg.Set(yamled.Path{"apiVersion"}, "kubeone.io/v1beta1")
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
		cfg.Set(yamled.Path{"cloudProvider", "cloudConfig"}, "<< cloudConfig is required for OpenStack >>")
	case "packet":
		cfg.Set(yamled.Path{"cloudProvider", "packet"}, providerVal)
		cfg.Set(yamled.Path{"cloudProvider", "external"}, true)
	case "vsphere":
		cfg.Set(yamled.Path{"cloudProvider", "vsphere"}, providerVal)
	case "none":
		cfg.Set(yamled.Path{"cloudProvider", "none"}, providerVal)
	}

	// Hosts
	if len(printOptions.ControlPlaneHosts) != 0 {
		if err := parseControlPlaneHosts(cfg, printOptions.ControlPlaneHosts); err != nil {
			return errors.Wrap(err, "unable to parse provided hosts")
		}
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
	if printOptions.EnablePodNodeSelector {
		cfg.Set(yamled.Path{"features", "podNodeSelector", "enable"}, printOptions.EnablePodSecurityPolicy)
		cfg.Set(yamled.Path{"features", "podNodeSelector", "config", "configFilePath"}, "")
	}
	if printOptions.EnablePodSecurityPolicy {
		cfg.Set(yamled.Path{"features", "podSecurityPolicy", "enable"}, printOptions.EnablePodSecurityPolicy)
	}
	if printOptions.EnablePodPresets {
		cfg.Set(yamled.Path{"features", "podPresets", "enable"}, printOptions.EnablePodPresets)
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

func parseControlPlaneHosts(cfg *yamled.Document, hostList string) error {
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

		cfg.Set(yamled.Path{"controlPlane", "hosts", i}, h)
	}

	return nil
}

// runMigrate migrates the KubeOneCluster manifest from v1alpha1 to v1beta1
func runMigrate(opts *globalOptions) error {
	// Convert old config yaml to new config yaml
	newConfigYAML, err := config.MigrateOldConfig(opts.ManifestFile)
	if err != nil {
		return errors.Wrap(err, "unable to migrate the provided configuration")
	}

	err = validateAndPrintConfig(newConfigYAML)
	if err != nil {
		return errors.Wrap(err, "unable to validate and print config")
	}

	return nil
}

// runGenerateMachineDeployments generates the MachineDeployments manifest
func runGenerateMachineDeployments(opts *globalOptions) error {
	s, err := opts.BuildState()
	if err != nil {
		return errors.Wrap(err, "failed to initialize State")
	}

	manifest, err := machinecontroller.GenerateMachineDeploymentsManifest(s)
	if err != nil {
		return errors.Wrap(err, "failed to generate machinedeployments manifest")
	}

	fmt.Println(manifest)

	return nil
}

func validateAndPrintConfig(cfgYaml interface{}) error {
	// Validate new config by unmarshaling
	var buffer bytes.Buffer
	err := yaml.NewEncoder(&buffer).Encode(cfgYaml)
	if err != nil {
		return errors.Wrap(err, "failed to encode new config as YAML")
	}

	cfg := &kubeoneapi.KubeOneCluster{}
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

const exampleManifest = `
apiVersion: kubeone.io/v1beta1
kind: KubeOneCluster
name: {{ .ClusterName }}

versions:
  kubernetes: "{{ .KubernetesVersion }}"

clusterNetwork:
  # the subnet used for pods (default: 10.244.0.0/16)
  podSubnet: "{{ .PodSubnet }}"
  # the subnet used for services (default: 10.96.0.0/12)
  serviceSubnet: "{{ .ServiceSubnet }}"
  # the domain name used for services (default: cluster.local)
  serviceDomainName: "{{ .ServiceDNS }}"
  # a nodePort range to reserve for services (default: 30000-32767)
  nodePortRange: "{{ .NodePortRange }}"
  # CNI plugin of choice. CNI can not be changed later at upgrade time.
  cni:
    # Only one CNI plugin can be defined at the same time
    # Supported CNI plugins:
    # * canal
    # * weave-net
    # * external - The CNI plugin can be installed as an addon or manually
	canal:
	  # MTU represents the maximum transmission unit.
	  # Default MTU value depends on the specified provider:
	  # * AWS - 8951 (9001 AWS Jumbo Frame - 50 VXLAN bytes)
	  # * GCE - 1410 (GCE specific 1460 bytes - 50 VXLAN bytes)
	  # * Hetzner - 1400 (Hetzner specific 1450 bytes - 50 VXLAN bytes)
	  # * OpenStack - 1400 (Hetzner specific 1450 bytes - 50 VXLAN bytes)
	  # * Default - 1450
	  mtu: 1450
    # weaveNet:
    #   # When true is set, secret will be automatically generated and
    #   # referenced in appropriate manifests. Currently only weave-net
    #   # supports encryption.
    #   encrypted: true
    # external: {}

cloudProvider:
  # Only one cloud provider can be defined at the same time.
  # Possible values:
  # aws: {}
  # azure: {}
  # digitalocean: {}
  # gce: {}
  # hetzner:
  #   networkID: ""
  # openstack: {}
  # packet: {}
  # vsphere: {}
  # none: {}
  {{ .CloudProviderName }}: {}
  # Set the kubelet flag '--cloud-provider=external' and deploy the external CCM for supported providers
  external: {{ .CloudProviderExternal }}
  # Path to file that will be uploaded and used as custom '--cloud-config' file.
  cloudConfig: "{{ .CloudProviderCloudCfg }}"

features:
  # Enable the PodNodeSelector admission plugin in API server.
  # More info: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#podnodeselector
  podNodeSelector:
    enable: {{ .EnablePodNodeSelector }}
    config:
      # configFilePath is a path on a local file system to the podNodeSelector
      # plugin config, which defines default and allowed node selectors.
      # configFilePath is is a required field.
      # More info: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#configuration-file-format-1
      configFilePath: ""
  # Enables PodSecurityPolicy admission plugin in API server, as well as creates
  # default 'privileged' PodSecurityPolicy, plus RBAC rules to authorize
  # 'kube-system' namespace pods to 'use' it.
  podSecurityPolicy:
    enable: {{ .EnablePodSecurityPolicy }}
  # Enables PodPresets admission plugin in API server.
  podPresets:
    enable: {{ .EnablePodPresets }}
  # Enables and configures audit log backend.
  # More info: https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#log-backend
  staticAuditLog:
    enable: {{ .EnableStaticAuditLog }}
    config:
      # PolicyFilePath is a path on local file system to the audit policy manifest
      # which defines what events should be recorded and what data they should include.
      # PolicyFilePath is a required field.
      # More info: https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#audit-policy
      policyFilePath: ""
      # LogPath is path on control plane instances where audit log files are stored
      logPath: "/var/log/kubernetes/audit.log"
      # LogMaxAge is maximum number of days to retain old audit log files
      logMaxAge: 30
      # LogMaxBackup is maximum number of audit log files to retain
      logMaxBackup: 3
      # LogMaxSize is maximum size in megabytes of audit log file before it gets rotated
      logMaxSize: 100
  # Enables dynamic audit logs.
  # After enablig this, operator should create auditregistration.k8s.io/v1alpha1
  # AuditSink object.
  # More info: https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#dynamic-backend
  dynamicAuditLog:
    enable: {{ .EnableDynamicAuditLog }}
  # Opt-out from deploying metrics-server
  # more info: https://github.com/kubernetes-incubator/metrics-server
  metricsServer:
    # enabled by default
    enable: {{ .EnableMetricsServer }}
  # Enable OpenID-Connect support in API server
  # More info: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens
  openidConnect:
    enable: {{ .EnableOpenIDConnect }}
    config:
      # The URL of the OpenID issuer, only HTTPS scheme will be accepted. If
      # set, it will be used to verify the OIDC JSON Web Token (JWT).
      issuerUrl: ""
      # The client ID for the OpenID Connect client, must be set if
      # issuer_url is set.
      clientId: "kubernetes"
      # The OpenID claim to use as the user name. Note that claims other than
      # the default ('sub') is not guaranteed to be unique and immutable. This
      # flag is experimental in kubernetes, please see the kubernetes
      # authentication documentation for further details.
      usernameClaim: "sub"
      # If provided, all usernames will be prefixed with this value. If not
      # provided, username claims other than 'email' are prefixed by the issuer
      # URL to avoid clashes. To skip any prefixing, provide the value '-'.
      usernamePrefix: "oidc:"
      # If provided, the name of a custom OpenID Connect claim for specifying
      # user groups. The claim value is expected to be a string or array of
      # strings. This flag is experimental in kubernetes, please see the
      # kubernetes authentication documentation for further details.
      groupsClaim: "groups"
      # If provided, all groups will be prefixed with this value to prevent
      # conflicts with other authentication strategies.
      groupsPrefix: "oidc:"
      # Comma-separated list of allowed JOSE asymmetric signing algorithms. JWTs
      # with a 'alg' header value not in this list will be rejected. Values are
      # defined by RFC 7518 https://tools.ietf.org/html/rfc7518#section-3.1.
      signingAlgs: "RS256"
      # A key=value pair that describes a required claim in the ID Token. If
      # set, the claim is verified to be present in the ID Token with a matching
      # value. Only single pair is currently supported.
      requiredClaim: ""
      # If set, the OpenID server's certificate will be verified by one of the
      # authorities in the oidc-ca-file, otherwise the host's root CA set will
      # be used.
      caFile: ""

systemPackages:
  # will add Docker and Kubernetes repositories to OS package manager
  configureRepositories: true # it's true by default

# Addons are Kubernetes manifests to be deployed after provisioning the cluster
addons:
  enable: false
  # In case when the relative path is provided, the path is relative
  # to the KubeOne configuration file.
  path: "./addons"

# The list of nodes can be overwritten by providing Terraform output.
# You are strongly encouraged to provide an odd number of nodes and
# have at least three of them.
# Remember to only specify your *master* nodes.
# controlPlane:
#   hosts:
#   - publicAddress: '1.2.3.4'
#     privateAddress: '172.18.0.1'
#     bastion: '4.3.2.1'
#     bastionPort: 22  # can be left out if using the default (22)
#     bastionUser: 'root'  # can be left out if using the default ('root')
#     sshPort: 22 # can be left out if using the default (22)
#     sshUsername: root
#     # You usually want to configure either a private key OR an
#     # agent socket, but never both. The socket value can be
#     # prefixed with "env:" to refer to an environment variable.
#     sshPrivateKeyFile: '/home/me/.ssh/id_rsa'
#     sshAgentSocket: 'env:SSH_AUTH_SOCK'
#     # Taints is used to apply taints to the node.
#     # If not provided defaults to TaintEffectNoSchedule, with key
#     # node-role.kubernetes.io/master for control plane nodes.
#     # Explicitly empty (i.e. taints: {}) means no taints will be applied.
#     taints:
#     - key: "node-role.kubernetes.io/master"
#       effect: "NoSchedule"

# A list of static workers, not managed by MachineController.
# The list of nodes can be overwritten by providing Terraform output.
# staticWorkers:
#   hosts:
#   - publicAddress: '1.2.3.5'
#     privateAddress: '172.18.0.2'
#     bastion: '4.3.2.1'
#     bastionPort: 22  # can be left out if using the default (22)
#     bastionUser: 'root'  # can be left out if using the default ('root')
#     sshPort: 22 # can be left out if using the default (22)
#     sshUsername: root
#     # You usually want to configure either a private key OR an
#     # agent socket, but never both. The socket value can be
#     # prefixed with "env:" to refer to an environment variable.
#     sshPrivateKeyFile: '/home/me/.ssh/id_rsa'
#     sshAgentSocket: 'env:SSH_AUTH_SOCK'
#     # Taints is used to apply taints to the node.
#     # Explicitly empty (i.e. taints: {}) means no taints will be applied.
#     # taints:
#     # - key: ""
#     #   effect: ""

# The API server can also be overwritten by Terraform. Provide the
# external address of your load balancer or the public addresses of
# the first control plane nodes.
# apiEndpoint:
#   host: '{{ .APIEndpointHost }}'
#   port: {{ .APIEndpointPort }}

# If the cluster runs on bare metal or an unsupported cloud provider,
# you can disable the machine-controller deployment entirely. In this
# case, anything you configure in your "workers" sections is ignored.
machineController:
  deploy: {{ .DeployMachineController }}

# Proxy is used to configure HTTP_PROXY, HTTPS_PROXY and NO_PROXY
# for Docker daemon and kubelet, and to be used when provisioning cluster
# (e.g. for curl, apt-get..).
# Also worker nodes managed by machine-controller will be configred according to
# proxy settings here. The caveat is that only proxy.http and proxy.noProxy will
# be used on worker machines.
# proxy:
#  http: '{{ .HTTPProxy }}'
#  https: '{{ .HTTPSProxy }}'
#  noProxy: '{{ .NoProxy }}'

# KubeOne can automatically create MachineDeployments to create
# worker nodes in your cluster. Each element in this "workers"
# list is a single deployment and must have a unique name.
# dynamicWorkers:
# - name: fra1-a
#   replicas: 1
#   providerSpec:
#     labels:
#       mylabel: 'fra1-a'
#     # SSH keys can be inferred from Terraform if this list is empty
#     # and your tf output contains a "ssh_public_keys" field.
#     # sshPublicKeys:
#     # - 'ssh-rsa ......'
#     # cloudProviderSpec corresponds 'provider.name' config
#     cloudProviderSpec:
#       ### the following params could be inferred by kubeone from terraform
#       ### output JSON:
#       # ami: 'ami-0332a5c40cf835528',
#       # availabilityZone: 'eu-central-1a',
#       # instanceProfile: 'mycool-profile',
#       # region: 'eu-central-1',
#       # securityGroupIDs: ['sg-01f34ffd8447e70c0']
#       # subnetId: 'subnet-2bff4f43',
#       # vpcId: 'vpc-819f62e9'
#       ### end of terraform inferred kubeone params
#       instanceType: 't3.medium'
#       diskSize: 50
#       diskType: 'gp2'
#     operatingSystem: 'ubuntu'
#     operatingSystemSpec:
#       distUpgradeOnBoot: true
# - name: fra1-b
#   replicas: 1
#   providerSpec:
#     labels:
#       mylabel: 'fra1-b'
#     cloudProviderSpec:
#       instanceType: 't3.medium'
#       diskSize: 50
#       diskType: 'gp2'
#     operatingSystem: 'ubuntu'
#     operatingSystemSpec:
#       distUpgradeOnBoot: true
# - name: fra1-c
#   replicas: 1
#   providerSpec:
#     labels:
#       mylabel: 'fra1-c'
#     cloudProviderSpec:
#       instanceType: 't3.medium'
#       diskSize: 50
#       diskType: 'gp2'
#     operatingSystem: 'ubuntu'
#     operatingSystemSpec:
#       distUpgradeOnBoot: true
`
