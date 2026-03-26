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
	"flag"
	"fmt"
	"strings"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/testhelper"
)

var updateFlag = flag.Bool("update", false, "update testdata files")

func Test_ContainerdConfigs(t *testing.T) {
	tests := []struct {
		name    string
		cluster *kubeoneapi.KubeOneCluster
	}{
		{
			name:    "simple",
			cluster: genCluster(),
		},
		{
			name: "override insecure registry",
			cluster: genCluster(withRegistryConfiguration(kubeoneapi.RegistryConfiguration{
				OverwriteRegistry: "some.registry",
				InsecureRegistry:  true,
			})),
		},
		{
			name: "insecure registry with skip_verify",
			cluster: genCluster(
				withRegistryConfiguration(kubeoneapi.RegistryConfiguration{
					OverwriteRegistry: "some.registry",
					InsecureRegistry:  true,
				}),
				withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
					"some.registry": {
						TLSConfig: &kubeoneapi.ContainerdTLSConfig{
							InsecureSkipVerify: true,
						},
					},
					"some-1.registry": {
						Mirrors: []string{"some.registry"},
						TLSConfig: &kubeoneapi.ContainerdTLSConfig{
							InsecureSkipVerify: true,
						},
					},
					"some-2.registry": {
						Mirrors: []string{"some.registry"},
						TLSConfig: &kubeoneapi.ContainerdTLSConfig{
							InsecureSkipVerify: false,
						},
					},
				}),
			),
		},
		{
			name: "docker.io mirror registry",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"docker.io": {
					Mirrors: []string{"https://custom.secure.registry"},
				},
			})),
		},
		{
			name: "mirror registry with override_path",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"docker.io": {
					Mirrors:      []string{"https://custom.secure.registry/v2/someproject"},
					OverridePath: true,
				},
			})),
		},
		{
			name: "mirror registry with plain-text http",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"docker.io": {
					Mirrors: []string{"http://custom.insecure.registry"},
				},
			})),
		},
		{
			name: "registry in subpath",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"gitlab.com/project1/repo1": {
					Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
						Auth: "token1",
					},
				},
			})),
		},
		{
			name: "registry in subpath and override_path",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"gitlab.com/project1/repo1": {
					Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
						Auth: "token1",
					},
				},
				"gitlab.com/project1/repo2": {
					Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
						Auth: "token1",
					},
					OverridePath: true,
				},
			})),
		},
		{
			name: "multi registry mirrors",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"registry.k8s.io": {
					Mirrors: []string{"https://some", "https://other"},
				},
				"*": {
					Mirrors: []string{"https://custom.insecure.registry"},
					TLSConfig: &kubeoneapi.ContainerdTLSConfig{
						InsecureSkipVerify: true,
					},
				},
			})),
		},
		{
			name: "registry/mirror-registry with auth",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"docker.io": {
					Mirrors: []string{"https://mirror.example.com"},
					Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
						Username: "testuser",
						Password: "testpass",
					},
				},
				"gcr.io": {
					Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
						Username: "testuser",
						Password: "testpass",
					},
				},
				"registry.k8s.io": {
					Mirrors: []string{"https://mirror.example.com", "https://mirror2.example.com"},
					Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
						Username: "testuser",
						Password: "testpass",
					},
				},
			})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder

			cr2Configs, err := marshalContainerdConfigs(tt.cluster)
			if err != nil {
				t.Errorf("marshalContainerdConfigs() error = %v", err)
			}

			for path, config := range cr2Configs.Iter() {
				fmt.Fprintf(&buf, "### %s\n", path)
				fmt.Fprintf(&buf, "%s\n", config)
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), buf.String(), *updateFlag)
		})
	}
}

type clusterOpts func(*kubeoneapi.KubeOneCluster)

func genCluster(opts ...clusterOpts) *kubeoneapi.KubeOneCluster {
	cls := &kubeoneapi.KubeOneCluster{
		Versions: kubeoneapi.VersionConfig{
			Kubernetes: "1.27.0",
		},
		RegistryConfiguration: nil,
		ContainerRuntime: kubeoneapi.ContainerRuntimeConfig{
			Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
		},
	}

	for _, o := range opts {
		o(cls)
	}

	return cls
}

func withRegistryConfiguration(regCfg kubeoneapi.RegistryConfiguration) clusterOpts {
	return func(cls *kubeoneapi.KubeOneCluster) {
		cls.RegistryConfiguration = &regCfg
	}
}

func withContainerdRegistry(regCfg map[string]kubeoneapi.ContainerdRegistry) clusterOpts {
	return func(cls *kubeoneapi.KubeOneCluster) {
		cls.ContainerRuntime.Containerd.Registries = regCfg
	}
}
