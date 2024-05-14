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
)

// TODO
func uiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "TODO",
		Long: heredoc.Doc(`
			TODO
		`),
		SilenceErrors: true,
		Args:          cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("hello ui")
			return nil
		},
	}

	return cmd
}
