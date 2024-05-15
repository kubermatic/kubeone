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
	"github.com/spf13/pflag"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"k8c.io/kubeone/pkg/dashboard"
)

type uiOpts struct {
	globalOptions
	// AutoApprove    bool `longflag:"auto-approve" shortflag:"y"`
	// DestroyWorkers bool `longflag:"destroy-workers"`
	// RemoveBinaries bool `longflag:"remove-binaries"`
	// TODO
}

// TODO
func uiCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &uiOpts{}

	cmd := &cobra.Command{
		Use:   "ui",
		Short: "TODO",
		Long: heredoc.Doc(`
			TODO
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
			return dashboard.Serve(state)
		},
	}
	return cmd
}
