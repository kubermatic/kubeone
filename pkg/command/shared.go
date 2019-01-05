package command

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func handleErrors(logger logrus.FieldLogger, action cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		err := action(ctx)
		if err != nil {
			logger.Errorln(err)
			err = cli.NewExitError("", 1)
		}

		return err
	}
}

func setupLogger(logger *logrus.Logger, action cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		if ctx.GlobalBool("verbose") {
			logger.SetLevel(logrus.DebugLevel)
		}

		return action(ctx)
	}
}

func loadClusterConfig(filename string) (*config.Cluster, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	cluster := config.Cluster{}
	if err := yaml.Unmarshal(content, &cluster); err != nil {
		return nil, fmt.Errorf("failed to decode file as JSON: %v", err)
	}

	return &cluster, nil
}

func applyTerraform(tf string, cluster *config.Cluster) error {
	if tf == "" {
		return nil
	}

	var (
		tfJSON []byte
		err    error
	)

	if tf == "-" {
		if tfJSON, err = ioutil.ReadAll(os.Stdin); err != nil {
			return fmt.Errorf("unable to load Terraform output from stdin: %v", err)
		}
	} else {
		if tfJSON, err = ioutil.ReadFile(tf); err != nil {
			return fmt.Errorf("unable to load Terraform output from file: %v", err)
		}
	}

	var tfConfig *terraform.Config
	if tfConfig, err = terraform.NewConfigFromJSON(tfJSON); err != nil {
		return fmt.Errorf("failed to parse Terraform config: %v", err)
	}

	return tfConfig.Apply(cluster)
}
