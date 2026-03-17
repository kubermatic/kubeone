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
	"strings"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/testhelper"
)

var updateFlag = flag.Bool("update", false, "update testdata files")

func Test_marshalContainerdConfig(t *testing.T) {
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
			name: "multi registry mirrors",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"registry.k8s.io": {
					Mirrors: []string{"https://some"},
				},
				"*": {
					Mirrors: []string{"https://custom.insecure.registry"},
					TLSConfig: &kubeoneapi.ContainerdTLSConfig{
						InsecureSkipVerify: true,
					},
				},
			})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalContainerdConfig(tt.cluster)
			if err != nil {
				t.Errorf("marshalContainerdConfig() error = %v", err)
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
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

func TestMarshalRegistryHostsConfig(t *testing.T) {
	tests := []struct {
		name          string
		cluster       *kubeoneapi.KubeOneCluster
		expectedPaths []string
	}{
		{
			name:    "simple - docker.io only",
			cluster: genCluster(),
			expectedPaths: []string{
				"/etc/containerd/certs.d/docker.io/hosts.toml",
			},
		},
		{
			name: "docker.io with mirror",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"docker.io": {
					Mirrors: []string{"https://custom.secure.registry"},
				},
			})),
			expectedPaths: []string{
				"/etc/containerd/certs.d/docker.io/hosts.toml",
			},
		},
		{
			name: "multiple registries",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"docker.io": {
					Mirrors: []string{"https://mirror.example.com"},
				},
				"registry.k8s.io": {
					Mirrors: []string{"https://k8s-mirror.example.com"},
				},
			})),
			expectedPaths: []string{
				"/etc/containerd/certs.d/docker.io/hosts.toml",
				"/etc/containerd/certs.d/registry.k8s.io/hosts.toml",
			},
		},
		{
			name: "insecure registry",
			cluster: genCluster(withRegistryConfiguration(kubeoneapi.RegistryConfiguration{
				OverwriteRegistry: "insecure.registry:5000",
				InsecureRegistry:  true,
			})),
			expectedPaths: []string{
				"/etc/containerd/certs.d/docker.io/hosts.toml",
				"/etc/containerd/certs.d/insecure.registry:5000/hosts.toml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MarshalRegistryHostsConfig(tt.cluster)

			// Check that all expected paths are present
			for _, expectedPath := range tt.expectedPaths {
				if _, ok := got[expectedPath]; !ok {
					t.Errorf("MarshalRegistryHostsConfig() missing expected path %q", expectedPath)
				}
			}

			// Check that we don't have unexpected paths
			if len(got) != len(tt.expectedPaths) {
				t.Errorf("MarshalRegistryHostsConfig() returned %d paths, expected %d", len(got), len(tt.expectedPaths))
				for path := range got {
					t.Logf("  got path: %s", path)
				}
			}

			// For each path, verify the content is valid TOML and contains expected fields
			for path, content := range got {
				if content == "" {
					t.Errorf("MarshalRegistryHostsConfig() path %q has empty content", path)
				}
				// Basic sanity check - content should contain server field
				if !strings.Contains(content, "server = ") {
					t.Errorf("MarshalRegistryHostsConfig() path %q content missing 'server' field", path)
				}
			}
		})
	}
}
