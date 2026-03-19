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

package containerruntime

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"

	"k8s.io/utils/ptr"
)

const (
	// containerdRegistryConfigPath is the directory for containerd registry host configs
	containerdRegistryConfigPath = "/etc/containerd/certs.d"
)

type containerdConfig struct {
	Version int                `toml:"version"`
	Metrics *containerdMetrics `toml:"metrics"`
	Plugins map[string]any     `toml:"plugins"`
}

type containerdMetrics struct {
	Address string `toml:"address"`
}

// containerdCRIImagesPlugin represents the "io.containerd.cri.v1.images" plugin in containerd 2.x.
type containerdCRIImagesPlugin struct {
	DiscardUnpackedLayers bool                    `toml:"discard_unpacked_layers"`
	PinnedImages          *containerdPinnedImages `toml:"pinned_images,omitempty"`
	Registry              *containerdCRIRegistry  `toml:"registry"`
}

// containerdPinnedImages represents the pinned_images config in containerd 2.x.
type containerdPinnedImages struct {
	Sandbox string `toml:"sandbox,omitempty"`
}

// containerdCRIRuntimePlugin represents the "io.containerd.cri.v1.runtime" plugin in containerd 2.x.
type containerdCRIRuntimePlugin struct {
	Containerd                         *containerdCRISettings  `toml:"containerd"`
	DeviceOwnershipFromSecurityContext bool                    `toml:"device_ownership_from_security_context"`
	CNI                                *containerdCRICNIConfig `toml:"cni"`
}

// containerdCRICNIConfig represents the CNI config under the runtime plugin in containerd 2.x.
type containerdCRICNIConfig struct {
	BinDirs []string `toml:"bin_dirs"`
	ConfDir string   `toml:"conf_dir"`
}

type containerdCRISettings struct {
	Runtimes map[string]containerdCRIRuntime `toml:"runtimes"`
}

type containerdCRIRuntime struct {
	RuntimeType string `toml:"runtime_type"`
	Options     any    `toml:"options"`
}

type containerdCRIRuncOptions struct {
	SystemdCgroup bool `toml:"SystemdCgroup"`
}

type containerdCRIRegistry struct {
	ConfigPath string                              `toml:"config_path"`
	Configs    map[string]containerdRegistryConfig `toml:"configs,omitempty"`
}

type containerdRegistryConfig struct {
	Auth *containerdRegistryAuth `toml:"auth,omitempty"`
}

type containerdRegistryAuth struct {
	Username      string `toml:"username,omitempty"`
	Password      string `toml:"password,omitempty"`
	Auth          string `toml:"auth,omitempty"`
	IdentityToken string `toml:"identitytoken,omitempty"`
}

// registryHostConfig holds the parsed mirror configuration for a single registry,
// used internally when building hosts.toml files.
type registryHostConfig struct {
	endpoints    []string
	overridePath bool
	insecure     bool
}

// hostsTomlConfig represents the top-level structure of a hosts.toml file.
type hostsTomlConfig struct {
	Server string                     `toml:"server"`
	Host   map[string]hostEntryConfig `toml:"host,omitempty"`
}

// hostEntryConfig represents a single host entry in a hosts.toml file.
type hostEntryConfig struct {
	Capabilities []string `toml:"capabilities"`
	SkipVerify   bool     `toml:"skip_verify,omitempty"`
	OverridePath bool     `toml:"override_path,omitempty"`
}

func marshalContainerdConfigToml(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	var sandboxImage string
	var err error

	if cluster.ContainerRuntime.Containerd != nil && cluster.ContainerRuntime.Containerd.SandboxImage != "" {
		sandboxImage = cluster.ContainerRuntime.Containerd.SandboxImage
	} else {
		sandboxImage, err = cluster.Versions.SandboxImage(cluster.RegistryConfiguration.ImageRegistry)
		if err != nil {
			return "", fmt.Errorf("failed to determine sandbox image: %w", err)
		}
	}

	criRegistry := &containerdCRIRegistry{
		ConfigPath: containerdRegistryConfigPath,
	}

	// Add registry credentials to CRI config for authentication.
	// Per containerd v2 docs, auth is configured under
	// [plugins."io.containerd.cri.v1.images".registry.configs."<registry>".auth]
	// The registry key must be the mirror host (with optional port), not the source registry.
	if cluster.ContainerRuntime.Containerd != nil && cluster.ContainerRuntime.Containerd.Registries != nil {
		for _, registry := range cluster.ContainerRuntime.Containerd.Registries {
			if registry.Auth != nil && len(registry.Mirrors) > 0 {
				if criRegistry.Configs == nil {
					criRegistry.Configs = make(map[string]containerdRegistryConfig)
				}
				// Auth applies to the mirror endpoints, not the source registry
				for _, mirror := range registry.Mirrors {
					host := mirror
					if u, parseErr := url.Parse(mirror); parseErr == nil && u.Host != "" {
						host = u.Host
					}
					criRegistry.Configs[host] = containerdRegistryConfig{
						Auth: &containerdRegistryAuth{
							Username:      registry.Auth.Username,
							Password:      registry.Auth.Password,
							Auth:          registry.Auth.Auth,
							IdentityToken: registry.Auth.IdentityToken,
						},
					}
				}
			}
		}
	}

	criImagesPlugin := containerdCRIImagesPlugin{
		DiscardUnpackedLayers: false,
		Registry:              criRegistry,
	}

	if sandboxImage != "" {
		criImagesPlugin.PinnedImages = &containerdPinnedImages{
			Sandbox: sandboxImage,
		}
	}

	criRuntimePlugin := containerdCRIRuntimePlugin{
		DeviceOwnershipFromSecurityContext: ptr.Deref(cluster.ContainerRuntime.Containerd.DeviceOwnershipFromSecurityContext, true),
		Containerd: &containerdCRISettings{
			Runtimes: map[string]containerdCRIRuntime{
				"runc": {
					RuntimeType: "io.containerd.runc.v2",
					Options: containerdCRIRuncOptions{
						SystemdCgroup: true,
					},
				},
			},
		},
		CNI: &containerdCRICNIConfig{
			BinDirs: []string{"/opt/cni/bin"},
			ConfDir: "/etc/cni/net.d",
		},
	}

	cfg := containerdConfig{
		Version: 3,
		Metrics: &containerdMetrics{
			// metrics available at http://127.0.0.1:1338/v1/metrics
			Address: "127.0.0.1:1338",
		},

		Plugins: map[string]interface{}{
			"io.containerd.cri.v1.images":  criImagesPlugin,
			"io.containerd.cri.v1.runtime": criRuntimePlugin,
		},
	}

	var buf strings.Builder
	enc := toml.NewEncoder(&buf)
	enc.Indent = ""
	err = enc.Encode(cfg)

	return buf.String(), fail.Runtime(err, "encoding containerd config")
}

