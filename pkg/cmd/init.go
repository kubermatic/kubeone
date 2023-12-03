/*
Copyright 2022 The KubeOne Authors.

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
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"k8c.io/kubeone/pkg/cmd/initcmd"
	"k8c.io/kubeone/pkg/fail"

	"k8s.io/apimachinery/pkg/util/sets"
)

type initOpts struct {
	Interactive       bool      `longflag:"interactive" shortflag:"i"`
	Provider          oneOfFlag `longflag:"provider"`
	ClusterName       string    `longflag:"cluster-name"`
	KubernetesVersion string    `longflag:"kubernetes-version"`
	Terraform         bool      `longflag:"terraform"`
	Path              string    `longflag:"path"`
}

func initCmd() *cobra.Command {
	validProviders := []string{}
	for k := range initcmd.ValidProviders {
		validProviders = append(validProviders, k)
	}

	opts := &initOpts{
		Provider: oneOfFlag{
			validSet:     sets.New(validProviders...),
			defaultValue: "none",
		},
	}

	clusterNameFlag := longFlagName(opts, "ClusterName")
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init new kubeone cluster configuration",
		Long: heredoc.Doc(`
			Initialize new KubeOne + terraform configuration for chosen provider.
		`),
		SilenceErrors: true,
		Example:       `kubeone init --provider aws`,
		RunE: func(_ *cobra.Command, args []string) error {
			if opts.KubernetesVersion == "" {
				return fail.Runtime(fmt.Errorf("--kubernetes-version is a required flag"), "flag validation")
			}

			return runInit(opts)
		},
	}

	providerUsageText := fmt.Sprintf("provider to initialize, possible values: %s", strings.Join(opts.Provider.PossibleValues(), ", "))

	cmd.Flags().BoolVarP(&opts.Interactive, longFlagName(opts, "Interactive"), shortFlagName(opts, "Interactive"), false, "run command in the interactive mode")
	cmd.Flags().BoolVar(&opts.Terraform, longFlagName(opts, "Terraform"), true, "generate terraform config")
	cmd.Flags().StringVar(&opts.ClusterName, clusterNameFlag, "", "name of the cluster")
	cmd.Flags().StringVar(&opts.KubernetesVersion, longFlagName(opts, "KubernetesVersion"), defaultKubeVersion, "kubernetes version")
	cmd.Flags().StringVar(&opts.Path, longFlagName(opts, "Path"), ".", "path where to write files")
	cmd.Flags().Var(&opts.Provider, longFlagName(opts, "Provider"), providerUsageText)

	return cmd
}

func runInit(opts *initOpts) error {
	if opts.Interactive {
		return runInitInteractive()
	}

	iOpts := initcmd.NewGenerateOpts(opts.Path, opts.Provider.String(), opts.ClusterName, opts.KubernetesVersion, opts.Terraform)

	return initcmd.GenerateConfigs(iOpts)
}

func runInitInteractive() error {
	opts, err := initcmd.InitInteractive(defaultKubeVersion)
	if err != nil {
		return err
	}

	return initcmd.GenerateConfigs(opts)
}
