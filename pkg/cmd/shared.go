package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

const (
	globalTerraformFlagName = "tfjson"
	globalVerboseFlagName   = "verbose"
)

// globalOptions are global globalOptions same for all commands
type globalOptions struct {
	TerraformState string
	Verbose        bool
}

func persistentGlobalOptions(fs *pflag.FlagSet) (*globalOptions, error) {
	verbose, err := fs.GetBool(globalVerboseFlagName)
	if err != nil {
		return nil, err
	}

	tfjson, err := fs.GetString(globalTerraformFlagName)
	if err != nil {
		return nil, err
	}

	return &globalOptions{
		Verbose:        verbose,
		TerraformState: tfjson,
	}, nil
}

func initLogger(verbose bool) *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05 MST",
	}

	if verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	return logger
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