// buildRegistryHostConfigs processes the registry mirrors, insecure registries,
// and returns a per-registry configuration.
func buildRegistryHostConfigs(cluster *kubeoneapi.KubeOneCluster) map[string]*registryHostConfig {
	configs := make(map[string]*registryHostConfig)

	// Start with default docker.io entry
	configs["docker.io"] = &registryHostConfig{
		endpoints: []string{"https://registry-1.docker.io"},
	}

	// Process insecure registry from RegistryConfiguration
	if cluster.RegistryConfiguration != nil {
		insecureRegistry := cluster.RegistryConfiguration.InsecureRegistryAddress()
		if insecureRegistry != "" {
			if _, ok := configs[insecureRegistry]; !ok {
				configs[insecureRegistry] = &registryHostConfig{}
			}
			configs[insecureRegistry].insecure = true
		}
	}

	// Process registry mirrors from ContainerRuntime configuration
	if cluster.ContainerRuntime.Containerd != nil && cluster.ContainerRuntime.Containerd.Registries != nil {
		for registryName, registry := range cluster.ContainerRuntime.Containerd.Registries {
			if _, ok := configs[registryName]; !ok {
				configs[registryName] = &registryHostConfig{}
			}
			rc := configs[registryName]

			if len(registry.Mirrors) > 0 {
				rc.endpoints = registry.Mirrors
			}
			rc.overridePath = registry.OverridePath

			if registry.TLSConfig != nil && registry.TLSConfig.InsecureSkipVerify {
				rc.insecure = true
			}
		}
	}

	return configs
}

// marshalContainerdConfigs returns a map of file path to file content for containerd
// registry host configuration files. Each key is a path like
// "/etc/containerd/certs.d/<registry>/hosts.toml" and the value is the TOML content.
func marshalContainerdConfigs(cluster *kubeoneapi.KubeOneCluster) (*orderedStringMap, error) {
	result := newOrderedMap()
	crConfig, err := marshalContainerdConfigToml(cluster)
	if err != nil {
		return nil, err
	}

	result.set(cluster.ContainerRuntime.ConfigPath(), crConfig)
	configs := buildRegistryHostConfigs(cluster)

	// Sort registry names for deterministic output
	registryNames := make([]string, 0, len(configs))
	for name := range configs {
		registryNames = append(registryNames, name)
	}
	sort.Strings(registryNames)

	for _, registryName := range registryNames {
		rc := configs[registryName]

		// Determine the server URL (the upstream registry)
		serverURL := fmt.Sprintf("https://%s", registryName)
		if registryName == "docker.io" {
			serverURL = "https://registry-1.docker.io"
		}

		cfg := hostsTomlConfig{
			Server: serverURL,
			Host:   make(map[string]hostEntryConfig),
		}

		// Add mirror host entries
		for _, endpoint := range rc.endpoints {
			if !strings.HasPrefix(endpoint, "http") {
				endpoint = "https://" + endpoint
			}
			cfg.Host[endpoint] = hostEntryConfig{
				Capabilities: []string{"pull", "resolve"},
				OverridePath: rc.overridePath,
				SkipVerify:   rc.insecure,
			}
		}

		// If insecure registry has no endpoints, add its own endpoint
		if rc.insecure && len(rc.endpoints) == 0 {
			cfg.Host[serverURL] = hostEntryConfig{
				Capabilities: []string{"pull", "resolve", "push"},
				SkipVerify:   true,
			}
		}

		var buf strings.Builder
		enc := toml.NewEncoder(&buf)
		enc.Indent = ""
		_ = enc.Encode(cfg)

		// Remove empty parent table header that TOML encoder generates for nested maps
		output := strings.ReplaceAll(buf.String(), "[host]\n", "")

		filePath := fmt.Sprintf("%s/%s/hosts.toml", containerdRegistryConfigPath, registryName)
		result.set(filePath, output)
	}

	return result, nil
}
