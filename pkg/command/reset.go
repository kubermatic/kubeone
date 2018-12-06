package command

import (
	"errors"
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// ResetCommand returns the structure for declaring the "reset" subcommand.
func ResetCommand(logger *logrus.Logger) cli.Command {
	return cli.Command{
		Name:      "reset",
		Usage:     "Undos all changes made by KubeOne to the configured machines",
		ArgsUsage: "CLUSTER_FILE",
		Action:    ResetAction(logger),
		Flags: []cli.Flag{
			cli.StringFlag{
				EnvVar: "TF_OUTPUT",
				Name:   "tfjson, t",
				Usage:  "path to terraform output JSON or - for stdin",
				Value:  "",
			},
			cli.BoolFlag{
				Name:  "destroy-workers",
				Usage: "de-provision all worker machines before resetting cluster",
			},
		},
	}
}

// ResetAction handles the "reset" subcommand.
func ResetAction(logger *logrus.Logger) cli.ActionFunc {
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

		if err = cluster.Validate(); err != nil {
			return fmt.Errorf("cluster is invalid: %v", err)
		}

		options := &installer.Options{
			Verbose:        ctx.GlobalBool("verbose"),
			DestroyWorkers: ctx.Bool("destroy-workers"),
		}

		worker := installer.NewInstaller(cluster, logger)
		_, err = worker.Reset(options)

		return err
	}))
}
