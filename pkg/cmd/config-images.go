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
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	kubeoneconfig "k8c.io/kubeone/pkg/apis/kubeone/config"
	kubeonescheme "k8c.io/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	kubeonev1beta3 "k8c.io/kubeone/pkg/apis/kubeone/v1beta3"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/templates/images"
)

type listImagesOpts struct {
	ManifestFile      string `longflag:"manifest" shortflag:"m"`
	Filter            string `longflag:"filter"`
	Provider          string `longflag:"provider"`
	KubernetesVersion string `longflag:"kubernetes-version" shortflag:"k"`
	AllImages         bool   `longflag:"all" shortflag:"a"`
}

func configImagesCmd(rootFlags *pflag.FlagSet) *cobra.Command {
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
			# To see default images list
			kubeone config images list

			# To see all images list for all Kubernetes versions
			kubeone config images list --all

			# To see base images list
			kubeone config images list --filter base

			# To see optional images list
			kubeone config images list --filter optional

			# To see images for a specific Kubernetes version
			kubeone config images list --kubernetes-version=1.26.0

			# To see images list affected by the registryConfiguration configuration (in case if any)
			kubeone config images list -m mycluster.yaml

			# To see images only related to a specific provider
			kubeone config images list --provider aws
		`),
		SilenceErrors: true,
		RunE: func(*cobra.Command, []string) error {
			manifestFile, err := rootFlags.GetString(longFlagName(opts, "ManifestFile"))
			if err != nil {
				return fail.Runtime(err, "getting ManifestFile flag")
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

	cmd.Flags().StringVar(
		&opts.KubernetesVersion,
		longFlagName(opts, "KubernetesVersion"),
		"",
		"filter images for a provided kubernetes version")

	cmd.Flags().BoolVar(
		&opts.AllImages,
		longFlagName(opts, "AllImages"),
		false,
		"list all images, including optional ones",
	)

	cmd.Flags().StringVar(
		&opts.Provider,
		longFlagName(opts, "Provider"),
		"",
		fmt.Sprintf("filter images for a specific cloud provider, one of [%s]",
			strings.Join(images.SupportedProviders(), "|")),
	)

	return cmd
}

func listImages(opts *listImagesOpts) error {
	listFilter := images.ListFilterNone

	switch opts.Filter {
	case "none":
	case "base":
		listFilter = images.ListFilterBase
	case "optional":
		listFilter = images.ListFilterOptional
	default:
		return fail.RuntimeError{
			Op:  "checking filter flag",
			Err: errors.New("--filter can be only one of [none|base|optional]"),
		}
	}

	imgResolver, err := newImageResolver(opts.KubernetesVersion, opts.ManifestFile)
	if err != nil {
		return err
	}

	// Determine the active provider: explicit flag takes priority, then
	// auto-detect from the manifest's cloudProvider field.
	provider := opts.Provider
	if provider == "" && opts.ManifestFile != "" {
		provider, err = detectProviderFromManifest(opts.ManifestFile)
		if err != nil {
			return err
		}
	}

	var images sets.Set[string]

	if opts.AllImages {
		images = sets.New(imgResolver.ListAll()...)
	} else {
		images = sets.New(imgResolver.List(listFilter)...)
	}

	if provider != "" {
		provImages, err := imgResolver.ListForProvider(provider)
		if err != nil {
			return fail.RuntimeError{Op: "listing images for provider", Err: err}
		}
		images = images.Intersection(sets.New(provImages...))
	}

	for _, img := range sets.List(images) {
		fmt.Println(img)
	}

	return nil
}

// detectProviderFromManifest reads the KubeOneCluster manifest and returns the
// cloud provider name as reported by CloudProviderSpec.Name().  Returns an
// empty string when the manifest file cannot be read or when no provider is
// configured ("none" / "unknown").
func detectProviderFromManifest(manifestFile string) (string, error) {
	configBuf, err := os.ReadFile(manifestFile)
	if err != nil {
		// manifest not accessible – silently skip provider detection
		return "", nil
	}

	apiVersion, err := kubeoneconfig.KubeOneClusterAPIVersion(configBuf)
	if err != nil {
		return "", fail.RuntimeError{Op: "parsing manifest for provider detection", Err: err}
	}

	var providerName string

	switch apiVersion {
	case kubeonev1beta2.SchemeGroupVersion.String():
		providerName, err = inspectCluster(configBuf, kubeonev1beta2.NewKubeOneCluster)
	case kubeonev1beta3.SchemeGroupVersion.String():
		providerName, err = inspectCluster(configBuf, kubeonev1beta3.NewKubeOneCluster)
	}
	if err != nil {
		return "", err
	}

	if providerName == "none" || providerName == "unknown" || providerName == "" {
		return "", nil
	}

	return providerName, nil
}

type cloudProviderNamer interface {
	CloudProviderName() string
	runtime.Object
}

// inspectCluster decodes a raw KubeOneCluster manifest into the
// versioned type T (created by newCluster), then calls inspectFn to
// extract the cloud provider name string.  T must implement runtime.Object.
func inspectCluster[T cloudProviderNamer](
	configBuf []byte,
	newCluster func() T,
) (string, error) {
	cluster := newCluster()
	if err := runtime.DecodeInto(kubeonescheme.Codecs.UniversalDecoder(), configBuf, cluster); err != nil {
		return "", fail.Config(err, fmt.Sprintf("decoding %s", cluster.GetObjectKind().GroupVersionKind()))
	}

	return cluster.CloudProviderName(), nil
}
