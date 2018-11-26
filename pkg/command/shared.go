package command

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
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

	if err = tfConfig.Validate(); err != nil {
		return fmt.Errorf("Terraform output is invalid: %v", err)
	}

	tfConfig.Apply(cluster)

	return nil
}

func loadMachineControllerCredentials(p config.ProviderName) (data map[string]string, err error) {
	data = map[string]string{}

	switch p {
	case config.ProviderNameAWS:
		if data["AWS_ACCESS_KEY_ID"], err = getEnvVar("AWS_ACCESS_KEY_ID"); err != nil {
			return
		}
		if data["AWS_SECRET_ACCESS_KEY"], err = getEnvVar("AWS_SECRET_ACCESS_KEY"); err != nil {
			return
		}

	case config.ProviderNameOpenStack:
		if data["OS_AUTH_URL"], err = getEnvVar("OS_AUTH_URL"); err != nil {
			return
		}
		if data["OS_USER_NAME"], err = getEnvVar("OS_USER_NAME"); err != nil {
			return
		}
		if data["OS_PASSWORD"], err = getEnvVar("OS_PASSWORD"); err != nil {
			return
		}
		if data["OS_DOMAIN_NAME"], err = getEnvVar("OS_DOMAIN_NAME"); err != nil {
			return
		}
		if data["OS_TENANT_NAME"], err = getEnvVar("OS_TENANT_NAME"); err != nil {
			return
		}

	case config.ProviderNameHetzner:
		if data["HZ_TOKEN"], err = getEnvVar("HZ_TOKEN"); err != nil {
			return
		}

	case config.ProviderNameDigitalOcean:
		if data["DO_TOKEN"], err = getEnvVar("DO_TOKEN"); err != nil {
			return
		}

	case config.ProviderNameVSphere:
		if data["VSPHERE_ADDRESS"], err = getEnvVar("VSPHERE_ADDRESS"); err != nil {
			return
		}
		if data["VSPHERE_USERNAME"], err = getEnvVar("VSPHERE_USERNAME"); err != nil {
			return
		}
		if data["VSPHERE_PASSWORD"], err = getEnvVar("VSPHERE_PASSWORD"); err != nil {
			return
		}

	default:
		err = fmt.Errorf("missing cloud provider credentials")
	}

	return
}

func getEnvVar(ev string) (data string, err error) {
	data = os.Getenv(ev)
	if data == "" {
		err = fmt.Errorf("%q environment variable is not set", ev)
	}
	return
}
