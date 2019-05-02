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
	"fmt"
	"strings"

	kyaml "github.com/ghodss/yaml"
	"github.com/kubermatic/kubeone/pkg/apis/kubeone/v1alpha1"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v2"
)

type migrateOptions struct {
	globalOptions
	Manifest string
}

// migrateCmd setups the migrate command
func migrateCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	mOpts := &migrateOptions{}
	cmd := &cobra.Command{
		Use:   "migrate <old-manifest>",
		Short: "Migrate the pre-v0.6.0 configuration manifest to the KubeOneCluster manifest",
		Long: `
Migrates the pre-v0.6.0 KubeOne API manifest to the KubeOneCluster manifest used as of v0.6.0.
The new manifest is printed to the standard output. The migrate command is unable to parse the Terraform output.
`,
		Args:    cobra.ExactArgs(1),
		Example: `kubeone migrate mycluster.yaml`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return errors.Wrap(err, "unable to get global flags")
			}

			logger := initLogger(gopts.Verbose)

			mOpts.Manifest = args[0]
			if mOpts.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runMigrate(logger, mOpts)
		},
	}

	return cmd
}

// runMigrate migrates the pre-v0.6.0 KubeOne API manifest to the KubeOneCluster manifest used as of v0.6.0.
func runMigrate(logger *logrus.Logger, migrateOptions *migrateOptions) error {
	newConfig, err := config.MigrateToKubeOneClusterAPI(migrateOptions.Manifest)
	if err != nil {
		return errors.Wrap(err, "unable to migrate the provided configuration")
	}

	var buffer bytes.Buffer

	err = yaml.NewEncoder(&buffer).Encode(newConfig)
	if err != nil {
		return errors.Wrap(err, "failed to encode new config as YAML")
	}

	config := v1alpha1.KubeOneCluster{}
	err = kyaml.Unmarshal(buffer.Bytes(), &config)
	if err != nil {
		return errors.Wrap(err, "failed to decode new config as YAML")
	}

	cfg := []interface{}{
		config,
	}

	configYaml, err := kubernetesToYAML(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to convert the new configuration to yaml")
	}

	fmt.Println(configYaml)

	return nil
}

// kubernetesToYAML properly encodes a list of resources as YAML. Straight up encoding as YAML leaves us with a
// non-standard data structure. Going through JSON eliminates the extra fields and keys and results in what you
// would expect to see. This function takes a slice of items to support creating a multi-document YAML string
// (separated with "---" between each item).
func kubernetesToYAML(data []interface{}) (string, error) {
	var buffer bytes.Buffer

	for _, item := range data {
		var (
			encodedItem []byte
			err         error
		)

		if str, ok := item.(string); ok {
			encodedItem = []byte(strings.TrimSpace(str))
		} else {
			encodedItem, err = kyaml.Marshal(item)
		}

		if err != nil {
			return "", fmt.Errorf("failed to marshal item: %v", err)
		}
		if _, err := buffer.Write(encodedItem); err != nil {
			return "", fmt.Errorf("failed to write into buffer: %v", err)
		}
		if _, err := buffer.WriteString("\n---\n"); err != nil {
			return "", fmt.Errorf("failed to write into buffer: %v", err)
		}
	}

	return buffer.String(), nil
}
