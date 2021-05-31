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
	"os"

	"github.com/spf13/cobra"

	clusterv1alpha1 "github.com/kubermatic/machine-controller/pkg/apis/cluster/v1alpha1"

	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/client-go/kubernetes/scheme"
	apiregscheme "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/scheme"
)

// rootCmd is the KubeOne base command

// Execute is the root command entry function
func Execute() {
	// quite unlikely to happen errors here, but in case if errors present:
	// let's panic
	if err := clusterv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	if err := apiextensionsscheme.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	if err := apiregscheme.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	rootCmd := newRoot()

	if err := rootCmd.Execute(); err != nil {
		debug, _ := rootCmd.PersistentFlags().GetBool(longFlagName(&globalOptions{}, "Debug"))

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
		Use:          "kubeone",
		Short:        "Kubernetes Cluster provisioning and maintaining tool",
		Long:         "Provision and maintain Kubernetes High-Availability clusters with ease",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	fs := rootCmd.PersistentFlags()

	fs.StringVarP(&opts.ManifestFile,
		longFlagName(opts, "ManifestFile"),
		shortFlagName(opts, "ManifestFile"),
		"./kubeone.yaml",
		"Path to the KubeOne config")

	fs.StringVarP(&opts.TerraformState,
		longFlagName(opts, "TerraformState"),
		shortFlagName(opts, "TerraformState"),
		"",
		"Source for terraform output in JSON - to read from stdin. If path is a file, contents will be used. If path is a dictionary, `terraform output -json` is executed in this path")

	fs.StringVarP(&opts.CredentialsFile,
		longFlagName(opts, "CredentialsFile"),
		shortFlagName(opts, "CredentialsFile"),
		"",
		"File to source credentials and secrets from")

	fs.BoolVarP(&opts.Verbose,
		longFlagName(opts, "Verbose"),
		shortFlagName(opts, "Verbose"),
		false,
		"verbose output")

	fs.BoolVarP(&opts.Debug,
		longFlagName(opts, "Debug"),
		shortFlagName(opts, "Debug"),
		false,
		"debug output with stacktrace")

	rootCmd.AddCommand(
		installCmd(fs),
		applyCmd(fs),
		upgradeCmd(fs),
		resetCmd(fs),
		kubeconfigCmd(fs),
		configCmd(fs),
		versionCmd(),
		statusCmd(fs),
		proxyCmd(fs),
		migrateCmd(fs),
		completionCmd(rootCmd),
		documentCmd(rootCmd),
	)

	return rootCmd
}
