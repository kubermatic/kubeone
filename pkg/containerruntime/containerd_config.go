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

	"k8c.io/kubeone/pkg/apis/kubeone"
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
	Mirrors map[string]containerdMirror `toml:"mirrors"`
}

type containerdMirror struct {
	Endpoint []string `toml:"endpoint"`
}

func marshalContainerdConfig(cluster *kubeone.KubeOneCluster) (string, error) {
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
			Mirrors: map[string]containerdMirror{
				"docker.io": {
					Endpoint: []string{"https://registry-1.docker.io"},
				},
			},
		},
	}

	insecureRegistry := cluster.RegistryConfiguration.InsecureRegistryAddress()
	if insecureRegistry != "" {
		criPlugin.Registry.Mirrors[insecureRegistry] = containerdMirror{
			Endpoint: []string{fmt.Sprintf("http://%s", insecureRegistry)},
		}
	}

	if reg := cluster.RegistryConfiguration; reg != nil {
		if len(reg.Mirrors) > 0 {
			criPlugin.Registry.Mirrors["docker.io"] = containerdMirror{
				Endpoint: reg.Mirrors,
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
