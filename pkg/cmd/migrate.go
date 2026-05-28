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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/tasks"
)

func migrateCmd(fs *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Commands for running different migrations",
	}
	cmd.AddCommand(migrateToContainerdCmd(fs))
	cmd.AddCommand(migrateToCCMCSICmd())

	return cmd
}

func migrateToContainerdCmd(fs *pflag.FlagSet) *cobra.Command {
	return &cobra.Command{
		Use:   "to-containerd",
		Short: "Migrate live cluster from docker to containerd",
		Long: heredoc.Doc(`

			Following the dockershim deprecation https://kubernetes.io/blog/2020/12/02/dockershim-faq/
			this command helps to migrate Container Runtime to ContainerD.
		`),
		Hidden: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(fs)
			if err != nil {
				return err
			}

			return runMigrateToContainerd(gopts)
		},
	}
}

func runMigrateToContainerd(opts *globalOptions) error {
	s, err := opts.BuildState()
	if err != nil {
		return err
	}

	if err = tasks.WithFindControlPlane(nil).Run(s); err != nil {
		return err
	}

	// Probe the cluster for the actual state and the needed tasks.
	probbing := tasks.WithHostnameOS(nil)
	probbing = tasks.WithProbes(probbing)

	if err = probbing.Run(s); err != nil {
		return err
	}

	if !s.LiveCluster.IsProvisioned() {
		return fail.RuntimeError{
			Op:  "containerd migration",
			Err: errors.New("the target cluster is not provisioned"),
		}
	}

	if !s.LiveCluster.Healthy() {
		return fail.RuntimeError{
			Op:  "containerd migration",
			Err: errors.New("the target cluster is not healthy, please run 'kubeone apply' first"),
		}
	}

	return tasks.WithContainerDMigration(nil).Run(s)
}

func migrateToCCMCSICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "to-ccm-csi",
		Short: "This command is NO OP since Kubernetes moved on from CSI migration and now all persistent volumes are managed by CSI",
		Long:  heredoc.Doc(``),
		RunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
	}

	var (
		autoApprove       bool
		completeMigration bool
	)

	cmd.Flags().BoolVarP(
		&autoApprove,
		"auto-approve",
		"y",
		false,
		"auto approve plan")

	cmd.Flags().BoolVarP(
		&completeMigration,
		"complete",
		"",
		false,
		"complete ccm/csi migration")

	return cmd
}
