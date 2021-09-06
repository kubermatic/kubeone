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
)

// Leader returns the first configured host. Only call this after
// validating the cluster config to ensure a leader exists.
func (c KubeOneCluster) Leader() (HostConfig, error) {
	for _, host := range c.ControlPlane.Hosts {
		if host.IsLeader {
			return host, nil
		}
	}
	return HostConfig{}, errors.New("leader not found")
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
func (h *HostConfig) SetOperatingSystem(os OperatingSystemName) {
	h.OperatingSystem = os
}

// SetLeader sets is the given host leader
func (h *HostConfig) SetLeader(leader bool) {
	h.IsLeader = leader
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
		return fmt.Errorf("unknown container runtime: %q", text)
	}

	return nil
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
	case p.Openstack != nil:
		return "openstack"
	case p.Packet != nil:
		return "packet"
	case p.Vsphere != nil:
		return "vsphere"
	case p.None != nil:
		return "none"
	}

	return ""
}

// CloudProviderInTree detects is there in-tree cloud provider implementation for specified provider.
// List of in-tree provider can be found here: https://github.com/kubernetes/kubernetes/tree/master/pkg/cloudprovider
func (p CloudProviderSpec) CloudProviderInTree() bool {
	if p.Openstack != nil || p.Vsphere != nil {
		return !p.External
	} else if p.AWS != nil || p.GCE != nil || p.Azure != nil {
		return true
	}

	return false
}

// CSIMigrationSupported returns if CSI migration is supported for the specified provider.
// NB: The CSI migration can be supported only if KubeOne supports CSI plugin and driver
// for the provider
func (p CloudProviderSpec) CSIMigrationSupported() bool {
	return p.External && (p.Openstack != nil || p.Vsphere != nil)
}

// CSIMigrationFeatureGates returns CSI migration feature gates in form of a map
// (to be used with Kubelet config) and string (to be used with kube-apiserver
// and kube-controller-manager)
// NB: We're intentionally not enabling CSIMigration feature gate because it's
// enabled by default since Kubernetes 1.18
// (https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/)
// This is a KubeOneCluster function because feature gates are Kubernetes-version dependent.
func (c KubeOneCluster) CSIMigrationFeatureGates(complete bool) (map[string]bool, string, error) {
	switch {
	case c.CloudProvider.Openstack != nil:
		featureGates := map[string]bool{
			"CSIMigrationOpenStack": true,
			"ExpandCSIVolumes":      true,
		}

		unregister := c.InTreePluginUnregisterFeatureGate()
		if complete && unregister != "" {
			featureGates[unregister] = true
		}

		return featureGates, marshalFeatureGates(featureGates), nil
	case c.CloudProvider.Vsphere != nil:
		featureGates := map[string]bool{
			"CSIMigrationvSphere": true,
		}

		unregister := c.InTreePluginUnregisterFeatureGate()
		if complete && unregister != "" {
			featureGates[unregister] = true
		}

		return featureGates, marshalFeatureGates(featureGates), nil
	}

	return nil, "", errors.New("csi migration is not supported for selected provider")
}

// CSIMigrationFeatureGates returns the name of the feature gate that's supposed to
// unregister the in-tree cloud provider.
// NB: This is a KubeOneCluster function because feature gates are Kubernetes-version dependent.
func (c KubeOneCluster) InTreePluginUnregisterFeatureGate() string {
	lessThan21, _ := semver.NewConstraint("< 1.21.0")
	ver, _ := semver.NewVersion(c.Versions.Kubernetes)

	switch {
	case c.CloudProvider.Openstack != nil:
		if lessThan21.Check(ver) {
			return "CSIMigrationOpenStackComplete"
		}
		return "InTreePluginOpenStackUnregister"
	case c.CloudProvider.Vsphere != nil:
		if lessThan21.Check(ver) {
			return "CSIMigrationvSphereComplete"
		}
		return "InTreePluginvSphereUnregister"
	}

	return ""
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
			return "", errors.Wrap(err, "unable to get absolute path to the cluster manifest")
		}
		addonsPath = filepath.Join(manifestAbsPath, addonsPath)
	}

	return addonsPath, nil
}
