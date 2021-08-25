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

func migrateToCCMCSICmd(fs *pflag.FlagSet) *cobra.Command {
	return &cobra.Command{
		Use:   "to-ccm-csi",
		Short: "Migrate live cluster from the in-tree cloud provider to external CCM and CSI plugin",
		// TODO(xmudrii): Add which providers are supported in the long description.
		Long: heredoc.Doc(`
			Following the in-tree cloud provider deprecation http://kep.k8s.io/2395
			this command helps to migrate from the in-tree cloud provider to external CCM and CSI plugin.
			This command is currently only available for some providers. We'll extend it for all providers with
			in-tree cloud provider implementation in the future.
		`),
		// TODO: Remove hidden once complete
		Hidden: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			gopts, err := persistentGlobalOptions(fs)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			return runMigrateToCCMCSI(gopts)
		},
	}
}

func runMigrateToContainerd(opts *globalOptions) error {
	s, err := opts.BuildState()
	if err != nil {
		return errors.Wrap(err, "failed to initialize State")
	}

	return errors.Wrap(tasks.WithContainerDMigration(nil).Run(s), "failed to get cluster status")
}

func runMigrateToCCMCSI(opts *globalOptions) error {
	s, err := opts.BuildState()
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

	return errors.Wrap(tasks.WithCCMCSIMigration(nil).Run(s), "failed to migrate to ccm/csi")
}
