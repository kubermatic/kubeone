package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// InstallCommand returns the structure for declaring the "install" subcommand.
func InstallCommand(logger *logrus.Logger) cli.Command {
	return cli.Command{
		Name:      "install",
		Usage:     "Installs Kubernetes onto pre-existing machines",
		ArgsUsage: "CLUSTER_FILE",
		Action:    InstallAction(logger),
		Flags: []cli.Flag{
			cli.StringFlag{
				EnvVar: "TF_OUTPUT",
				Name:   "tfjson, t",
				Usage:  "path to terraform output JSON or - for stdin",
				Value:  "",
			},
			cli.StringFlag{
				Name:  "backup, b",
				Usage: "path to where the PKI backup .tar.gz file should be placed (default: location of cluster config file)",
				Value: "",
			},
		},
	}
}

// InstallAction wrapper for logger
func InstallAction(logger *logrus.Logger) cli.ActionFunc {
	return handleErrors(logger, setupLogger(logger, func(ctx *cli.Context) error {
		clusterFile := ctx.Args().First()
		if clusterFile == "" {
			return errors.New("no cluster config file given")
		}

		cluster, err := loadClusterConfig(clusterFile)
		if err != nil {
			return fmt.Errorf("failed to load cluster: %v", err)
		}

		tf := ctx.String("tfjson")
		if err = applyTerraform(tf, cluster); err != nil {
			return err
		}

		if err = cluster.DefaultAndValidate(); err != nil {
			return err
		}

		options, err := createInstallerOptions(clusterFile, cluster, ctx)
		if err != nil {
			return fmt.Errorf("failed to create installer options: %v", err)
		}
		if err = applyTerraform(tf, cluster); err != nil {
			return fmt.Errorf("failed to setup PKI backup: %v", err)
		}

		worker := installer.NewInstaller(cluster, logger)
		_, err = worker.Install(options)

		return err
	}))
}

func createInstallerOptions(clusterFile string, cluster *config.Cluster, ctx *cli.Context) (*installer.Options, error) {
	backupFile := ctx.String("backup")
	if len(backupFile) == 0 {
		fullPath, _ := filepath.Abs(clusterFile)
		clusterName := cluster.Name

		backupFile = filepath.Join(filepath.Dir(fullPath), fmt.Sprintf("%s.tar.gz", clusterName))
	}

	// refuse to overwrite existing backups (NB: since we attempt to
	// write to the file later on to check for write permissions, we
	// inadvertently create a zero byte file even if the first step
	// of the installer fails; for this reason it's okay to find an
	// existing, zero byte backup)
	stat, err := os.Stat(backupFile)
	if err != nil && stat != nil && stat.Size() > 0 {
		return nil, fmt.Errorf("backup %s already exists, refusing to overwrite", backupFile)
	}

	// try to write to the file before doing anything else
	f, err := os.OpenFile(backupFile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s for writing", backupFile)
	}
	defer f.Close()

	return &installer.Options{
		BackupFile: backupFile,
		Verbose:    ctx.GlobalBool("verbose"),
	}, nil
}
