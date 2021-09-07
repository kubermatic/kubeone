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
	"path/filepath"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/tasks"
)

type installOpts struct {
	globalOptions
	BackupFile string `longflag:"backup" shortflag:"b"`
	NoInit     bool   `longflag:"no-init"`
	Force      bool   `longflag:"force"`
}

func (opts *installOpts) BuildState() (*state.State, error) {
	s, err := opts.globalOptions.BuildState()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build state")
	}

	s.ForceInstall = opts.Force
	s.BackupFile = opts.BackupFile
	if s.BackupFile == "" {
		fullPath, _ := filepath.Abs(opts.ManifestFile)
		clusterName := s.Cluster.Name
		s.BackupFile = filepath.Join(filepath.Dir(fullPath), fmt.Sprintf("%s.tar.gz", clusterName))
	}

	// refuse to overwrite existing backups (NB: since we attempt to
	// write to the file later on to check for write permissions, we
	// inadvertently create a zero byte file even if the first step
	// of the installer fails; for this reason it's okay to find an
	// existing, zero byte backup)
	stat, err := os.Stat(s.BackupFile)
	if err != nil && stat != nil && stat.Size() > 0 {
		return nil, errors.Errorf("backup %s already exists, refusing to overwrite", opts.BackupFile)
	}

	// try to write to the file before doing anything else
	f, err := os.OpenFile(s.BackupFile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open %q for writing", opts.BackupFile)
	}

	return s, f.Close()
}

// installCmd setups install command
func installCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &installOpts{}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Kubernetes",
		Long: heredoc.Doc(`
			Install Kubernetes on pre-existing machines

			This command takes KubeOne manifest which contains information about hosts and how the cluster should be provisioned.
			It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.
		`),
		Example: `kubeone install -m mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			opts.globalOptions = *gopts
			return runInstall(opts)
		},
	}

	cmd.Flags().StringVarP(
		&opts.BackupFile,
		longFlagName(opts, "BackupFile"),
		shortFlagName(opts, "BackupFile"),
		"",
		"path to where the PKI backup .tar.gz file should be placed (default: location of cluster config file)")

	cmd.Flags().BoolVar(
		&opts.NoInit,
		longFlagName(opts, "NoInit"),
		false,
		"don't initialize the cluster (only install binaries)")

	cmd.Flags().BoolVar(
		&opts.Force,
		longFlagName(opts, "Force"),
		false,
		"use force to install new binary versions (!dangerous!)")

	return cmd
}

// runInstall provisions Kubernetes on the provided machines
func runInstall(opts *installOpts) error {
	s, err := opts.BuildState()
	if err != nil {
		return errors.Wrap(err, "failed to initialize State")
	}

	// Validate credentials
	_, err = credentials.ProviderCredentials(s.Cluster.CloudProvider, opts.CredentialsFile)
	if err != nil {
		return errors.Wrap(err, "failed to validate credentials")
	}

	if opts.NoInit {
		return errors.Wrap(tasks.WithBinariesOnly(nil).Run(s), "failed to install kubernetes binaries")
	}

	// Probe the cluster for the actual state and the needed tasks.
	probbing := tasks.WithHostnameOS(nil)
	probbing = tasks.WithProbes(probbing)

	if err = probbing.Run(s); err != nil {
		return err
	}

	return errors.Wrap(tasks.WithFullInstall(nil).Run(s), "failed to install the cluster")
}
