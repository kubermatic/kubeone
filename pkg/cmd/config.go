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

	DeployMachineController bool

	EnablePodSecurityPolicy bool
	EnableDynamicAuditLog   bool
	EnableMetricsServer     bool
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
		Example: `kubeone config print`,
		RunE: func(_ *cobra.Command, args []string) error {
			return runPrint(pOpts)
		},
	}

	cmd.Flags().StringVarP(&pOpts.ClusterName, "cluster-name", "n", "", "cluster name")
	cmd.Flags().StringVarP(&pOpts.KubernetesVersion, "kubernetes-version", "k", defaultKubernetesVersion, "Kubernetes version")
	cmd.Flags().StringVarP(&pOpts.CloudProviderName, "provider", "p", defaultCloudProviderName, "cloud provider name (aws, digitalocean, gce, hetzner, packet, openstack, none)")

	cmd.Flags().BoolVarP(&pOpts.DeployMachineController, "deploy-machine-controller", "", true, "deploy kubermatic machine-controller")

	cmd.Flags().BoolVarP(&pOpts.EnablePodSecurityPolicy, "enable-pod-security-policy", "", false, "enable PodSecurityPolicy")
	cmd.Flags().BoolVarP(&pOpts.EnableDynamicAuditLog, "enable-dynamic-audit-log", "", false, "enable DynamicAuditLog")
	cmd.Flags().BoolVarP(&pOpts.EnableMetricsServer, "enable-metrics-server", "", true, "enable metrics-server")

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
	}

	// machine-controller
	if !printOptions.DeployMachineController {
		cfg.Set(yamled.Path{"machineController", "deploy"}, printOptions.DeployMachineController)
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

	// Print the manifest
	err := validateAndPrintConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to validate and print config")
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
