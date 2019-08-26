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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kubermatic/kubeone/pkg/upgrader"
)

type upgradeOptions struct {
	globalOptions

	ForceUpgrade              bool
	Manifest                  string
	UpgradeMachineDeployments bool
}

func upgradeCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	uopts := &upgradeOptions{}
	cmd := &cobra.Command{
		Use:   "upgrade <manifest>",
		Short: "Upgrade Kubernetes",
		Long: `Upgrade Kubernetes

This command takes KubeOne manifest which contains information about hosts and how the cluster should be provisioned.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.`,
		Args:    cobra.ExactArgs(1),
		Example: `kubeone upgrade mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			logger := initLogger(gopts.Verbose)
			uopts.TerraformState = gopts.TerraformState
			uopts.Verbose = gopts.Verbose

			uopts.Manifest = args[0]
			if uopts.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runUpgrade(logger, uopts)
		},
	}

	cmd.Flags().BoolVarP(&uopts.ForceUpgrade, "force", "f", false, "force start upgrade process")
	cmd.Flags().BoolVarP(&uopts.UpgradeMachineDeployments, "upgrade-machine-deployments", "", false, "upgrade MachineDeployments objects")

	return cmd
}

// runUpgrade upgrades Kubernetes on the provided machines
func runUpgrade(logger *logrus.Logger, upgradeOptions *upgradeOptions) error {
	cluster, err := loadClusterConfig(upgradeOptions.Manifest, upgradeOptions.TerraformState, logger)
	if err != nil {
		return errors.Wrap(err, "failed to load cluster")
	}

	options := createUpgradeOptions(upgradeOptions)
	return upgrader.NewUpgrader(cluster, logger).Upgrade(options)
}

func createUpgradeOptions(options *upgradeOptions) *upgrader.Options {
	return &upgrader.Options{
		CredentialsFile:           options.CredentialsFilePath,
		ForceUpgrade:              options.ForceUpgrade,
		Verbose:                   options.Verbose,
		UpgradeMachineDeployments: options.UpgradeMachineDeployments,
	}
}
