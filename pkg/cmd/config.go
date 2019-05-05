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
	"bytes"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v2"

	kubeonev1alpha1 "github.com/kubermatic/kubeone/pkg/apis/kubeone/v1alpha1"
	"github.com/kubermatic/kubeone/pkg/config"

	kyaml "sigs.k8s.io/yaml"
)

type migrateOptions struct {
	globalOptions
	Manifest string
}

// configCmd setups the config command
func configCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Commands for working with the KubeOneCluster configuration manifests",
	}

	cmd.AddCommand(migrateCmd(rootFlags))

	return cmd
}

// migrateCmd setups the migrate command
func migrateCmd(_ *pflag.FlagSet) *cobra.Command {
	mOpts := &migrateOptions{}
	cmd := &cobra.Command{
		Use:   "migrate <cluster-manifest>",
		Short: "Migrate the pre-v0.6.0 configuration manifest to the KubeOneCluster manifest",
		Long: `
Migrate the pre-v0.6.0 KubeOne configuration manifest to the KubeOneCluster manifest used as of v0.6.0.
The new manifest is printed on the standard output.
`,
		Args:    cobra.ExactArgs(1),
		Example: `kubeone migrate mycluster.yaml`,
		RunE: func(_ *cobra.Command, args []string) error {
			mOpts.Manifest = args[0]
			if mOpts.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runMigrate(mOpts)
		},
	}

	return cmd
}

// runMigrate migrates the pre-v0.6.0 KubeOne API manifest to the KubeOneCluster manifest used as of v0.6.0
func runMigrate(migrateOptions *migrateOptions) error {
	// Convert old config yaml to new config yaml
	newConfigYAML, err := config.MigrateToKubeOneClusterAPI(migrateOptions.Manifest)
	if err != nil {
		return errors.Wrap(err, "unable to migrate the provided configuration")
	}

	// Validate new config by unmarshaling
	var buffer bytes.Buffer
	err = yaml.NewEncoder(&buffer).Encode(newConfigYAML)
	if err != nil {
		return errors.Wrap(err, "failed to encode new config as YAML")
	}

	newConfig := &kubeonev1alpha1.KubeOneCluster{}
	err = kyaml.UnmarshalStrict(buffer.Bytes(), &newConfig)
	if err != nil {
		return errors.Wrap(err, "failed to decode new config")
	}

	// Print new config yaml
	err = yaml.NewEncoder(os.Stdout).Encode(newConfigYAML)
	if err != nil {
		return errors.Wrap(err, "failed to encode new config as YAML")
	}

	return nil
}
