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
	"os"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"
)

// KubeconfigCommand returns the structure for declaring the "install" subcommand.
func kubeconfigCmd(rootFlags *pflag.FlagSet) *cobra.Command {
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

			st, err := gopts.BuildState()
			if err != nil {
				return err
			}

			konfig, err := kubeconfig.Download(st)
			if err != nil {
				return err
			}
			fmt.Println(string(konfig))

			return nil
		},
	}

	cmd.AddCommand(kubeconfigGenerateCmd(rootFlags))

	return cmd
}

type kubeconfigGenerateOpts struct {
	globalOptions
	CommonName        string        `longflag:"cn"`
	OrganizationNames []string      `longflag:"on"`
	ShortSuperAdmin   bool          `longflag:"super-admin" shortflag:"s"`
	TTL               time.Duration `longflag:"ttl"`
}

func kubeconfigGenerateCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &kubeconfigGenerateOpts{}

	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate kubeconfig",
		Long:    "Generate kubeconfig with given certificate parameters.",
		Example: `kubeone kubeconfig generate -m mycluster.yaml -t tf.json --super-admin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}
			opts.globalOptions = *gopts

			st, err := gopts.BuildState()
			if err != nil {
				return err
			}

			return runKubeconfigGenerate(st, opts)
		},
	}

	cmd.Flags().StringVar(
		&opts.CommonName,
		longFlagName(opts, "CommonName"),
		"",
		"CommonName (CN) for the generated client certificate.")

	cmd.Flags().StringArrayVar(
		&opts.OrganizationNames,
		longFlagName(opts, "OrganizationNames"),
		[]string{},
		"OrganizationName (ON) for the generated client certificate.")

	cmd.Flags().BoolVarP(
		&opts.ShortSuperAdmin,
		longFlagName(opts, "ShortSuperAdmin"),
		shortFlagName(opts, "ShortSuperAdmin"),
		false,
		"Generate superadmin kubeconfig, shorthand for --cn <USER>@<HOSTNAME> --on system:masters")

	cmd.Flags().DurationVar(
		&opts.TTL,
		longFlagName(opts, "TTL"),
		1*time.Hour,
		"Time To Live for the generated certificate.")

	return cmd
}

func runKubeconfigGenerate(st *state.State, opts *kubeconfigGenerateOpts) error {
	cn := opts.CommonName
	on := opts.OrganizationNames

	if opts.ShortSuperAdmin {
		on = []string{"system:masters"}
	}

	if len(on) == 0 {
		return fail.NewConfigError("--on flag", "can not be empty")
	}

	if cn == "" {
		hostname, _ := os.Hostname()
		username := os.Getenv("USER")
		if username == "" {
			username = "DEFAULT"
		}
		cn = fmt.Sprintf("%s@%s", username, hostname)
	}

	konfig, err := kubeconfig.GenerateSuperAdmin(st, cn, on, opts.TTL)
	if err != nil {
		return err
	}

	fmt.Println(string(konfig))

	return nil
}
