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

package images

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/templates/images"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
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
)

// versionConstant names from Kubernetes source that store version information.
const (
	versionConstEtcd    = "DefaultEtcdVersion" // Constant name for etcd version
	versionConstCoreDNS = "CoreDNSVersion"     // Constant name for CoreDNS version
)

// kubernetesRegistry is the default container registry for Kubernetes components.
const kubernetesRegistry = "registry.k8s.io"

// GetControlPlaneImages returns a list of container images used on control plane node.
func GetControlPlaneImages(ctx context.Context, version string) ([]string, error) {
	images := make([]string, 0)

	// start with core kubernetes images
	for _, component := range []string{componentAPIServer, componentControllerManager, componentScheduler, componentKubeProxy} {
		images = append(images, getGenericImage(kubernetesRegistry, component, version))
	}

	for _, component := range []string{componentEtcd, componentCoreDNS} {
		img, err := getComponentImage(ctx, component, version)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	pauseImage, err := kubeoneapi.SandboxImage(version, kubernetesRegistry)
	if err != nil {
		return nil, err
	}

	images = append(images, pauseImage)

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
		dests, err := retagImage(log, source, registry)
		if err != nil {
			return 0, len(images), fmt.Errorf("failed to prepare image: %w", err)
		}

		log := log.WithFields(logrus.Fields{
			"count": fmt.Sprintf("%d/%d", i+1, len(images)),
			"image": source,
		})

		if dryRun {
			for _, d := range dests {
				log.WithField("target", d).Info("Dry run")
			}

			continue
		}

		for _, dest := range dests {
			if err := copyWithRetry(ctx, log, source, dest, userAgent, insecure); err != nil {
				log.Errorf("Failed to copy image to %s: %v", dest, err)
				failedImages = append(failedImages, fmt.Sprintf("  - %s → %s", source, dest))
			}
		}
	}

	copied := len(images) - len(failedImages)
	if len(failedImages) > 0 {
		return copied, len(images), fmt.Errorf("failed images:\n%s", strings.Join(failedImages, "\n"))
	}

	return copied, len(images), nil
}

func retagImage(log logrus.FieldLogger, source, registry string) ([]string, error) {
	ref, err := name.ParseReference(source)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference: %w", err)
	}

	repo := ref.Context().RepositoryStr()
	tag := ref.Identifier()

	// Special Case: CoreDNS Requires Dual Retagging
	//
	// Background:
	// CoreDNS is the only Kubernetes image whose repository layout changes
	// depending on how the cluster pulls images.
	//
	// 1. Default kubeadm behavior (no custom registry):
	//    - kubeadm uses the upstream image:
	//          registry.k8s.io/coredns/coredns:<tag>
	//      (Notice the nested path "coredns/coredns")
	//
	// 2. kubeadm with a custom imageRepository (no containerd mirrors):
	//    - kubeadm constructs the image as:
	//          <custom-registry>/coredns:<tag>
	//      NOT:
	//          <custom-registry>/coredns/coredns:<tag>
	//    - This is hardcoded in kubeadm's GetDNSImage():
	//      for custom registries, kubeadm does NOT preserve the nested path.
	//
	// 3. KubeOne with containerd registry mirrors enabled:
	//    - containerd mirrors rewrite registry.k8s.io → <mirror>, but they
	//      still expect the original repository path:
	//          <mirror>/coredns/coredns:<tag>
	//    - Therefore, the nested path *must* exist in the custom registry
	//      so that containerd mirror lookups resolve correctly.
	//
	// Because both consumers (kubeadm and containerd mirrors) may pull different
	// versions of the path, KubeOne must mirror CoreDNS under BOTH targets:
	//
	//      1. <registry>/coredns/coredns:<tag>   (for containerd mirrors)
	//      2. <registry>/coredns:<tag>           (for kubeadm custom registry)
	//
	// Summary of required behavior:
	// - Source: registry.k8s.io/coredns/coredns:<tag>
	// - Destination mirrors:
	//        a) <registry>/coredns/coredns:<tag>
	//        b) <registry>/coredns:<tag>
	//
	// This dual-retagging ensures compatibility with:
	//   - kubeadm’s custom registry logic
	//   - containerd registry mirror rewrites
	//   - legacy consumers still expecting nested paths
	if repo == "coredns/coredns" {
		targets := []string{
			fmt.Sprintf("%s/coredns/coredns:%s", registry, tag),
			fmt.Sprintf("%s/coredns:%s", registry, tag),
		}

		log.WithField("targets", targets).Debug("CoreDNS dual-image retagging")

		return targets, nil
	}

	// Default case
	dest := fmt.Sprintf("%s/%s:%s", registry, repo, tag)

	return []string{dest}, nil
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

// FetchLatestPatchForKubernetesVersion fetches the latest patch for a specific Kubernetes version.
func FetchLatestPatchForKubernetesVersion(version string) (string, error) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(stableVersionURL, version), nil)
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

func GetKubeoneImages(ctx context.Context, filter string, versions []string) ([]string, error) {
	var err error
	var cpImages []string
	controlPlaneOnly := false
	listFilter := images.ListFilterNone
	switch filter {
	case "none":
	case "control-plane":
		controlPlaneOnly = true
	case "base":
		listFilter = images.ListFilterBase
	case "optional":
		listFilter = images.ListFilterOpional
	default:
		return nil, fail.RuntimeError{
			Op:  "checking filter flag",
			Err: errors.New("--filter can be only one of [none|base|optional|control-plane]"),
		}
	}

	imageSet := sets.New[string]()
	for _, ver := range versions {
		version := fmt.Sprintf("v%s", ver)
		cpImages, err = GetControlPlaneImages(ctx, version)
		if err != nil {
			return nil, err
		}

		imageSet.Insert(cpImages...)
		if !controlPlaneOnly {
			resolver := newImageResolver(version)
			images := resolver.List(listFilter)
			imageSet.Insert(images...)
		}
	}

	return sets.List(imageSet), nil
}

func newImageResolver(kubernetesVersion string) *images.Resolver {
	var resolveropts []images.Opt
	if kubernetesVersion != "" {
		resolveropts = append(resolveropts,
			images.WithKubernetesVersionGetter(func() string {
				return kubernetesVersion
			}),
		)
	}

	return images.NewResolver(resolveropts...)
}
