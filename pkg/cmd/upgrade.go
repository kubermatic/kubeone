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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kubermatic/kubeone/pkg/credentials"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/tasks"
)

type upgradeOpts struct {
	globalOptions
	ForceUpgrade              bool `longflag:"force" shortflag:"f"`
	UpgradeMachineDeployments bool `longflag:"upgrade-machine-deployments"`
}

func (opts *upgradeOpts) BuildState() (*state.State, error) {
	s, err := opts.globalOptions.BuildState()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build state")
	}

	s.ForceUpgrade = opts.ForceUpgrade
	s.UpgradeMachineDeployments = opts.UpgradeMachineDeployments
	return s, nil
}

func upgradeCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &upgradeOpts{}
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

			opts.globalOptions = *gopts
			opts.ManifestFile = args[0]

			return runUpgrade(opts)
		},
	}

	cmd.Flags().BoolVarP(
		&opts.ForceUpgrade,
		longFlagName(opts, "ForceUpgrade"),
		shortFlagName(opts, "ForceUpgrade"),
		false,
		"force start upgrade process")

	cmd.Flags().BoolVar(
		&opts.UpgradeMachineDeployments,
		longFlagName(opts, "UpgradeMachineDeployments"),
		false,
		"upgrade MachineDeployments objects")

	return cmd
}

// runUpgrade upgrades Kubernetes on the provided machines
func runUpgrade(opts *upgradeOpts) error {
	s, err := opts.BuildState()
	if err != nil {
		return errors.Wrap(err, "failed to initialize State")
	}

	// Validate credentials
	_, err = credentials.ProviderCredentials(s.Cluster.CloudProvider.Name, opts.CredentialsFile)
	if err != nil {
		return errors.Wrap(err, "failed to validate credentials")
	}

	return errors.Wrap(tasks.WithUpgrade(nil).Run(s), "failed to upgrade cluster")
}
