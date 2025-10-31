/*
Copyright 2025 The KubeOne Authors.

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
	"context"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	kubeonevalidation "k8c.io/kubeone/pkg/apis/kubeone/validation"
	"k8c.io/kubeone/pkg/images"
	"k8c.io/kubeone/pkg/semverutil"
)

type mirrorImagesOpts struct {
	globalOptions
	Filter             string `longflag:"filter"`
	KubernetesVersions string `longflag:"kubernetes-versions" shortflag:"k"`
	Insecure           bool   `longflag:"insecure"`
	DryRun             bool   `longflag:"dry-run"`
	Registry           string
}

func mirrorImagesCmd(*pflag.FlagSet) *cobra.Command {
	opts := &mirrorImagesOpts{}

	cmd := &cobra.Command{
		Use:   "mirror-images [registry]",
		Short: "Mirror images to another registry",
		Long: heredoc.Doc(`
            Mirror images used by KubeOne to another registry.
            This command lists the images (including control-plane images) and copies them to the specified target registry.
        `),
		Example: heredoc.Doc(`
            # Mirror all images to a target registry
            kubeone mirror-images myregistry.com

            # Mirror images for a specific Kubernetes versions
            kubeone mirror-images --kubernetes-versions=v1.26.0,v1.29.0 --filter=control-plane myregistry.com
        `),
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, args []string) error {
			var err error
			logger := newLogger(opts.Verbose, opts.LogFormat)
			if len(args) != 1 {
				return fmt.Errorf("error: registry is required. Usage: kubeone mirror-images [registry]")
			}
			opts.Registry = args[0]

			// Determine the list of Kubernetes versions to process
			var versions []string
			if opts.KubernetesVersions == "" {
				// Call the default-versions() function if no version is provided
				versions, err = defaultVersions(logger)
				if err != nil {
					return fmt.Errorf("failed to get default Kubernetes versions: %w", err)
				}
			} else {
				// Split the provided versions by commas
				if versions, err = validatedVersions(logger, opts.KubernetesVersions); err != nil {
					return err
				}
			}

			return mirrorImages(logger, opts, versions)
		},
	}

	cmd.Flags().StringVar(
		&opts.Filter,
		longFlagName(opts, "Filter"),
		"none",
		"images list filter, one of the [none|base|optional|control-plane]")

	cmd.Flags().StringVar(
		&opts.KubernetesVersions,
		longFlagName(opts, "KubernetesVersions"),
		"",
		"Kubernetes versions (comma-separated, format: vX.Y[.Z])")

	cmd.Flags().BoolVar(
		&opts.DryRun,
		longFlagName(opts, "DryRun"),
		false,
		"Only print the names of source and destination images",
	)

	cmd.Flags().BoolVar(
		&opts.Insecure,
		longFlagName(opts, "Insecure"),
		false,
		"insecure option to bypass TLS certificate verification",
	)

	return cmd
}

func validatedVersions(logger *logrus.Logger, kubeversions string) ([]string, error) {
	versions := strings.Split(kubeversions, ",")
	for i, version := range versions {
		ver, err := semver.NewVersion(version)
		if err != nil {
			return nil, fmt.Errorf("invalid Kubernetes version: %w", err)
		}

		versions[i] = ver.String()
		logger.Info(fmt.Sprintf("🚢 Extracted the desired Kubernetes Version: %s", version))
	}

	return versions, nil
}

func defaultVersions(logger *logrus.Logger) ([]string, error) {
	var allVersions []string

	// Get the range of minor versions between the provided min and max.
	minorVersions, err := semverutil.GetMinorRange(kubeonevalidation.MinimumSupportedVersion, kubeonevalidation.MaximumSupportedVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve minor version range: %w", err)
	}
	for _, minor := range minorVersions {
		// Get the latest patch for the given minor version.
		latestPatch, err := images.FetchLatestPatchForKubernetesVersion(minor)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch latest patch for version %s: %w", minor, err)
		}

		patchRange, err := semverutil.GetPatchRange(fmt.Sprintf("%s.0", minor), latestPatch)
		if err != nil {
			return nil, fmt.Errorf("failed to get patch range from for %s: %w", minor, err)
		}

		logger.Infof("🚢 Extracted patches for Kubernetes version %s", minor)
		allVersions = append(allVersions, patchRange...)
	}

	return allVersions, nil
}

func mirrorImages(logger *logrus.Logger, opts *mirrorImagesOpts, versions []string) error {
	var verb string
	var count, fullCount int

	ctx := context.Background()

	logger.Info("🚀 Collecting images used by kubeone ...")

	imageList, err := images.GetKubeoneImages(ctx, opts.Filter, versions)
	if err != nil {
		return fmt.Errorf("failed to get KubeOne images: %w", err)
	}

	logger.WithField("registry", opts.Registry).Info("📦 Mirroring images…")
	count, fullCount, err = images.CopyImages(ctx, logger, opts.DryRun, opts.Insecure, imageList, opts.Registry, "kubeone")
	if err != nil {
		return fmt.Errorf("failed to mirror all images (successfully copied %d/%d): %w", count, fullCount, err)
	}

	verb = "mirroring"
	if opts.DryRun {
		verb = "mirroring (dry-run)"
	}

	logger.WithFields(logrus.Fields{"copied-image-count": count, "all-image-count": fullCount}).Info(fmt.Sprintf("✅ Finished %s images.", verb))

	return nil
}
