package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// options are global options same for all commands
type options struct {
	TerraformState string
	Verbose        bool
}

var o = &options{}

// rootCmd is the KubeOne base command
var rootCmd = &cobra.Command{
	Use:   "kubeone",
	Short: "Kubernetes Cluster provisioning and maintaining tool",
	Long:  "Provision and maintain Kubernetes High-Availability clusters with ease",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute is the root command entry function
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// init adds subcommands and setups flags
func init() {
	rootCmd.PersistentFlags().StringVarP(&o.TerraformState, "tfjson", "t", "", "path to terraform output JSON or - for stdin")
	rootCmd.PersistentFlags().BoolVarP(&o.Verbose, "verbose", "v", false, "verbose")

	addCommands()
}

// addCommands add subcommands to the root tree
func addCommands() {
	rootCmd.AddCommand(installCmd())
	rootCmd.AddCommand(resetCmd())
	rootCmd.AddCommand(kubeconfigCmd())
}
