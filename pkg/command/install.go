package command

import (
	"errors"
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// InstallCommand returns the structure for declaring the "install" subcommand.
func InstallCommand(logger *logrus.Logger) cli.Command {
	return cli.Command{
		Name:      "install",
		Usage:     "Installs Kubernetes onto pre-existing machines",
		ArgsUsage: "MANIFEST_FILE",
		Action:    InstallAction(logger),
		Flags: []cli.Flag{
			cli.StringFlag{
				EnvVar: "TF_OUTPUT",
				Name:   "tfjson, t",
				Usage:  "path to terraform output JSON or - for stdin",
				Value:  "",
			},
		},
	}
}

// InstallAction wrapper for logger
func InstallAction(logger *logrus.Logger) cli.ActionFunc {
	return handleErrors(logger, setupLogger(logger, func(ctx *cli.Context) error {
		manifestFile := ctx.Args().First()
		if manifestFile == "" {
			return errors.New("no manifest file given")
		}

		manifest, err := loadManifest(manifestFile)
		if err != nil {
			return fmt.Errorf("failed to load manifest: %v", err)
		}

		tf := ctx.String("tfjson")
		if err = applyTerraform(tf, manifest); err != nil {
			return err
		}

		if err = manifest.Validate(); err != nil {
			return fmt.Errorf("manifest is invalid: %v", err)
		}

		worker := installer.NewInstaller(manifest, logger)
		_, err = worker.Install(ctx.GlobalBool("verbose"))

		return err
	}))
}
