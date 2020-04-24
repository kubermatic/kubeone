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

	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/tasks"
)

type resetOpts struct {
	globalOptions
	DestroyWorkers bool `longflag:"destroy-workers"`
	RemoveBinaries bool `longflag:"remove-binaries"`
}

func (opts *resetOpts) BuildState() (*state.State, error) {
	s, err := opts.globalOptions.BuildState()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build State")
	}

	s.DestroyWorkers = opts.DestroyWorkers
	s.RemoveBinaries = opts.RemoveBinaries
	return s, nil
}

// resetCmd setups reset command
func resetCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &resetOpts{}

	cmd := &cobra.Command{
		Use:   "reset <manifest>",
		Short: "Revert changes",
		Long: `
Undo all changes done by KubeOne to the configured machines.

This command takes KubeOne manifest which contains information about hosts.
It's possible to source information about hosts from Terraform output, using the
'--tfjson' flag.
`,
		Args:    cobra.ExactArgs(1),
		Example: `kubeone reset mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			opts.globalOptions = *gopts
			opts.ManifestFile = args[0]

			return runReset(opts)
		},
	}

	cmd.Flags().BoolVar(
		&opts.DestroyWorkers,
		longFlagName(opts, "DestroyWorkers"),
		true,
		"destroy all worker machines before resetting the cluster")

	cmd.Flags().BoolVar(
		&opts.RemoveBinaries,
		longFlagName(opts, "RemoveBinaries"),
		false,
		"remove kubernetes binaries after resetting the cluster")

	return cmd
}

// runReset resets all machines provisioned by KubeOne
func runReset(opts *resetOpts) error {
	s, err := opts.BuildState()
	if err != nil {
		return errors.Wrap(err, "failed to initialize State")
	}

	return errors.Wrap(tasks.WithReset(nil).Run(s), "failed to reset the cluster")
}
