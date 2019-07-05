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

	"github.com/kubermatic/kubeone/pkg/installer"
)

type resetOptions struct {
	globalOptions
	Manifest       string
	DestroyWorkers bool
	RemoveBinaries bool
}

// resetCmd setups reset command
func resetCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	ropts := &resetOptions{}
	cmd := &cobra.Command{
		Use:   "reset <manifest>",
		Short: "Revert changes",
		Long: `Undo all changes done by KubeOne to the configured machines.

This command takes KubeOne manifest which contains information about hosts.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.`,
		Args:    cobra.ExactArgs(1),
		Example: `kubeone reset mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			logger := initLogger(gopts.Verbose)
			ropts.TerraformState = gopts.TerraformState
			ropts.Verbose = gopts.Verbose

			ropts.Manifest = args[0]
			if ropts.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runReset(logger, ropts)
		},
	}

	cmd.Flags().BoolVarP(&ropts.DestroyWorkers, "destroy-workers", "", true, "destroy all worker machines before resetting the cluster")
	cmd.Flags().BoolVarP(&ropts.RemoveBinaries, "remove-binaries", "", false, "remove kubernetes binaries after resetting the cluster")

	return cmd
}

// runReset resets all machines provisioned by KubeOne
func runReset(logger *logrus.Logger, resetOptions *resetOptions) error {
	if resetOptions.Manifest == "" {
		return errors.New("no cluster config file given")
	}

	cluster, err := loadClusterConfig(resetOptions.Manifest, resetOptions.TerraformState, logger, "", "")
	if err != nil {
		return errors.Wrap(err, "failed to load cluster")
	}

	options := &installer.Options{
		Verbose:        resetOptions.Verbose,
		DestroyWorkers: resetOptions.DestroyWorkers,
		RemoveBinaries: resetOptions.RemoveBinaries,
	}

	return installer.NewInstaller(cluster, logger).Reset(options)
}
