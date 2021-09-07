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

	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/tasks"
)

func migrateCmd(fs *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Commands for running different migrations",
	}

	cmd.AddCommand(migrateToContainerdCmd(fs))
	cmd.AddCommand(migrateToCCMCSICmd(fs))
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
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(fs)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			return runMigrateToContainerd(gopts)
		},
	}
}

func runMigrateToContainerd(opts *globalOptions) error {
	s, err := opts.BuildState()
	if err != nil {
		return errors.Wrap(err, "failed to initialize State")
	}

	// Probe the cluster for the actual state and the needed tasks.
	probbing := tasks.WithHostnameOS(nil)
	probbing = tasks.WithProbes(probbing)

	if err = probbing.Run(s); err != nil {
		return err
	}

	if !s.LiveCluster.IsProvisioned() {
		return errors.New("the target cluster is not provisioned")
	}
	if !s.LiveCluster.Healthy() {
		return errors.New("the target cluster is not healthy, please run 'kubeone apply' first")
	}

	return errors.Wrap(tasks.WithContainerDMigration(nil).Run(s), "failed to get cluster status")
}

type migrateCCMOptions struct {
	globalOptions
	AutoApprove       bool `longflag:"auto-approve" shortflag:"y"`
	CompleteMigration bool `longflag:"complete"`
}

func (opts *migrateCCMOptions) buildCCMMigrationState() (*state.State, error) {
	s, err := opts.globalOptions.BuildState()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build state")
	}

	s.CCMMigration = true
	s.CCMMigrationComplete = opts.CompleteMigration

	return s, nil
}

func migrateToCCMCSICmd(fs *pflag.FlagSet) *cobra.Command {
	opts := &migrateCCMOptions{}

	cmd := &cobra.Command{
		Use:   "to-ccm-csi",
		Short: "Migrate live cluster from the in-tree cloud provider to external cloud-controller-manager (CCM) and CSI plugin",
		Long: heredoc.Doc(`
			Following the in-tree cloud provider deprecation (http://kep.k8s.io/2395),
			this command helps to migrate existing clusters from the in-tree cloud provider to external
			cloud-controller-manager (CCM) and CSI plugin.

			Note: if your cluster was created with .cloudProvider.external enabled, the CCM/CSI migration is not needed
			because the cluster is already using external CCM.

			Migration is currently available for OpenStack and vSphere. Other providers will be added in future KubeOne releases.

			The migration is done in two phases:

			  * Phase 1: deploy external CCM and CSI plugin, while leaving in-tree provider enabled.
			    Kubernetes API server and kube-controller-manager are configured to:
				  - use controllers integrated in external CCM instead of in-tree cloud provider
				    for all cloud-related operations
				  - redirect all volumes-related operations to the CSI plugin
				The existing worker nodes will continue to use in-tree provider (that's why it's still left enabled),
				so therefore, all worker nodes managed by machine-controller must be rolled out after phase 1 is complete.

			  * Phase 2: complete the CCM/CSI migration by fully-disabling in-tree provider. To trigger the phase 2,
			    users need to run "kubeone migrate to-ccm-csi" command with the "--complete" flag. This should be
			    done after all worker nodes managed by machine-controller are rolled-out.

			Make sure to familiarize yourself with the CCM/CSI migration requirements by checking the following document:
			https://docs.kubermatic.com/kubeone/v1.3/guides/ccm_csi_migration/
		`),
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(fs)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			opts.globalOptions = *gopts

			return runMigrateToCCMCSI(opts)
		},
	}

	cmd.Flags().BoolVarP(
		&opts.AutoApprove,
		longFlagName(opts, "AutoApprove"),
		shortFlagName(opts, "AutoApprove"),
		false,
		"auto approve plan")

	cmd.Flags().BoolVarP(
		&opts.CompleteMigration,
		longFlagName(opts, "CompleteMigration"),
		shortFlagName(opts, "CompleteMigration"),
		false,
		"complete ccm/csi migration")

	return cmd
}

func runMigrateToCCMCSI(opts *migrateCCMOptions) error {
	s, err := opts.buildCCMMigrationState()
	if err != nil {
		return errors.Wrap(err, "failed to initialize State")
	}

	// Validate credentials
	_, err = credentials.ProviderCredentials(s.Cluster.CloudProvider, opts.CredentialsFile)
	if err != nil {
		return errors.Wrap(err, "failed to validate credentials")
	}

	// Probe the cluster for the actual state and the needed tasks.
	probbing := tasks.WithHostnameOS(nil)
	probbing = tasks.WithProbes(probbing)

	if err = probbing.Run(s); err != nil {
		return err
	}

	if !s.LiveCluster.IsProvisioned() {
		return errors.New("the target cluster is not provisioned")
	}
	if !s.LiveCluster.Healthy() {
		return errors.New("the target cluster is not healthy, please run 'kubeone apply' first")
	}

	s.Logger.Warnln("This command will migrate your cluster from in-tree cloud provider to the external CCM and CSI plugin.")
	s.Logger.Warnln("Make sure to familiarize yourself with the process by checking the following document:")
	s.Logger.Warnln("https://docs.kubermatic.com/kubeone/v1.3/guides/ccm_csi_migration/")

	confirm, err := confirmCommand(opts.AutoApprove)
	if err != nil {
		return err
	}

	if !confirm {
		s.Logger.Println("Operation canceled.")
		return nil
	}

	return errors.Wrap(tasks.WithCCMCSIMigration(nil).Run(s), "failed to migrate to ccm/csi")
}
