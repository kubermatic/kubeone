/*
Copyright 2026 The KubeOne Authors.

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
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/clusterstatus/etcdstatus"
)

func etcdOperationsCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "etcd",
		Short: "Perform etcd operations",
		Long: heredoc.Doc(`
			Perform operations on the etcd cluster members of a KubeOne-managed Kubernetes cluster.
		`),
	}

	cmd.AddCommand(
		etcdMembersCmd(rootFlags),
	)

	return cmd
}

type etcdMembersOpts struct {
	globalOptions
	OutputFormat string `longflag:"output" shortflag:"o"`
}

func etcdMembersCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &etcdMembersOpts{}

	cmd := &cobra.Command{
		Use:           "members",
		Short:         "List etcd members",
		Long:          "List the current members of the etcd cluster.",
		SilenceErrors: true,
		Example:       `kubeone operations etcd members -m mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			s, err := gopts.BuildState()
			if err != nil {
				return err
			}

			memberList, err := etcdstatus.MemberList(s)
			if err != nil {
				return err
			}

			switch opts.OutputFormat {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")

				return enc.Encode(memberList.Members)
			default:
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
				fmt.Fprintln(w, "ID\tNAME\tPEER URLS\tCLIENT URLS\tIS LEARNER")
				for _, m := range memberList.Members {
					fmt.Fprintf(w, "%d\t%s\t%v\t%v\t%v\n",
						m.ID,
						m.Name,
						m.PeerURLs,
						m.ClientURLs,
						m.IsLearner,
					)
				}

				return w.Flush()
			}
		},
	}

	cmd.Flags().StringVarP(
		&opts.OutputFormat,
		longFlagName(opts, "OutputFormat"),
		shortFlagName(opts, "OutputFormat"),
		"table",
		"output format (table|json)")

	return cmd
}
