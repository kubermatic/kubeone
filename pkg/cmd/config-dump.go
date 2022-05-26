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
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8c.io/kubeone/pkg/apis/kubeone/config"
	kubeonescheme "k8c.io/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1beta1 "k8c.io/kubeone/pkg/apis/kubeone/v1beta1"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/templates"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

type configDumpOpts struct {
	globalOptions
}

func configDumpCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &configDumpOpts{}

	cmd := &cobra.Command{
		Use:           "dump",
		Short:         "Merge the KubeOneCluster manifest with the Terraform state and dump it to the stdout",
		SilenceErrors: true,
		Example:       `kubeone config dump -m kubeone.yaml -t tf.json`,
		RunE: func(*cobra.Command, []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			opts.globalOptions = *gopts

			return dumpConfig(opts)
		},
	}

	return cmd
}

func dumpConfig(opts *configDumpOpts) error {
	logger := newLogger(opts.Verbose, opts.LogFormat)

	// Read the TypeMeta. We need the API version, so we can convert the
	// internal representation to the original API version.
	// NB: We can't always convert to the latest API version because we might
	// lose information (e.g. the AssetConfiguration API has been removed in
	// the v1beta2 API).
	manifest, err := os.ReadFile(opts.ManifestFile)
	if err != nil {
		return fail.Runtime(err, "reading KubeOneCluster manifest file")
	}

	typeMeta := runtime.TypeMeta{}
	if err = yaml.Unmarshal(manifest, &typeMeta); err != nil {
		return fail.Runtime(err, "unmarshaling cluster typeMeta")
	}

	// Load the KubeOneCluster manifest.
	// This merges the provided manifest with the Terraform output, defaults
	// the merged manifest, converts it to the internal representations, and
	// then validates it.
	cluster, err := config.LoadKubeOneCluster(opts.ManifestFile, opts.TerraformState, opts.CredentialsFile, logger)
	if err != nil {
		return err
	}

	// Convert the internal KubeOneCluster manifest to the versioned manifest.
	// NB: validation works only on the internal representation, so if we want
	// to validate the merged manifest, we can't avoid this step.
	var objs []runtime.Object
	switch typeMeta.APIVersion {
	case kubeonev1beta1.SchemeGroupVersion.String():
		versionedCluster := kubeonev1beta1.NewKubeOneCluster()
		if cErr := kubeonescheme.Scheme.Convert(cluster, versionedCluster, nil); cErr != nil {
			return fail.Config(cErr, fmt.Sprintf("converting %s to internal object", versionedCluster.GroupVersionKind()))
		}

		objs = append(objs, versionedCluster)
	case kubeonev1beta2.SchemeGroupVersion.String():
		versionedCluster := kubeonev1beta2.NewKubeOneCluster()
		if cErr := kubeonescheme.Scheme.Convert(cluster, versionedCluster, nil); cErr != nil {
			return fail.Config(cErr, fmt.Sprintf("converting %s to internal object", versionedCluster.GroupVersionKind()))
		}

		objs = append(objs, versionedCluster)
	default:
		return fail.ConfigError{
			Op:  "checking KubeOneCluster apiVersion",
			Err: errors.New("invalid"),
		}
	}

	// Convert the KubeOneCluster struct to the YAML representation
	clusterYAML, err := templates.KubernetesToYAML(objs)
	if err != nil {
		return err
	}

	fmt.Println(clusterYAML)

	return nil
}
