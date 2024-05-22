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

	"k8c.io/kubeone/pkg/dashboard"
)

type uiOpts struct {
	globalOptions
	Port int `longflag:"port" shortflag:"port"`
}

// uiCmd setups ui command
func uiCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &uiOpts{}

	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Show UI",
		Long: heredoc.Doc(`
			Starts a webserver providing a minimalistic overview of the KubeOne Kubernetes Cluster.

			By default port 8080 will be used, you can customize the port via the port flag. 
		`),
		SilenceErrors: true,
		Args:          cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}
			opts.globalOptions = *gopts

			state, err := opts.BuildState()
			if err != nil {
				return err
			}

			return dashboard.Serve(state, opts.Port)
		},
	}

	cmd.Flags().IntVarP(
		&opts.Port,
		longFlagName(opts, "port"),
		shortFlagName(opts, "port"),
		8080,
		"port on which webserver is running")

	return cmd
}
