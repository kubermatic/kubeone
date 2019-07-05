/*
Copyright 2019 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kubermatic/kubeone/pkg/kubeconfig"
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

			logger := initLogger(gopts.Verbose)
			kopts.TerraformState = gopts.TerraformState
			kopts.Verbose = gopts.Verbose

			kopts.Manifest = args[0]
			if kopts.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runKubeconfig(logger, kopts)
		},
	}

	return cmd
}

// runKubeconfig downloads kubeconfig file
func runKubeconfig(logger *logrus.Logger, kubeconfigOptions *kubeconfigOptions) error {
	if kubeconfigOptions.Manifest == "" {
		return errors.New("no cluster config file given")
	}

	cluster, err := loadClusterConfig(kubeconfigOptions.Manifest, kubeconfigOptions.TerraformState, logger, "", "")
	if err != nil {
		return errors.Wrap(err, "failed to load cluster")
	}

	konfig, err := kubeconfig.Download(cluster)
	if err != nil {
		return err
	}

	fmt.Println(string(konfig))

	return nil
}
