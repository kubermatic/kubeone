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
		ArgsUsage: "MANIFEST_FILE",
		Action:    ResetAction(logger),
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

// ResetAction handles the "reset" subcommand.
func ResetAction(logger *logrus.Logger) cli.ActionFunc {
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
		_, err = worker.Reset(ctx.GlobalBool("verbose"))

		return err
	}))
}
