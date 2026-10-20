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

	"k8c.io/kubeone/pkg/addons"
	"k8c.io/kubeone/pkg/tasks"
)

func addonsCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addons",
		Short: "Manage addons",
	}

	cmd.AddCommand(
		addonsListCmd(rootFlags),
		addonsApplyCmd(rootFlags),
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
		RunE: func(_ *cobra.Command, _ []string) error {
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
		"output format (table|json).",
	)

	return cmd
}

type addonsApplyOpts struct {
	globalOptions
	Names []string `longflag:"name"`
}

func addonsApplyCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &addonsApplyOpts{}

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply user addons",
		Long: heredoc.Doc(`
			Apply the user addons and helm releases defined in the KubeOne manifest to an existing cluster.

			New addons and helm releases are installed, existing ones are updated, and
			addons marked with 'delete: true' as well as helm releases that are not
			declared anymore are removed. The rest of the cluster is not touched.
			Embedded addons (e.g. CNI, CCM, CSI) are managed by 'kubeone apply'
			and are not affected by this command.

			Use '--name' to apply only the given addons or helm releases (by release name).
		`),
		SilenceErrors: true,
		Example: heredoc.Doc(`
			kubeone addons apply -m mycluster.yaml -t terraformoutput.json
			kubeone addons apply -m mycluster.yaml -t terraformoutput.json --name my-addon`),
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			s, err := gopts.BuildState()
			if err != nil {
				return err
			}

			if err = tasks.WithFindControlPlane(nil).Run(s); err != nil {
				return err
			}

			return tasks.WithUserAddons(nil, opts.Names...).Run(s)
		},
	}

	cmd.Flags().StringSliceVar(
		&opts.Names,
		longFlagName(opts, "Names"),
		nil,
		"apply only the user addons and helm releases with the given names (can be repeated).",
	)

	return cmd
}
