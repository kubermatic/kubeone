package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/kubermatic/kubeone/pkg/installer"
	"github.com/kubermatic/kubeone/pkg/manifest"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func InstallCommand(logger *logrus.Logger) cli.Command {
	return cli.Command{
		Name:      "install",
		Usage:     "Installs Kubernetes onto pre-existing machines",
		Action:    InstallAction(logger),
		ArgsUsage: "MANIFEST_FILE",
	}
}

func InstallAction(logger *logrus.Logger) cli.ActionFunc {
	return handleErrors(logger, setupLogger(logger, func(ctx *cli.Context) error {
		manifestFile := ctx.Args().First()
		if len(manifestFile) == 0 {
			return errors.New("no manifest file given")
		}

		manifest, err := loadManifest(manifestFile)
		if err != nil {
			return fmt.Errorf("failed to load manifest: %v", err)
		}

		err = manifest.Validate()
		if err != nil {
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
