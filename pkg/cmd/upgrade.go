package cmd

import (
	"errors"
	"fmt"

	"github.com/kubermatic/kubeone/pkg/upgrader"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type upgradeOptions struct {
	globalOptions
	Manifest     string
	ForceUpgrade bool
}

func upgradeCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	uopts := &upgradeOptions{}
	cmd := &cobra.Command{
		Use:   "upgrade <manifest>",
		Short: "Upgrade Kubernetes",
		Long: `Upgrade Kubernetes

This command takes KubeOne manifest which contains information about hosts and how the cluster should be provisioned.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.`,
		Hidden:  true,
		Args:    cobra.ExactArgs(1),
		Example: `kubeone upgrade mycluster.yaml -t terraformoutput.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			logger := initLogger(gopts.Verbose)
			uopts.TerraformState = gopts.TerraformState
			uopts.Verbose = gopts.Verbose

			if len(args) != 1 {
				return errors.New("expected path to a cluster config file as an argument")
			}

			uopts.Manifest = args[0]
			if uopts.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runUpgrade(logger, uopts)
		},
	}

	cmd.Flags().BoolVarP(&uopts.ForceUpgrade, "force", "f", false, "force start upgrade process")

	return cmd
}

// runUpgrade upgrades Kubernetes on the provided machines
func runUpgrade(logger *logrus.Logger, upgradeOptions *upgradeOptions) error {
	cluster, err := loadClusterConfig(upgradeOptions.Manifest)
	if err != nil {
		return fmt.Errorf("failed to load cluster: %v", err)
	}

	options := createUpgradeOptions(upgradeOptions)

	if err = applyTerraform(upgradeOptions.TerraformState, cluster); err != nil {
		return fmt.Errorf("failed to parse terraform state: %v", err)
	}

	if err = cluster.DefaultAndValidate(); err != nil {
		return err
	}

	return upgrader.NewUpgrader(cluster, logger).Upgrade(options)
}

func createUpgradeOptions(options *upgradeOptions) *upgrader.Options {
	return &upgrader.Options{
		ForceUpgrade: options.ForceUpgrade,
		Verbose:      options.Verbose,
	}
}
