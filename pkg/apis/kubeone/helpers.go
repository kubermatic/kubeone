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

package kubeone

import (
	"bytes"
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/templates/resources"
)

const (
	credentialSecretName = "kube-system/kubeone-registry-credentials" //nolint:gosec
)

// Leader returns the first configured host. Only call this after
// validating the cluster config to ensure a leader exists.
func (c KubeOneCluster) Leader() (HostConfig, error) {
	for _, host := range c.ControlPlane.Hosts {
		if host.IsLeader {
			return host, nil
		}
	}

	return HostConfig{}, fail.ConfigError{
		Op:  "leader",
		Err: errors.New("not found"),
	}
}

func (c KubeOneCluster) RandomHost() HostConfig {
	//nolint:gosec
	// G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec)
	n := rand.Int31n(int32(len(c.ControlPlane.Hosts)))

	return c.ControlPlane.Hosts[n]
}

// Followers returns all but the first configured host. Only call
// this after validating the cluster config to ensure hosts exist.
func (c KubeOneCluster) Followers() []HostConfig {
	followers := []HostConfig{}
	for _, h := range c.ControlPlane.Hosts {
		if !h.IsLeader {
			followers = append(followers, h)
		}
	}

	return followers
}

// IsManagedNode reports whether given node name is known to the KubeOne configuration
func (c *KubeOneCluster) IsManagedNode(nodename string) bool {
	for _, host := range append(c.ControlPlane.Hosts, c.StaticWorkers.Hosts...) {
		if host.Hostname == nodename {
			return true
		}
	}

	return false
}

// SetHostname sets the hostname for the given host
func (h *HostConfig) SetHostname(hostname string) {
	h.Hostname = hostname
}

// SetOperatingSystem sets the operating system for the given host
func (h *HostConfig) SetOperatingSystem(os OperatingSystemName) error {
	if h.OperatingSystem.IsValid() {
		h.OperatingSystem = os

		return nil
	}

	return fail.ConfigValidation(fmt.Errorf("unknown operating system %q", os))
}

func (osName OperatingSystemName) IsValid() bool {
	// linter exhaustive will make sure this switch is iterating over all current and future possibilities
	switch osName {
	case OperatingSystemNameUbuntu:
	case OperatingSystemNameDebian:
	case OperatingSystemNameCentOS:
	case OperatingSystemNameRHEL:
	case OperatingSystemNameRockyLinux:
	case OperatingSystemNameAmazon:
	case OperatingSystemNameFlatcar:
	case OperatingSystemNameUnknown:
	default:
		return false
	}

	return true
}

// SetLeader sets is the given host leader
func (h *HostConfig) SetLeader(leader bool) {
	h.IsLeader = leader
}

func (c KubeOneCluster) OperatingSystemManagerEnabled() bool {
	if c.Addons.Enabled() {
		for _, embeddedAddon := range c.Addons.Addons {
			if embeddedAddon.Name == resources.AddonOperatingSystemManager && !embeddedAddon.Delete {
				return true
			}
		}
	}

	return false
}

func (c KubeOneCluster) OperatingSystemManagerQueuedForDeletion() bool {
	if c.Addons.Enabled() {
		for _, embeddedAddon := range c.Addons.Addons {
			if embeddedAddon.Name == resources.AddonOperatingSystemManager && embeddedAddon.Delete {
				return true
			}
		}
	}

	return false
}

func (crc ContainerRuntimeConfig) MachineControllerFlags() []string {
	var mcFlags []string
	switch {
	case crc.Docker != nil:
		if len(crc.Docker.RegistryMirrors) > 0 {
			mcFlags = append(mcFlags,
				fmt.Sprintf("-node-registry-mirrors=%s", strings.Join(crc.Docker.RegistryMirrors, ",")),
			)
		}
	case crc.Containerd != nil:
		// example output:
		// -node-containerd-registry-mirrors=docker.io=custom.tld
		// -node-containerd-registry-mirrors=docker.io=https://secure-custom.tld
		// -node-containerd-registry-mirrors=k8s.gcr.io=http://somewhere
		// -node-insecure-registries=docker.io,k8s.gcr.io
		var (
			registryNames                 []string
			insecureSet                   = map[string]struct{}{}
			registryCredentialsSecretFlag bool
		)

		for registry := range crc.Containerd.Registries {
			registryNames = append(registryNames, registry)
		}

		// because iterating over map is randomized, we need this to have a "stable" output list
		sort.Strings(registryNames)

		for _, registryName := range registryNames {
			containerdRegistry := crc.Containerd.Registries[registryName]
			if containerdRegistry.TLSConfig != nil && containerdRegistry.TLSConfig.InsecureSkipVerify {
				insecureSet[registryName] = struct{}{}
			}

			for _, mirror := range containerdRegistry.Mirrors {
				mcFlags = append(mcFlags,
					fmt.Sprintf("-node-containerd-registry-mirrors=%s=%s", registryName, mirror),
				)
			}

			if containerdRegistry.Auth != nil {
				registryCredentialsSecretFlag = true
			}
		}

		if registryCredentialsSecretFlag {
			mcFlags = append(mcFlags,
				fmt.Sprintf("-node-registry-credentials-secret=%s", credentialSecretName),
			)
		}

		if len(insecureSet) > 0 {
			insecureNames := []string{}

			for insecureName := range insecureSet {
				insecureNames = append(insecureNames, insecureName)
			}

			sort.Strings(insecureNames)
			mcFlags = append(mcFlags,
				fmt.Sprintf("-node-insecure-registries=%s", strings.Join(insecureNames, ",")),
			)
		}
	}

	return mcFlags
}

