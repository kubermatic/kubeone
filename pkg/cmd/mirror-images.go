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
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	kubeonevalidation "k8c.io/kubeone/pkg/apis/kubeone/validation"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/semverutil"
	"k8c.io/kubeone/pkg/templates/images"
)

// kubeadmConstantsTemplate is the template URL for fetching Kubernetes component version constants
const kubeadmConstantsTemplate = "https://raw.githubusercontent.com/kubernetes/kubernetes/%s/cmd/kubeadm/app/constants/constants.go"

// stableVersionURL is the template URL for fetching the latest patch for a specific Kubernetes version.
const stableVersionURL = "https://dl.k8s.io/release/stable-%s.txt"

// component names used to identify Kubernetes system components.
const (
	componentEtcd              = "etcd"                    // etcd key-value store
	componentAPIServer         = "kube-apiserver"          // Kubernetes API server
	componentControllerManager = "kube-controller-manager" // Controller manager component
	componentScheduler         = "kube-scheduler"          // Scheduler component
	componentKubeProxy         = "kube-proxy"              // Network proxy component
	componentCoreDNS           = "coredns"                 // CoreDNS DNS server
	componentPause             = "pause"                   // Pause container
)

// versionConstant names from Kubernetes source that store version information.
const (
	versionConstPause   = "PauseVersion"       // Constant name for pause container version
	versionConstEtcd    = "DefaultEtcdVersion" // Constant name for etcd version
	versionConstCoreDNS = "CoreDNSVersion"     // Constant name for CoreDNS version
)

// kubernetesRegistry is the default container registry for Kubernetes components.
const kubernetesRegistry = "registry.k8s.io"

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
		latestPatch, err := fetchLatestPatchForKubernetesVersion(minor)
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

// fetchLatestPatchForKubernetesVersion fetches the latest patch for a specific Kubernetes version.
func fetchLatestPatchForKubernetesVersion(version string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(stableVersionURL, version), nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if the response status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func mirrorImages(logger *logrus.Logger, opts *mirrorImagesOpts, versions []string) error {
	var err error
	var cpImages []string
	ctx := context.Background()
	controlPlaneOnly := false
	listFilter := images.ListFilterNone
	switch opts.Filter {
	case "none":
	case "control-plane":
		controlPlaneOnly = true
	case "base":
		listFilter = images.ListFilterBase
	case "optional":
		listFilter = images.ListFilterOpional
	default:
		return fail.RuntimeError{
			Op:  "checking filter flag",
			Err: errors.New("--filter can be only one of [none|base|optional|control-plane]"),
		}
	}

	logger.Info("🚀 Collecting images used by kubeone ...")

	imageSet := sets.New[string]()
	for _, ver := range versions {
		version := fmt.Sprintf("v%s", ver)
		cpImages, err = getControlPlaneImages(ctx, version)
		if err != nil {
			return err
		}

		imageSet.Insert(cpImages...)
		if !controlPlaneOnly {
			resolver, resolveErr := newImageResolver(version, "")
			if resolveErr != nil {
				return resolveErr
			}
			images := resolver.List(listFilter)
			imageSet.Insert(images...)
		}
	}

	var verb string
	var count, fullCount int
	logger.WithField("registry", opts.Registry).Info("📦 Mirroring images…")
	count, fullCount, err = CopyImages(ctx, logger, opts.DryRun, opts.Insecure, sets.List(imageSet), opts.Registry, "kubeone")
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

// getControlPlaneImages returns a list of container images used on control plane node.
func getControlPlaneImages(ctx context.Context, version string) ([]string, error) {
	images := make([]string, 0)

	// start with core kubernetes images
	for _, component := range []string{componentAPIServer, componentControllerManager, componentScheduler, componentKubeProxy} {
		images = append(images, getGenericImage(kubernetesRegistry, component, version))
	}

	for _, component := range []string{componentEtcd, componentCoreDNS, componentPause} {
		img, err := getComponentImage(ctx, component, version)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, nil
}

func getGenericImage(prefix, image, tag string) string {
	return fmt.Sprintf("%s/%s:%s", prefix, image, tag)
}

func getComponentImage(ctx context.Context, component, version string) (string, error) {
	var target string
	registry := kubernetesRegistry
	switch component {
	case componentEtcd:
		target = versionConstEtcd
	case componentCoreDNS:
		registry += "/coredns"
		target = versionConstCoreDNS
	case componentPause:
		target = versionConstPause
	}

	tag, err := getConstantValue(ctx, version, target)
	if err != nil {
		return "", err
	}

	return getGenericImage(registry, component, tag), nil
}

func getConstantValue(ctx context.Context, version, constant string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(kubeadmConstantsTemplate, version), nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, constant+" =") {
			return strings.Trim(strings.SplitN(line, "=", 2)[1], `" `), nil
		}
	}

	return "", fmt.Errorf("cannot find the value for %s", constant)
}

func CopyImages(ctx context.Context, log logrus.FieldLogger, dryRun, insecure bool, images []string, registry, userAgent string) (int, int, error) {
	var failedImages []string
	for i, source := range images {
		dest, err := retagImage(log, source, registry)
		if err != nil {
			return 0, len(images), fmt.Errorf("failed to prepare image: %w", err)
		}

		log := log.WithFields(logrus.Fields{
			"count": fmt.Sprintf("%d/%d", i+1, len(images)),
			"image": source,
		})

		if dryRun {
			log.Info("Dry run:")

			continue
		}

		if err := copyWithRetry(ctx, log, source, dest, userAgent, insecure); err != nil {
			log.Errorf("Failed to copy image: %v", err)
			failedImages = append(failedImages, fmt.Sprintf("  - %s", source))
		}
	}

	copied := len(images) - len(failedImages)
	if len(failedImages) > 0 {
		return copied, len(images), fmt.Errorf("failed images:\n%s", strings.Join(failedImages, "\n"))
	}

	return copied, len(images), nil
}

func retagImage(log logrus.FieldLogger, source, registry string) (string, error) {
	ref, err := name.ParseReference(source)
	if err != nil {
		return "", fmt.Errorf("invalid image reference: %w", err)
	}

	dest := fmt.Sprintf("%s/%s:%s", registry, ref.Context().RepositoryStr(), ref.Identifier())
	log.WithField("target", dest).Debug("Image retagged")

	return dest, nil
}

func copyWithRetry(ctx context.Context, log logrus.FieldLogger, src, dst, userAgent string, insecure bool) error {
	opts := []crane.Option{
		crane.WithContext(ctx),
		crane.WithUserAgent(userAgent),
	}

	if insecure {
		opts = append(opts, crane.Insecure)
	}

	retryPolicy := wait.Backoff{
		Steps:    5,
		Duration: 1 * time.Second,
		Factor:   1.5,
		Jitter:   0.1,
	}

	return retry.OnError(retryPolicy, func(error) bool { return true }, func() error {
		log.Info("Copying image...")
		err := crane.Copy(src, dst, opts...)
		if err != nil {
			log.Error("Copying image:", err)
		}

		return err
	})
}
