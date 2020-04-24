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

	"github.com/kubermatic/kubeone/pkg/clusterstatus"
	"github.com/kubermatic/kubeone/pkg/kubeconfig"
)

type statusOptions struct {
	globalOptions
	Manifest string
}

// statusCmd returns the structure for declaring the "status" subcommand.
func statusCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &statusOptions{}
	cmd := &cobra.Command{
		Use:   "status <manifest>",
		Short: "Status of the cluster",
		Long: `Status of the cluster.

This command takes KubeOne manifest which contains information about hosts.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.`,
		Args:    cobra.ExactArgs(1),
		Example: `kubeone status mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			opts.globalOptions = *gopts
			opts.Manifest = args[0]

			return runStatus(opts)
		},
	}

	return cmd
}

// runStatus gets cluster status
func runStatus(opts *statusOptions) error {
	s, err := opts.BuildState()
	if err != nil {
		return errors.Wrap(err, "failed to initialize State")
	}

	if err = kubeconfig.BuildKubernetesClientset(s); err != nil {
		return err
	}

	return clusterstatus.PrintClusterStatus(s)
}
