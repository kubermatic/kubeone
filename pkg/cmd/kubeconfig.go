package cmd

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type kubeconfigOptions struct {
	globalOptions
	Manifest string
}

// KubeconfigCommand returns the structure for declaring the "install" subcommand.
func kubeconfigCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	kopts := &kubeconfigOptions{}
	cmd := &cobra.Command{
		Use:   "kubeconfig <manifest>",
		Short: "Download the kubeconfig file from master",
		Long: `Download the kubeconfig file from master.

This command takes KubeOne manifest which contains information about hosts.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.`,
		Args:    cobra.ExactArgs(1),
		Example: `kubeone kubeconfig mycluster.yaml -t terraformoutput.json`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			kopts.TerraformState = gopts.TerraformState
			kopts.Verbose = gopts.Verbose

			kopts.Manifest = args[0]
			if kopts.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runKubeconfig(kopts)
		},
	}

	return cmd
}

// runKubeconfig downloads kubeconfig file
func runKubeconfig(kubeconfigOptions *kubeconfigOptions) error {
	if kubeconfigOptions.Manifest == "" {
		return errors.New("no cluster config file given")
	}

	cluster, err := loadClusterConfig(kubeconfigOptions.Manifest)
	if err != nil {
		return errors.Wrap(err, "failed to load cluster")
	}

	// apply terraform
	if err = applyTerraform(kubeconfigOptions.TerraformState, cluster); err != nil {
		return err
	}

	if err = cluster.DefaultAndValidate(); err != nil {
		return err
	}

	kubeconfig, err := util.DownloadKubeconfig(cluster)
	if err != nil {
		return err
	}

	fmt.Println(string(kubeconfig))

	return nil
}
