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

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	kubeonevalidation "github.com/kubermatic/kubeone/pkg/apis/kubeone/validation"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/yamled"

	kyaml "sigs.k8s.io/yaml"
)

const (
	// defaultKubernetesVersion is default Kubernetes version for the example configuration file
	defaultKubernetesVersion = "1.16.1"
	// defaultCloudProviderName is cloud provider to build the example configuration file for
	defaultCloudProviderName = "aws"
)

type printOptions struct {
	globalOptions
	FullConfig bool

	ClusterName       string
	KubernetesVersion string

	CloudProviderName     string
	CloudProviderExternal bool
	CloudProviderCloudCfg string

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
	EnableStaticAuditLog    bool
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
For the full reference of the configuration manifest, run the print command with --full flag.
`,
		Args:    cobra.ExactArgs(0),
		Example: fmt.Sprintf("kubeone config print --provider digitalocean --kubernetes-version %s --cluster-name example", defaultKubernetesVersion),
		RunE: func(_ *cobra.Command, args []string) error {
			return runPrint(pOpts)
		},
	}

	// General
	cmd.Flags().BoolVarP(&pOpts.FullConfig, "full", "f", false, "show full manifest")

	cmd.Flags().StringVarP(&pOpts.ClusterName, "cluster-name", "n", "demo-cluster", "cluster name")
	cmd.Flags().StringVarP(&pOpts.KubernetesVersion, "kubernetes-version", "k", defaultKubernetesVersion, "Kubernetes version")
	cmd.Flags().StringVarP(&pOpts.CloudProviderName, "provider", "p", defaultCloudProviderName, "cloud provider name (aws, digitalocean, gce, hetzner, packet, openstack, vsphere, none)")

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
	cmd.Flags().BoolVarP(&pOpts.EnableStaticAuditLog, "enable-static-audit-log", "", false, "enable StaticAuditLog")
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
	if printOptions.FullConfig {
		p := kubeoneapi.CloudProviderName(printOptions.CloudProviderName)
		switch p {
		case kubeoneapi.CloudProviderNameDigitalOcean, kubeoneapi.CloudProviderNamePacket, kubeoneapi.CloudProviderNameHetzner:
			printOptions.CloudProviderExternal = true
		case kubeoneapi.CloudProviderNameOpenStack:
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

		// CloudProvider validation
		errs := kubeonevalidation.ValidateCloudProviderSpec(cfg.CloudProvider, nil)
		if len(errs) != 0 {
			return errors.Errorf("unable to validate cloud provider spec: %s", errs.ToAggregate().Error())
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

func createAndPrintManifest(printOptions *printOptions) error {
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
		cfg.Set(yamled.Path{"cloudProvider", "cloudConfig"}, "<< cloudConfig is required for OpenStack >>")
	}

	// Hosts
	if len(printOptions.Hosts) != 0 {
		if err := parseHosts(cfg, printOptions.Hosts); err != nil {
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

	cfg := &kubeoneapi.KubeOneCluster{}
	err = kyaml.UnmarshalStrict(buffer.Bytes(), &cfg)
	if err != nil {
		return errors.Wrap(err, "failed to decode new config")
	}

	// CloudProvider validation
	errs := kubeonevalidation.ValidateCloudProviderSpec(cfg.CloudProvider, nil)
	if len(errs) != 0 {
		return errors.Errorf("unable to validate cloud provider spec: %s", errs.ToAggregate().Error())
	}

	// Print new config yaml
	err = yaml.NewEncoder(os.Stdout).Encode(cfgYaml)
	if err != nil {
		return errors.Wrap(err, "failed to encode new config as YAML")
	}

	return nil
}

const exampleManifest = `
apiVersion: kubeone.io/v1alpha1
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
    # possible values:
    # * canal
    # * weave-net
    provider: canal
    # when selected CNI provider support encryption and encrypted: true is
    # set, secret will be automatically generated and referenced in appropriate
    # manifests. Currently only weave-net supports encryption.
    encrypted: false

cloudProvider:
  # Supported cloud provider names:
  # * aws
  # * digitalocean
  # * hetzner
  # * none
  # * openstack
  # * packet
  # * vsphere
  name: "{{ .CloudProviderName }}"
  # Set the kubelet flag '--cloud-provider=external' and deploy the external CCM for supported providers
  external: {{ .CloudProviderExternal }}
  # Path to file that will be uploaded and used as custom '--cloud-config' file.
  cloudConfig: "{{ .CloudProviderCloudCfg }}"

features:
  # Enables PodSecurityPolicy admission plugin in API server, as well as creates
  # default 'privileged' PodSecurityPolicy, plus RBAC rules to authorize
  # 'kube-system' namespace pods to 'use' it.
  podSecurityPolicy:
    enable: {{ .EnablePodSecurityPolicy }}
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

# The list of nodes can be overwritten by providing Terraform output.
# You are strongly encouraged to provide an odd number of nodes and
# have at least three of them.
# Remember to only specify your *master* nodes.
# hosts:
# - publicAddress: '1.2.3.4'
#   privateAddress: '172.18.0.1'
#   bastion: '4.3.2.1'
#   bastionPort: 22  # can be left out if using the default (22)
#   bastionUser: 'root'  # can be left out if using the default ('root')
#   sshPort: 22 # can be left out if using the default (22)
#   sshUsername: ubuntu
#   # You usually want to configure either a private key OR an
#   # agent socket, but never both. The socket value can be
#   # prefixed with "env:" to refer to an environment variable.
#   sshPrivateKeyFile: '/home/me/.ssh/id_rsa'
#   sshAgentSocket: 'env:SSH_AUTH_SOCK'

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
  # Defines for what provider the machine-controller will be configured (defaults to cloudProvider.Name)
  # provider: ""

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
# workers:
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
