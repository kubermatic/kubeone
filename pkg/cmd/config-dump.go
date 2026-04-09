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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	kubeonescheme "k8c.io/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/tasks"
	"k8c.io/kubeone/pkg/templates"

	"k8s.io/apimachinery/pkg/runtime"
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
	st, err := opts.BuildState()
	if err != nil {
		return err
	}

	if err = tasks.WithFindControlPlane(nil).Run(st); err != nil {
		return err
	}

	v1beta2Cluster := kubeonev1beta2.NewKubeOneCluster()
	if err = kubeonescheme.Scheme.Convert(st.Cluster, v1beta2Cluster, nil); err != nil {
		return fail.Config(err, fmt.Sprintf("converting %s to internal object", v1beta2Cluster.GroupVersionKind()))
	}

	// Convert the KubeOneCluster struct to the YAML representation
	clusterYAML, err := templates.KubernetesToYAML([]runtime.Object{v1beta2Cluster})
	if err != nil {
		return err
	}

	fmt.Println(clusterYAML)

	return nil
}
