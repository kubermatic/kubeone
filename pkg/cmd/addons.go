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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/addons"
)

func addonsCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addons",
		Short: "Manage addons",
	}

	cmd.AddCommand(
		addonsListCmd(rootFlags),
	)

	return cmd
}

type addonsListOpts struct {
	globalOptions
	OutputFormat string `longflag:"output" shortflag:"o"`
}

func addonsListCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &addonsListOpts{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List addons",
		SilenceErrors: true,
		Example:       `kubeone -m mycluster.yaml -t terraformoutput.json addons list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			s, err := gopts.BuildState()
			if err != nil {
				return err
			}

			return addons.List(s, opts.OutputFormat)
		},
	}

	cmd.Flags().StringVarP(
		&opts.OutputFormat,
		longFlagName(opts, "OutputFormat"),
		shortFlagName(opts, "OutputFormat"),
		"table",
		"output format (table|json).")

	return cmd
}