func (crc ContainerRuntimeConfig) String() string {
	switch {
	case crc.Containerd != nil:
		return "containerd"
	case crc.Docker != nil:
		return "docker"
	}

	return "unknown"
}

func (crc *ContainerRuntimeConfig) UnmarshalText(text []byte) error {
	switch {
	case bytes.Equal(text, []byte("docker")):
		*crc = ContainerRuntimeConfig{Docker: &ContainerRuntimeDocker{}}
	case bytes.Equal(text, []byte("containerd")):
		*crc = ContainerRuntimeConfig{Containerd: &ContainerRuntimeContainerd{}}
	default:
		return fail.ConfigValidation(fmt.Errorf("unknown container runtime: %q", text))
	}

	return nil
}

func (crc ContainerRuntimeConfig) ConfigPath() string {
	switch {
	case crc.Containerd != nil:
		return "/etc/containerd/config.toml"
	case crc.Docker != nil:
		return "/etc/docker/daemon.json"
	}

	return ""
}

func (crc ContainerRuntimeConfig) CRISocket() string {
	switch {
	case crc.Containerd != nil:
		return "/run/containerd/containerd.sock"
	case crc.Docker != nil:
		return "/var/run/dockershim.sock"
	}

	return ""
}

// CloudProviderName returns name of the cloud provider
func (p CloudProviderSpec) CloudProviderName() string {
	switch {
	case p.AWS != nil:
		return "aws"
	case p.Azure != nil:
		return "azure"
	case p.DigitalOcean != nil:
		return "digitalocean"
	case p.GCE != nil:
		return "gce"
	case p.Hetzner != nil:
		return "hetzner"
	case p.Nutanix != nil:
		return "nutanix"
	case p.Openstack != nil:
		return "openstack"
	case p.EquinixMetal != nil:
		return "equinixmetal"
	case p.Vsphere != nil:
		return "vsphere"
	case p.VMwareCloudDirector != nil:
		return "vmwareCloudDirector"
	case p.None != nil:
		return "none"
	}

	return ""
}

// MachineControllerCloudProvider returns name of the cloud provider for machine-controller
// It handles special cases where the cloud provider name in KubeOne might differ to that required in machine-controller.
func (p CloudProviderSpec) MachineControllerCloudProvider() string {
	switch {
	case p.VMwareCloudDirector != nil:
		return "vmware-cloud-director"
	default:
		return p.CloudProviderName()
	}
}

// CloudProviderInTree detects is there in-tree cloud provider implementation for specified provider.
// List of in-tree provider can be found here: https://github.com/kubernetes/kubernetes/tree/master/pkg/cloudprovider
func (p CloudProviderSpec) CloudProviderInTree() bool {
	if p.AWS != nil || p.Azure != nil || p.Openstack != nil || p.Vsphere != nil {
		return !p.External
	} else if p.GCE != nil {
		return true
	}

	return false
}

// CSIMigrationSupported returns if CSI migration is supported for the specified provider.
// NB: The CSI migration can be supported only if KubeOne supports CSI plugin and driver
// for the provider
func (p CloudProviderSpec) CSIMigrationSupported() bool {
	return p.External && (p.Azure != nil || p.Openstack != nil || p.Vsphere != nil)
}

// CSIMigrationFeatureGates returns CSI migration feature gates in form of a map
// (to be used with Kubelet config) and string (to be used with kube-apiserver
// and kube-controller-manager)
// NB: We're intentionally not enabling CSIMigration feature gate because it's
// enabled by default since Kubernetes 1.18
// (https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/)
// This is a KubeOneCluster function because feature gates are Kubernetes-version dependent.
func (c KubeOneCluster) CSIMigrationFeatureGates(complete bool) (map[string]bool, string, error) {
	var featureGates map[string]bool

	switch {
	case c.CloudProvider.Azure != nil:
		featureGates = map[string]bool{
			"CSIMigrationAzureDisk": true,
			"CSIMigrationAzureFile": true,
		}
	case c.CloudProvider.Openstack != nil:
		featureGates = map[string]bool{
			"CSIMigrationOpenStack": true,
			"ExpandCSIVolumes":      true,
		}
	case c.CloudProvider.Vsphere != nil:
		featureGates = map[string]bool{
			"CSIMigrationvSphere": true,
		}
	default:
		return nil, "", fail.ConfigValidation(fmt.Errorf("csi migration is not supported for selected provider"))
	}

	if complete {
		for _, u := range c.InTreePluginUnregisterFeatureGate() {
			featureGates[u] = true
		}
	}

	return featureGates, marshalFeatureGates(featureGates), nil
}

