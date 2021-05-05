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

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	kubeonev1beta1 "k8c.io/kubeone/pkg/apis/kubeone/v1beta1"
	"k8c.io/kubeone/pkg/templates/images"

	"sigs.k8s.io/yaml"
)

type listImagesOpts struct {
	ManifestFile string `longflag:"manifest" shortflag:"m"`
	Filter       string `longflag:"filter"`
}

func imagesCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "images manipulations",
	}

	cmd.AddCommand(listImagesCmd(rootFlags))
	return cmd
}

func listImagesCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	opts := &listImagesOpts{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List images that will be used",
		Example: heredoc.Doc(`
			# To see all images list
			kubeone config images list

			# To see base images list
			kubeone config images list --filter base

			# To see optional images list
			kubeone config images list --filter optional

			# To see images list affected by the registryConfiguration configuration (in case if any)
			kubeone config images list -m mycluster.yaml
		`),
		RunE: func(*cobra.Command, []string) error {
			manifestFile, err := rootFlags.GetString(longFlagName(opts, "ManifestFile"))
			if err != nil {
				return errors.WithStack(err)
			}
			opts.ManifestFile = manifestFile

			return listImages(opts)
		},
	}

	cmd.Flags().StringVar(
		&opts.Filter,
		longFlagName(opts, "Filter"),
		"none",
		"images list filter, one of the [none|base|optional]")

	return cmd
}

func listImages(opts *listImagesOpts) error {
	listFilter := images.ListFilterNone

	switch opts.Filter {
	case "none":
	case "base":
		listFilter = images.ListFilterBase
	case "optional":
		listFilter = images.ListFilterOpional
	default:
		return fmt.Errorf("--filter can be only one of [none|base|optional]")
	}

	var resolveropts []images.Opt

	// FOR FUTURE READER: we only attempt to read the ManifestFile, but if it's not there, we don't care.
	configBuf, err := os.ReadFile(opts.ManifestFile)
	if err == nil {
		// Custom loading of the config is needed to avoid "normal" validation process, but we here don't care about
		// validity of the config, the only part that's needed is `.RegistryConfiguration`
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
		resolveropts = append(resolveropts, overRegGetter)
	}

	imgResolver := images.NewResolver(resolveropts...)
	for _, img := range imgResolver.List(listFilter) {
		fmt.Println(img)
	}

	return nil
}
