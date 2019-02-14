package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func upgradeCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade <manifest>",
		Short: "Upgrade Kubernetes",
		Long: `Upgrade Kubernetes

This command takes KubeOne manifest which contains information about hosts and how the cluster should be provisioned.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.`,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("not implemented yet")
		},
	}
	return cmd
}
