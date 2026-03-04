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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	clientv3 "go.etcd.io/etcd/client/v3"

	"k8c.io/kubeone/pkg/etcdutil"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/tabwriter"
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
		etcdDefragmentCmd(rootFlags),
		etcdDisarmCmd(rootFlags),
		etcdMembersCmd(rootFlags),
		etcdSnapshotCmd(rootFlags),
	)

	return cmd
}

type etcdMembersOpts struct {
	globalOptions
	OutputFormat string `longflag:"output" shortflag:"o"`
}

type etcdMember struct {
	ID         uint64
	Name       string
	PeerURLs   []string
	ClientURLs []string
	IsLearner  bool
	Alarms     []string
}

func (em etcdMember) TableHeader() string {
	return "ID\tNAME\tPEER-URLS\tCLIENT-URLS\tIS-LEARNER\tALARMS"
}

func (em etcdMember) TableFormat() string {
	return "%d\t%s\t%v\t%v\t%v\t%v\n"
}

func etcdMembersCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &etcdMembersOpts{}

	cmd := &cobra.Command{
		Use:           "members",
		Short:         "List etcd members",
		Long:          "List the current members of the etcd cluster.",
		SilenceErrors: true,
		Example:       `kubeone etcd members -m mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			s, err := gopts.BuildState()
			if err != nil {
				return err
			}

			etcdcli, err := etcdutil.NewClient(s)
			if err != nil {
				return err
			}
			defer etcdcli.Close()

			memberList, err := etcdcli.MemberList(s.Context)
			if err != nil {
				return fail.Etcd(err, "member listing")
			}

			alarmList, err := clientv3.NewMaintenance(etcdcli).AlarmList(s.Context)
			if err != nil {
				return fail.Etcd(err, "alarms listing")
			}

			alarmsByMember := make(map[uint64][]string)
			for _, a := range alarmList.Alarms {
				alarmsByMember[a.MemberID] = append(alarmsByMember[a.MemberID], a.Alarm.String())
			}

			var response []etcdMember
			for _, m := range memberList.Members {
				alarms := alarmsByMember[m.ID]
				if alarms == nil {
					alarms = []string{}
				}
				response = append(response, etcdMember{
					ID:         m.ID,
					Name:       m.Name,
					PeerURLs:   m.PeerURLs,
					ClientURLs: m.ClientURLs,
					IsLearner:  m.IsLearner,
					Alarms:     alarms,
				})
			}

			switch opts.OutputFormat {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")

				return enc.Encode(response)
			default:
				tab := tabwriter.NewWithPadding(os.Stdout, 2)
				member := etcdMember{}
				fmt.Fprintln(tab, member.TableHeader())
				tableFormat := member.TableFormat()

				for _, m := range response {
					fmt.Fprintf(tab, tableFormat,
						m.ID,
						m.Name,
						m.PeerURLs,
						m.ClientURLs,
						m.IsLearner,
						m.Alarms,
					)
				}

				return tab.Flush()
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

type etcdDisarmOpts struct {
	globalOptions
	All bool `longflag:"all"`
}

func etcdDisarmCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &etcdDisarmOpts{}

	cmd := &cobra.Command{
		Use:           "disarm [member-name]",
		Short:         "Disarm etcd alarms",
		Long:          "Disarm all active alarms on a specific etcd member (by name), or on all members with --all.",
		SilenceErrors: true,
		Example: heredoc.Doc(`
			# Disarm alarms on a specific member
			kubeone etcd disarm master-0 -m mycluster.yaml -t terraformoutput.json

			# Disarm alarms on all members
			kubeone etcd disarm --all -m mycluster.yaml -t terraformoutput.json
		`),
		Args: func(_ *cobra.Command, args []string) error {
			if !opts.All && len(args) == 0 {
				return fmt.Errorf("requires a member name argument or --all flag")
			}
			if opts.All && len(args) > 0 {
				return fmt.Errorf("--all and a member name are mutually exclusive")
			}
			if len(args) > 1 {
				return fmt.Errorf("accepts at most one member name, got %d", len(args))
			}

			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			s, err := gopts.BuildState()
			if err != nil {
				return err
			}

			etcdcli, err := etcdutil.NewClient(s)
			if err != nil {
				return err
			}
			defer etcdcli.Close()

			maintenance := clientv3.NewMaintenance(etcdcli)

			if opts.All {
				err = disarmAll(s.Context, maintenance)
				if err != nil {
					return fail.Etcd(err, "disarming all alarms")
				}

				s.Logger.Infof("Disarmed all alarms on all members")

				return nil
			}

			memberName := args[0]

			etcdRing, err := etcdcli.MemberList(s.Context)
			if err != nil {
				return fail.Etcd(err, "member listing")
			}

			var memberID uint64
			for _, m := range etcdRing.Members {
				if m.Name == memberName {
					memberID = m.ID

					break
				}
			}

			if memberID == 0 {
				return fail.Etcd(fmt.Errorf("etcd member %q not found", memberName), "searching memberID")
			}

			alarmList, err := maintenance.AlarmList(s.Context)
			if err != nil {
				return fail.Etcd(err, "alarm listing")
			}

			disarmed := 0
			for _, a := range alarmList.Alarms {
				if a.MemberID != memberID {
					continue
				}

				_, err = maintenance.AlarmDisarm(s.Context, (*clientv3.AlarmMember)(a))
				if err != nil {
					return fail.Etcd(err, "disarming alarm %s on member %s", a.Alarm, memberName)
				}

				disarmed++
			}

			s.Logger.Infof("Disarmed %d alarm(s) on member %q.\n", disarmed, memberName)

			return nil
		},
	}

	cmd.Flags().BoolVar(
		&opts.All,
		longFlagName(opts, "All"),
		false,
		"disarm alarms on all etcd members")

	return cmd
}

func disarmAll(ctx context.Context, maintenance clientv3.Maintenance) error {
	_, err := maintenance.AlarmDisarm(ctx, &clientv3.AlarmMember{})

	return fail.Etcd(err, "disarming all members")
}

func etcdDefragmentCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "defragment <member-name>",
		Short:         "Defragment etcd members",
		Long:          "Defragment the etcd storage of a specific member (by name).",
		SilenceErrors: true,
		Example: heredoc.Doc(`
			# Defragment a specific member
			kubeone etcd defragment control-plane-0 -m mycluster.yaml -t terraformoutput.json
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			s, err := gopts.BuildState()
			if err != nil {
				return err
			}

			etcdcli, err := etcdutil.NewClient(s)
			if err != nil {
				return err
			}
			defer etcdcli.Close()

			maintenance := clientv3.NewMaintenance(etcdcli)

			etcdRing, err := etcdcli.MemberList(s.Context)
			if err != nil {
				return fail.Etcd(err, "member listing")
			}
			memberName := args[0]

			var endpoints []string
			for _, m := range etcdRing.Members {
				if m.Name == memberName {
					endpoints = m.ClientURLs

					break
				}
			}

			if len(endpoints) == 0 {
				return fail.Etcd(fmt.Errorf("etcd member %q not found", memberName), "searching member endpoints")
			}

			for _, endpoint := range endpoints {
				_, err = maintenance.Defragment(s.Context, endpoint)
				if err != nil {
					return fail.Etcd(err, "defragmenting member %s endpoint %s", memberName, endpoint)
				}
				s.Logger.Infof("Defragmented member %q endpoint %q.", memberName, endpoint)
			}

			return nil
		},
	}

	return cmd
}

func etcdSnapshotCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "snapshot <file>",
		Short:         "Save an etcd snapshot to a file",
		Long:          "Save a point-in-time snapshot of the etcd cluster to a local file.",
		SilenceErrors: true,
		Example:       `kubeone etcd snapshot etcd.db -m mycluster.yaml -t terraformoutput.json`,
		Args:          cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			s, err := gopts.BuildState()
			if err != nil {
				return err
			}

			etcdcli, err := etcdutil.NewClient(s)
			if err != nil {
				return err
			}
			defer etcdcli.Close()

			maintenance := clientv3.NewMaintenance(etcdcli)

			snapshotResp, err := maintenance.SnapshotWithVersion(s.Context)
			if err != nil {
				return fail.Etcd(err, "requesting snapshot")
			}
			defer snapshotResp.Snapshot.Close()

			outPath := args[0]
			var output io.WriteCloser = os.Stdout
			defer output.Close()

			if outPath != "-" {
				f, errF := os.Create(outPath)
				if errF != nil {
					return fail.Runtime(errF, "creating snapshot file %q", outPath)
				}
				output = f
			}

			if _, err = io.Copy(output, snapshotResp.Snapshot); err != nil {
				return fail.Runtime(err, "writing snapshot to %q", outPath)
			}

			s.Logger.Infof("Snapshot saved to %q (etcd version: %s).\n", outPath, snapshotResp.Version)

			return nil
		},
	}

	return cmd
}
