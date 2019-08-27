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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/credentials"
	"github.com/kubermatic/kubeone/pkg/installer"
)

type installOptions struct {
	globalOptions
	Manifest   string
	BackupFile string
}

// installCmd setups install command
func installCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	iopts := &installOptions{}
	cmd := &cobra.Command{
		Use:   "install <manifest>",
		Short: "Install Kubernetes",
		Long: `
Install Kubernetes on pre-existing machines

This command takes KubeOne manifest which contains information about hosts and how the cluster should be provisioned.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.
`,
		Args:    cobra.ExactArgs(1),
		Example: `kubeone install mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			logger := initLogger(gopts.Verbose)
			iopts.TerraformState = gopts.TerraformState
			iopts.Verbose = gopts.Verbose
			iopts.CredentialsFilePath = gopts.CredentialsFilePath

			iopts.Manifest = args[0]
			if iopts.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runInstall(logger, iopts)
		},
	}

	cmd.Flags().StringVarP(&iopts.BackupFile, "backup", "b", "", "path to where the PKI backup .tar.gz file should be placed (default: location of cluster config file)")

	return cmd
}

// runInstall provisions Kubernetes on the provided machines
func runInstall(logger *logrus.Logger, installOptions *installOptions) error {
	cluster, err := loadClusterConfig(installOptions.Manifest, installOptions.TerraformState, logger)
	if err != nil {
		return errors.Wrap(err, "failed to load cluster")
	}

	options, err := createInstallerOptions(installOptions.Manifest, cluster, installOptions)
	if err != nil {
		return errors.Wrap(err, "failed to create installer options")
	}

	// Validate credentials
	_, err = credentials.ProviderCredentials(cluster.CloudProvider.Name, installOptions.CredentialsFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to validate credentials")
	}

	return installer.NewInstaller(cluster, logger).Install(options)
}

func createInstallerOptions(clusterFile string, cluster *kubeoneapi.KubeOneCluster, options *installOptions) (*installer.Options, error) {
	if len(options.BackupFile) == 0 {
		fullPath, _ := filepath.Abs(clusterFile)
		clusterName := cluster.Name
		options.BackupFile = filepath.Join(filepath.Dir(fullPath), fmt.Sprintf("%s.tar.gz", clusterName))
	}

	// refuse to overwrite existing backups (NB: since we attempt to
	// write to the file later on to check for write permissions, we
	// inadvertently create a zero byte file even if the first step
	// of the installer fails; for this reason it's okay to find an
	// existing, zero byte backup)
	stat, err := os.Stat(options.BackupFile)
	if err != nil && stat != nil && stat.Size() > 0 {
		return nil, errors.Errorf("backup %s already exists, refusing to overwrite", options.BackupFile)
	}

	// try to write to the file before doing anything else
	f, err := os.OpenFile(options.BackupFile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, errors.Errorf("cannot open %s for writing", options.BackupFile)
	}
	defer f.Close()

	return &installer.Options{
		CredentialsFile: options.CredentialsFilePath,
		BackupFile:      options.BackupFile,
		Verbose:         options.Verbose,
	}, nil
}
