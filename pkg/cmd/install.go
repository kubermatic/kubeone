package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type installOptions struct {
	options
	Manifest   string
	BackupFile string
}

// installCmd setups install command
func installCmd() *cobra.Command {
	var io = &installOptions{}
	var installCmd = &cobra.Command{
		Use:   "install <manifest>",
		Short: "Install Kubernetes",
		Long: `Install Kubernetes on pre-existing machines

This command takes KubeOne manifest which contains information about hosts and how the cluster should be provisioned.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := initLogger()
			io.TerraformState = o.TerraformState
			io.Verbose = o.Verbose

			if len(args) != 1 {
				return errors.New("expected path to a cluster config file as an argument")
			}

			io.Manifest = args[0]
			if io.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runInstall(logger, io)
		},
	}

	installCmd.Flags().StringVarP(&io.BackupFile, "backup", "b", "", "path to where the PKI backup .tar.gz file should be placed (default: location of cluster config file)")

	return installCmd
}

// runInstall provisions Kubernetes on the provided machines
func runInstall(logger *logrus.Logger, installOptions *installOptions) error {
	cluster, err := loadClusterConfig(installOptions.Manifest)
	if err != nil {
		return fmt.Errorf("failed to load cluster: %v", err)
	}

	if err = applyTerraform(installOptions.TerraformState, cluster); err != nil {
		return err
	}

	options, err := createInstallerOptions(installOptions.Manifest, cluster, installOptions)
	if err != nil {
		return fmt.Errorf("failed to create installer options: %v", err)
	}

	// TODO(xmudrii): why do we call terraform two times?
	if err = applyTerraform(installOptions.TerraformState, cluster); err != nil {
		return fmt.Errorf("failed to setup PKI backup: %v", err)
	}

	if err = cluster.DefaultAndValidate(); err != nil {
		return err
	}

	return installer.NewInstaller(cluster, logger).Install(options)
}

func createInstallerOptions(clusterFile string, cluster *config.Cluster, options *installOptions) (*installer.Options, error) {
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
		return nil, fmt.Errorf("backup %s already exists, refusing to overwrite", options.BackupFile)
	}

	// try to write to the file before doing anything else
	f, err := os.OpenFile(options.BackupFile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s for writing", options.BackupFile)
	}
	defer f.Close()

	return &installer.Options{
		BackupFile: options.BackupFile,
		Verbose:    options.Verbose,
	}, nil
}
