package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/kubermatic/kubeone/pkg/installer"
	"github.com/kubermatic/kubeone/pkg/manifest"
)

// InstallCommand wapper for logger
func InstallCommand(logger *logrus.Logger) cli.Command {
	return cli.Command{
		Name:   "install",
		Usage:  "Installs Kubernetes onto pre-existing machines",
		Action: InstallAction(logger),
		Flags: []cli.Flag{
			cli.StringFlag{
				EnvVar: "MANIFEST_FILE",
				Name:   "manifest, m",
				Usage:  "path to the kubeone manifest",
				Value:  "manifest.yaml",
			},
			cli.StringFlag{
				EnvVar: "TF_OUTPUT",
				Name:   "tfjson, t",
				Usage:  "path to terraform output JSON",
				Value:  "",
			},
		},
	}
}

// InstallAction wrapper for logger
func InstallAction(logger *logrus.Logger) cli.ActionFunc {
	return handleErrors(logger, setupLogger(logger, func(ctx *cli.Context) error {
		manifestFile := ctx.String("manifest")
		if manifestFile == "" {
			return errors.New("no manifest file given")
		}

		manifest, err := loadManifest(manifestFile)
		if err != nil {
			return fmt.Errorf("failed to load manifest: %v", err)
		}

		if tf := ctx.String("tfjson"); tf != "" {
			tfjson, err1 := ioutil.ReadFile(tf)
			if err1 != nil {
				return fmt.Errorf("unable to load tfjson: %v", err1)
			}

			if err2 := manifest.Merge(tfjson); err != nil {
				return fmt.Errorf("tfjson is invalid %v", err2)
			}
		}

		if err = manifest.Validate(); err != nil {
			return fmt.Errorf("manifest is invalid: %v", err)
		}

		worker := installer.NewInstaller(manifest, logger)
		_, err = worker.Run()

		return err
	}))
}

func loadManifest(filename string) (*manifest.Manifest, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	manifest := manifest.Manifest{}
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("failed to decode file as JSON: %v", err)
	}

	return &manifest, nil
}
