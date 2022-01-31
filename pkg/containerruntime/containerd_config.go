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
	"strings"

	"github.com/BurntSushi/toml"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

type containerdConfig struct {
	Version int                    `toml:"version"`
	Metrics *containerdMetrics     `toml:"metrics"`
	Plugins map[string]interface{} `toml:"plugins"`
}

type containerdMetrics struct {
	Address string `toml:"address"`
}

type containerdCRIPlugin struct {
	Containerd *containerdCRISettings `toml:"containerd"`
	Registry   *containerdCRIRegistry `toml:"registry"`
}

type containerdCRISettings struct {
	Runtimes map[string]containerdCRIRuntime `toml:"runtimes"`
}

type containerdCRIRuntime struct {
	RuntimeType string      `toml:"runtime_type"`
	Options     interface{} `toml:"options"`
}

type containerdCRIRuncOptions struct {
	SystemdCgroup bool
}

type containerdCRIRegistry struct {
	Mirrors map[string]containerdRegistryMirror `toml:"mirrors"`
	Configs map[string]containerdRegistryConfig `toml:"configs"`
}

type containerdRegistryMirror struct {
	Endpoint []string `toml:"endpoint"`
}

type containerdRegistryConfig struct {
	TLS  *containerdRegistryTLSConfig `toml:"tls"`
	Auth *containerdRegistryAuth      `toml:"auth"`
}

type containerdRegistryAuth struct {
	Username      string `toml:"username"`
	Password      string `toml:"password"`
	Auth          string `toml:"auth"`
	IdentityToken string `toml:"identitytoken"`
}

type containerdRegistryTLSConfig struct {
	InsecureSkipVerify bool `toml:"insecure_skip_verify"`
}

func marshalContainerdConfig(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	criPlugin := containerdCRIPlugin{
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
		Registry: &containerdCRIRegistry{
			Mirrors: map[string]containerdRegistryMirror{
				"docker.io": {
					Endpoint: []string{"https://registry-1.docker.io"},
				},
			},
		},
	}

	if cluster.RegistryConfiguration != nil {
		insecureRegistry := cluster.RegistryConfiguration.InsecureRegistryAddress()
		if insecureRegistry != "" {
			criPlugin.Registry.Mirrors[insecureRegistry] = containerdRegistryMirror{
				Endpoint: []string{fmt.Sprintf("http://%s", insecureRegistry)},
			}
		}
	}

	if regs := cluster.ContainerRuntime.Containerd.Registries; regs != nil {
		criPlugin.Registry = &containerdCRIRegistry{
			Mirrors: map[string]containerdRegistryMirror{},
			Configs: map[string]containerdRegistryConfig{},
		}

		for registryName, registry := range regs {
			criPlugin.Registry.Mirrors[registryName] = containerdRegistryMirror{
				Endpoint: registry.Mirrors,
			}

			if registry.TLSConfig != nil {
				criPlugin.Registry.Configs[registryName] = containerdRegistryConfig{
					TLS: &containerdRegistryTLSConfig{
						InsecureSkipVerify: registry.TLSConfig.InsecureSkipVerify,
					},
				}
			}

			if registry.Auth != nil {
				regConfig := criPlugin.Registry.Configs[registryName]
				regConfig.Auth = &containerdRegistryAuth{
					Username:      registry.Auth.Username,
					Password:      registry.Auth.Password,
					Auth:          registry.Auth.Auth,
					IdentityToken: registry.Auth.IdentityToken,
				}
				criPlugin.Registry.Configs[registryName] = regConfig
			}
		}
	}

	cfg := containerdConfig{
		Version: 2,
		Metrics: &containerdMetrics{
			// metrics available at http://127.0.0.1:1338/v1/metrics
			Address: "127.0.0.1:1338",
		},

		Plugins: map[string]interface{}{
			"io.containerd.grpc.v1.cri": criPlugin,
		},
	}

	var buf strings.Builder
	enc := toml.NewEncoder(&buf)
	enc.Indent = ""
	err := enc.Encode(cfg)

	return buf.String(), err
}
