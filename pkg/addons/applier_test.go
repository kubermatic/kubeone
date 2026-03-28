/*
Copyright 2026 The KubeOne Authors.

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

package addons

import (
	"reflect"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

func Test_containerdRegistryCredentials(t *testing.T) {
	tests := []struct {
		name     string
		config   *kubeoneapi.ContainerRuntimeContainerd
		expected []registryCredentialsContainer
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: nil,
		},
		{
			name: "registry without auth",
			config: &kubeoneapi.ContainerRuntimeContainerd{
				Registries: map[string]kubeoneapi.ContainerdRegistry{
					"docker.io": {
						Mirrors: []string{"https://mirror.example.com"},
					},
				},
			},
			expected: nil,
		},
		{
			name: "registry with auth no mirrors",
			config: &kubeoneapi.ContainerRuntimeContainerd{
				Registries: map[string]kubeoneapi.ContainerdRegistry{
					"docker.io": {
						Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
							Username: "user",
							Password: "pass",
						},
					},
				},
			},
			expected: []registryCredentialsContainer{
				{
					RegistryName: "docker.io",
					Auth:         kubeoneapi.ContainerdRegistryAuthConfig{Username: "user", Password: "pass"},
				},
			},
		},
		{
			name: "registry with auth and mirrors",
			config: &kubeoneapi.ContainerRuntimeContainerd{
				Registries: map[string]kubeoneapi.ContainerdRegistry{
					"docker.io": {
						Mirrors: []string{"https://mirror.example.com"},
						Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
							Username: "user",
							Password: "pass",
						},
					},
				},
			},
			expected: []registryCredentialsContainer{
				{
					RegistryName: "docker.io",
					Auth:         kubeoneapi.ContainerdRegistryAuthConfig{Username: "user", Password: "pass"},
				},
				{
					RegistryName: "mirror.example.com",
					Auth:         kubeoneapi.ContainerdRegistryAuthConfig{Username: "user", Password: "pass"},
				},
			},
		},
		{
			name: "subpath registry with auth",
			config: &kubeoneapi.ContainerRuntimeContainerd{
				Registries: map[string]kubeoneapi.ContainerdRegistry{
					"gitlab.com/project1/repo1": {
						Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
							Auth: "token1",
						},
					},
				},
			},
			expected: []registryCredentialsContainer{
				{
					RegistryName: "gitlab.com",
					Auth:         kubeoneapi.ContainerdRegistryAuthConfig{Auth: "token1"},
				},
			},
		},
		{
			name: "custom port registry with auth and mirrors",
			config: &kubeoneapi.ContainerRuntimeContainerd{
				Registries: map[string]kubeoneapi.ContainerdRegistry{
					"myregistry.io:5000": {
						Mirrors: []string{"https://mirror.myregistry.io"},
						Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
							Auth: "token1",
						},
					},
				},
			},
			expected: []registryCredentialsContainer{
				{
					RegistryName: "myregistry.io:5000",
					Auth:         kubeoneapi.ContainerdRegistryAuthConfig{Auth: "token1"},
				},
				{
					RegistryName: "mirror.myregistry.io",
					Auth:         kubeoneapi.ContainerdRegistryAuthConfig{Auth: "token1"},
				},
			},
		},
		{
			name: "dedup mirror same as source host",
			config: &kubeoneapi.ContainerRuntimeContainerd{
				Registries: map[string]kubeoneapi.ContainerdRegistry{
					"docker.io": {
						Mirrors: []string{"https://docker.io"},
						Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
							Username: "user",
							Password: "pass",
						},
					},
				},
			},
			expected: []registryCredentialsContainer{
				{
					RegistryName: "docker.io",
					Auth:         kubeoneapi.ContainerdRegistryAuthConfig{Username: "user", Password: "pass"},
				},
			},
		},
		{
			name: "wildcard registry with auth and mirrors",
			config: &kubeoneapi.ContainerRuntimeContainerd{
				Registries: map[string]kubeoneapi.ContainerdRegistry{
					"*": {
						Mirrors: []string{"https://docker.io"},
						Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
							Username: "user",
							Password: "pass",
						},
					},
				},
			},
			expected: []registryCredentialsContainer{
				{
					RegistryName: "docker.io",
					Auth:         kubeoneapi.ContainerdRegistryAuthConfig{Username: "user", Password: "pass"},
				},
			},
		},
		{
			name: "wildcard registry with auth",
			config: &kubeoneapi.ContainerRuntimeContainerd{
				Registries: map[string]kubeoneapi.ContainerdRegistry{
					"*": {
						Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
							Username: "user",
							Password: "pass",
						},
					},
				},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containerdRegistryCredentials(tt.config)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("containerdRegistryCredentials() =\n  %+v\nwant:\n  %+v", got, tt.expected)
			}
		})
	}
}
