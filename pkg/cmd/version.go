package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var versionString = "development"

// versionCmd setups version command
func versionCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display KubeOne version",
		Long:  `Prints the exact version number, as embedded by the build system.`,
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, args []string) error {
			fmt.Printf("KubeOne %s\n", versionString)
			return nil
		},
	}

	return cmd
}
