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

type installOpts struct {
	globalOptions
	BackupFile               string `longflag:"backup" shortflag:"b"`
	NoInit                   bool   `longflag:"no-init"`
	Force                    bool   `longflag:"force"`
	CreateMachineDeployments bool   `longflag:"create-machine-deployments"`
}

func (opts *installOpts) BuildState() (*state.State, error) {
	s, err := opts.globalOptions.BuildState()
	if err != nil {
		return nil, err
	}

	s.ForceInstall = opts.Force
	s.BackupFile = defaultBackupPath(opts.BackupFile, opts.ManifestFile, s.Cluster.Name)
	s.CreateMachineDeployments = opts.CreateMachineDeployments

	return s, initBackup(s.BackupFile)
}

// installCmd setups install command
func installCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &installOpts{}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "[DEPRECATED] Install Kubernetes",
		Long: heredoc.Doc(`
			[DEPRECATED] This command is deprecated, please use kubeone apply instead.

			Install Kubernetes on pre-existing machines

			This command takes KubeOne manifest which contains information about hosts and how the cluster should be provisioned.
			It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.
		`),
		Example:       `kubeone install -m mycluster.yaml -t terraformoutput.json`,
		Hidden:        true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			opts.globalOptions = *gopts

			return runInstall(opts)
		},
	}

	cmd.Flags().StringVarP(
		&opts.BackupFile,
		longFlagName(opts, "BackupFile"),
		shortFlagName(opts, "BackupFile"),
		"",
		"path to where the PKI backup .tar.gz file should be placed (default: location of cluster config file)")

	cmd.Flags().BoolVar(
		&opts.NoInit,
		longFlagName(opts, "NoInit"),
		false,
		"don't initialize the cluster (only install binaries)")

	cmd.Flags().BoolVar(
		&opts.CreateMachineDeployments,
		longFlagName(opts, "CreateMachineDeployments"),
		true,
		"create MachineDeployments objects")

	cmd.Flags().BoolVar(
		&opts.Force,
		longFlagName(opts, "Force"),
		false,
		"use force to install new binary versions (!dangerous!)")

	return cmd
}

// runInstall provisions Kubernetes on the provided machines
func runInstall(opts *installOpts) error {
	s, err := opts.BuildState()
	if err != nil {
		return err
	}

	s.Logger.Warn("The \"kubeone install\" command is deprecated and will be removed in KubeOne 1.6. Please use \"kubeone apply\" instead.")

	// Validate credentials
	if err = validateCredentials(s, opts.CredentialsFile); err != nil {
		return err
	}

	if opts.NoInit {
		return tasks.WithBinariesOnly(nil).Run(s)
	}

	// Probe the cluster for the actual state and the needed tasks.
	probbing := tasks.WithHostnameOS(nil)
	probbing = tasks.WithProbes(probbing)

	if err = probbing.Run(s); err != nil {
		return err
	}

	return tasks.WithFullInstall(nil).Run(s)
}
