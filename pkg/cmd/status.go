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

	"k8c.io/kubeone/pkg/tasks"
)

// statusCmd returns the structure for declaring the "status" subcommand.
func statusCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of the cluster",
		Long: heredoc.Doc(`
			Status of the cluster.

			This command takes KubeOne manifest which contains information about hosts. It's possible to source information about
			hosts from Terraform output, using the '--tfjson' flag.
		`),
		Example:       `kubeone status -m mycluster.yaml -t terraformoutput.json`,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			return runStatus(gopts)
		},
	}

	return cmd
}

// runStatus gets cluster status
func runStatus(opts *globalOptions) error {
	s, err := opts.BuildState()
	if err != nil {
		return err
	}

	return tasks.WithClusterStatus(nil).Run(s)
}
