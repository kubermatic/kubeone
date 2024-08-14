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
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/tasks"
)

type upgradeOpts struct {
	globalOptions
	ForceUpgrade              bool `longflag:"force" shortflag:"f"`
	UpgradeMachineDeployments bool `longflag:"upgrade-machine-deployments"`
	PruneImages               bool `longflag:"prune-images"`
}

func (opts *upgradeOpts) BuildState() (*state.State, error) {
	s, err := opts.globalOptions.BuildState()
	if err != nil {
		return nil, err
	}

	s.ForceUpgrade = opts.ForceUpgrade
	s.UpgradeMachineDeployments = opts.UpgradeMachineDeployments

	return s, nil
}

func upgradeCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &upgradeOpts{}
	cmd := &cobra.Command{
		Use:   "upgrade <manifest>",
		Short: "[DEPRECATED] Upgrade Kubernetes",
		Long: heredoc.Doc(`
			[DEPRECATED] This command is deprecated, please use kubeone apply instead.

			Upgrade Kubernetes

			This command takes KubeOne manifest which contains information about hosts and how the cluster should be provisioned.
			It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.
		`),
		Example:       `kubeone upgrade -m mycluster.yaml -t terraformoutput.json`,
		Hidden:        true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			opts.globalOptions = *gopts

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

	cmd.Flags().BoolVar(
		&opts.PruneImages,
		longFlagName(opts, "PruneImages"),
		false,
		"delete unused container images on control plane and static worker nodes")

	return cmd
}

// runUpgrade upgrades Kubernetes on the provided machines
func runUpgrade(opts *upgradeOpts) error {
	s, err := opts.BuildState()
	if err != nil {
		return err
	}

	s.Logger.Warn("The \"kubeone upgrade\" command is deprecated and will be removed in KubeOne 1.6. Please use \"kubeone apply\" instead.")

	// Validate credentials
	if err = validateCredentials(s, opts.CredentialsFile); err != nil {
		return err
	}

	// Probe the cluster for the actual state and the needed tasks.
	probbing := tasks.WithHostnameOS(nil)
	probbing = tasks.WithProbes(probbing)

	if err = probbing.Run(s); err != nil {
		return err
	}

	return tasks.WithUpgrade(nil).Run(s)
}
