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
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"
)

type kubeconfigOpts struct {
	globalOptions
	SuperAdmin bool `longflag:"super-admin" shortflag:"s"`
}

// KubeconfigCommand returns the structure for declaring the "install" subcommand.
func kubeconfigCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &kubeconfigOpts{}

	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Download the kubeconfig file from master",
		Long: heredoc.Doc(`
			Download the kubeconfig file from master.

			This command takes KubeOne manifest which contains information about hosts. It's possible to source information about
			hosts from Terraform output, using the '--tfjson' flag.
		`),
		Example:       `kubeone kubeconfig -m mycluster.yaml -t terraformoutput.json`,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			opts.globalOptions = *gopts
			st, err := opts.BuildState()
			if err != nil {
				return err
			}

			return runKubeconfig(st, opts.SuperAdmin)
		},
	}

	cmd.Flags().BoolVarP(
		&opts.SuperAdmin,
		longFlagName(opts, "SuperAdmin"),
		shortFlagName(opts, "SuperAdmin"),
		false,
		"generate short-lived system:masters kubeconfig")

	return cmd
}

// runKubeconfig downloads kubeconfig file
func runKubeconfig(st *state.State, generateSuperAdmin bool) error {
	var (
		konfig []byte
		err    error
	)

	if generateSuperAdmin {
		konfig, err = kubeconfig.GenerateSuperAdmin(st)
	} else {
		konfig, err = kubeconfig.Download(st)
	}
	if err != nil {
		return err
	}

	fmt.Println(string(konfig))

	return nil
}
