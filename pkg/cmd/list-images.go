/*
Copyright 2021 The KubeOne Authors.

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
	"sigs.k8s.io/yaml"

	kubeonev1beta1 "k8c.io/kubeone/pkg/apis/kubeone/v1beta1"
	"k8c.io/kubeone/pkg/templates/images"
)

type listImagesOpts struct {
	ManifestFile string `longflag:"manifest" shortflag:"m"`
}

func listImagesCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &listImagesOpts{}

	cmd := &cobra.Command{
		Use:     "list-images",
		Short:   "List images that will be used",
		Example: `kubeone list-images -m mycluster.yaml -t terraformoutput.json`,
		RunE: func(*cobra.Command, []string) error {
			manifestFile, err := rootFlags.GetString(longFlagName(opts, "ManifestFile"))
			if err != nil {
				return errors.WithStack(err)
			}
			opts.ManifestFile = manifestFile

			return listImages(opts)
		},
	}

	return cmd
}

func listImages(opts *listImagesOpts) error {
	var imgopts []images.Opt

	configBuf, err := os.ReadFile(opts.ManifestFile)
	if err == nil {
		var conf kubeonev1beta1.KubeOneCluster
		if err = yaml.Unmarshal(configBuf, &conf); err != nil {
			return err
		}

		overRegGetter := images.WithOverwriteRegistryGetter(func() string {
			if rc := conf.RegistryConfiguration; rc != nil {
				return rc.OverwriteRegistry
			}
			return ""
		})
		imgopts = append(imgopts, overRegGetter)
	}

	imgResolver := images.NewResolver(imgopts...)
	for _, img := range imgResolver.ListAll() {
		fmt.Println(img)
	}

	// * render addons, extract images from Deployments/STS
	return nil
}
