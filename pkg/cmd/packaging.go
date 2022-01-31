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
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func completionCmd(rootCmd *cobra.Command) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "completion <bash|zsh>",
		Short: "Generates completion scripts for bash and zsh",
		Long: heredoc.Doc(`
			To load completion run into your current shell run

			. <(kubeone completion <shell>)
		`),
		Example:   "kubeone completion bash",
		ValidArgs: []string{"bash", "zsh"},
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			switch args[0] {
			case "bash":
				err = rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				err = rootCmd.GenZshCompletion(os.Stdout)
			}

			return
		},
	}

	return cmd
}

func documentCmd(rootCmd *cobra.Command) *cobra.Command {
	var path string
	var cmd = &cobra.Command{
		Use:   "document <man|md|rest|yaml>",
		Short: "Generates documentation",
		Long: heredoc.Doc(`
			Documentation can be generated as man pages, markdown, restructured text docs or yaml
		`),
		Example:   "kubeone document man",
		ValidArgs: []string{"man", "md", "rest", "yaml"},
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			switch args[0] {
			case "man":
				header := &doc.GenManHeader{
					Title:   "KubeOne",
					Section: "1",
				}
				err = doc.GenManTree(rootCmd, header, path)
			case "md":
				err = doc.GenMarkdownTree(rootCmd, path)
			case "rest":
				err = doc.GenReSTTree(rootCmd, path)
			case "yaml":
				err = doc.GenYamlTree(rootCmd, path)
			}

			return
		},
	}
	cmd.Flags().StringVarP(&path, "output-dir", "o", "/tmp/", "Directory to populate with documentation")

	return cmd
}