// CSIMigrationFeatureGates returns the name of the feature gate that's supposed to
// unregister the in-tree cloud provider.
// NB: This is a KubeOneCluster function because feature gates are Kubernetes-version dependent.
func (c KubeOneCluster) InTreePluginUnregisterFeatureGate() []string {
	lessThan21, _ := semver.NewConstraint("< 1.21.0")
	ver, _ := semver.NewVersion(c.Versions.Kubernetes)

	switch {
	case c.CloudProvider.Azure != nil:
		if lessThan21.Check(ver) {
			return []string{"CSIMigrationAzureDiskComplete", "CSIMigrationAzureFileComplete"}
		}

		return []string{"InTreePluginAzureDiskUnregister", "InTreePluginAzureFileUnregister"}
	case c.CloudProvider.Openstack != nil:
		if lessThan21.Check(ver) {
			return []string{"CSIMigrationOpenStackComplete"}
		}

		return []string{"InTreePluginOpenStackUnregister"}
	case c.CloudProvider.Vsphere != nil:
		if lessThan21.Check(ver) {
			return []string{"CSIMigrationvSphereComplete"}
		}

		return []string{"InTreePluginvSphereUnregister"}
	}

	return nil
}

func marshalFeatureGates(fgm map[string]bool) string {
	keys := []string{}
	for k, v := range fgm {
		keys = append(keys, fmt.Sprintf("%s=%t", k, v))
	}

	sort.Strings(keys)

	return strings.Join(keys, ",")
}

// ImageRegistry returns the image registry to use or the passed in
// default if no override is specified
func (r *RegistryConfiguration) ImageRegistry(defaultRegistry string) string {
	if r != nil && r.OverwriteRegistry != "" {
		return r.OverwriteRegistry
	}

	return defaultRegistry
}

// InsecureRegistryAddress returns the registry that should be configured
// as insecure if InsecureRegistry is enabled
func (r *RegistryConfiguration) InsecureRegistryAddress() string {
	insecureRegistry := ""
	if r != nil && r.InsecureRegistry {
		insecureRegistry = r.OverwriteRegistry
	}

	return insecureRegistry
}

func (ads *Addons) Enabled() bool {
	return ads != nil && ads.Enable
}

// RelativePath returns addons path relative to the KubeOneCluster manifest file
// path
func (ads *Addons) RelativePath(manifestFilePath string) (string, error) {
	addonsPath := ads.Path
	if !filepath.IsAbs(addonsPath) && manifestFilePath != "" {
		manifestAbsPath, err := filepath.Abs(filepath.Dir(manifestFilePath))
		if err != nil {
			return "", fail.Runtime(err, "getting absolute path to the cluster manifest")
		}
		addonsPath = filepath.Join(manifestAbsPath, addonsPath)
	}

	return addonsPath, nil
}

// DefaultAssetConfiguration determines what image repository should be used
// for Kubernetes and metrics-server images. The AssetsConfiguration has the
// highest priority, then comes the RegistryConfiguration.
// This function is needed because the AssetsConfiguration API has been removed
// in the v1beta2 API, so we can't use defaulting
func (c *KubeOneCluster) DefaultAssetConfiguration() {
	if c.RegistryConfiguration == nil || c.RegistryConfiguration.OverwriteRegistry == "" {
		// We default AssetConfiguration only if RegistryConfiguration.OverwriteRegistry
		// is used
		return
	}

	c.AssetConfiguration.Kubernetes.ImageRepository = defaults(
		c.AssetConfiguration.Kubernetes.ImageRepository,
		c.RegistryConfiguration.OverwriteRegistry,
	)
	c.AssetConfiguration.CoreDNS.ImageRepository = defaults(
		c.AssetConfiguration.CoreDNS.ImageRepository,
		c.RegistryConfiguration.OverwriteRegistry,
	)
	c.AssetConfiguration.Etcd.ImageRepository = defaults(
		c.AssetConfiguration.Etcd.ImageRepository,
		c.RegistryConfiguration.OverwriteRegistry,
	)
	c.AssetConfiguration.MetricsServer.ImageRepository = defaults(
		c.AssetConfiguration.MetricsServer.ImageRepository,
		c.RegistryConfiguration.OverwriteRegistry,
	)
}

func defaults(input, defaultValue string) string {
	if input != "" {
		return input
	}

	return defaultValue
}

func MapStringStringToString(m1 map[string]string, pairSeparator string) string {
	var pairs []string
	for k, v := range m1 {
		pairs = append(pairs, fmt.Sprintf("%s%s%s", k, pairSeparator, v))
	}
	sort.Strings(pairs)

	return strings.Join(pairs, ",")
}
