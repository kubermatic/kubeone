package command

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kubermatic/kubeone/pkg/installer"
	"github.com/kubermatic/kubeone/pkg/manifest"
	"github.com/kubermatic/kubeone/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
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
				Value:  "manifest.json",
			},
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
		manifestFile := ctx.String("manifest")
		if manifestFile == "" {
			return errors.New("no manifest file given")
		}

		manifest, err := loadManifest(manifestFile)
		if err != nil {
			return fmt.Errorf("failed to load manifest: %v", err)
		}

		if tf := ctx.String("tfjson"); tf != "" {
			var tfJSON []byte
			if tf == "-" {
				if tfJSON, err = ioutil.ReadAll(os.Stdin); err != nil {
					return fmt.Errorf("unable to load terraform output from stdin: %v", err)
				}
			} else {
				if tfJSON, err = ioutil.ReadFile(tf); err != nil {
					return fmt.Errorf("unable to load terraform output from file: %v", err)
				}
			}

			var tfConfig *terraform.Config
			if tfConfig, err = terraform.NewConfigFromJSON(tfJSON); err != nil {
				return fmt.Errorf("failed to parse terraform config: %v", err)
			}
			tfConfig.Apply(manifest)
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
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("failed to decode file as JSON: %v", err)
	}

	return &manifest, nil
}
