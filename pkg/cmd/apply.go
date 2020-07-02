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
	"github.com/kubermatic/kubeone/pkg/tasks"
)

type applyOpts struct {
	globalOptions
	NoInit      bool `longflag:"no-init"`
	Force       bool `longflag:"force"`
	AutoApprove bool `longflag:"auto-approve"`
}

func applyCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &applyOpts{}

	cmd := &cobra.Command{
		Hidden: true, // for now
		Use:    "apply",
		Short:  "apply reconcile the cluster",
		Long: `
Reconcile (Install/Upgrade/Repair/Restore) Kubernetes cluster on pre-existing machines

This command takes KubeOne manifest which contains information about hosts and how the cluster should be provisioned.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.
`,
		Example: `kubeone apply -m mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			opts.globalOptions = *gopts

			return runApply(opts)
		},
	}

	cmd.Flags().BoolVar(
		&opts.NoInit,
		longFlagName(opts, "NoInit"),
		false,
		"don't initialize the cluster (only install binaries)")

	cmd.Flags().BoolVar(
		&opts.AutoApprove,
		longFlagName(opts, "AutoApprove"),
		false,
		"auto approve plan")

	cmd.Flags().BoolVar(
		&opts.Force,
		longFlagName(opts, "Force"),
		false,
		"use force to install new binary versions (!dangerous!)")

	return cmd
}

func runApply(opts *applyOpts) error {
	s, err := opts.BuildState()
	if err != nil {
		return errors.Wrap(err, "failed to initialize State")
	}

	// Validate credentials
	_, err = credentials.ProviderCredentials(s.Cluster.CloudProvider, opts.CredentialsFile)
	if err != nil {
		return errors.Wrap(err, "failed to validate credentials")
	}

	probbing := tasks.WithHostnameOS(nil)
	probbing = tasks.WithProbes(probbing)

	if err = probbing.Run(s); err != nil {
		return err
	}

	// later in this point we going to make decision and run different tasks, should we run install or upgrade based on
	// the state we accumulated in s.LiveCluster

	return nil
}
