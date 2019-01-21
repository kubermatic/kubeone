package cmd

import (
	"errors"
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type resetOptions struct {
	options
	Manifest       string
	DestroyWorkers bool
}

// resetCmd setups reset command
func resetCmd() *cobra.Command {
	var ro = &resetOptions{}
	var resetCmd = &cobra.Command{
		Use:   "reset <manifest>",
		Short: "Revert changes",
		Long: `Undo all changes done by KubeOne to the configured machines.

This command takes KubeOne manifest which contains information about hosts.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := initLogger()
			ro.TerraformState = o.TerraformState
			ro.Verbose = o.Verbose

			if len(args) != 1 {
				return errors.New("expected path to a cluster config file as an argument")
			}

			ro.Manifest = args[0]
			if ro.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runReset(logger, ro)
		},
	}

	resetCmd.Flags().BoolVarP(&ro.DestroyWorkers, "destroy-workers", "", false, "destroy all worker machines before resetting cluster")

	return resetCmd
}

// runReset resets all machines provisioned by KubeOne
func runReset(logger *logrus.Logger, resetOptions *resetOptions) error {
	if resetOptions.Manifest == "" {
		return errors.New("no cluster config file given")
	}

	cluster, err := loadClusterConfig(resetOptions.Manifest)
	if err != nil {
		return fmt.Errorf("failed to load cluster: %v", err)
	}

	if err = applyTerraform(resetOptions.TerraformState, cluster); err != nil {
		return err
	}

	if err = cluster.DefaultAndValidate(); err != nil {
		return err
	}

	options := &installer.Options{
		Verbose:        resetOptions.Verbose,
		DestroyWorkers: resetOptions.DestroyWorkers,
	}

	return installer.NewInstaller(cluster, logger).Reset(options)
}
