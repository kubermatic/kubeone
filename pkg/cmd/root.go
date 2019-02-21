package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the KubeOne base command

// Execute is the root command entry function
func Execute() {
	rootCmd := newRoot()

	if err := rootCmd.Execute(); err != nil {
		debug, _ := rootCmd.PersistentFlags().GetBool(globalDebugFlagName)
		if debug {
			fmt.Printf("%+v\n", err)
		} else {
			fmt.Println(err)
		}
		os.Exit(-1)
	}
}

func newRoot() *cobra.Command {
	opts := &globalOptions{}
	rootCmd := &cobra.Command{
		Use:   "kubeone",
		Short: "Kubernetes Cluster provisioning and maintaining tool",
		Long:  "Provision and maintain Kubernetes High-Availability clusters with ease",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	fs := rootCmd.PersistentFlags()

	fs.StringVarP(&opts.TerraformState, globalTerraformFlagName, "t", "", "path to terraform output JSON or - for stdin")
	fs.BoolVarP(&opts.Verbose, globalVerboseFlagName, "v", false, "verbose")
	fs.BoolVarP(&opts.Debug, globalDebugFlagName, "d", false, "debug")

	rootCmd.AddCommand(
		installCmd(fs),
		upgradeCmd(fs),
		resetCmd(fs),
		kubeconfigCmd(fs),
	)

	return rootCmd
}
