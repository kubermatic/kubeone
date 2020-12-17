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

package state

import (
	"reflect"
	"testing"

	"github.com/Masterminds/semver"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

func TestState_ContainerRuntimeConfig(t *testing.T) {
	tests := []struct {
		name    string
		version string
		input   kubeoneapi.ContainerRuntimeConfig
		want    kubeoneapi.ContainerRuntimeConfig
	}{
		{
			name:    "docker kubernetes 1.22.0",
			version: "1.22.0",
			input: kubeoneapi.ContainerRuntimeConfig{
				Docker: &kubeoneapi.ContainerRuntimeDocker{},
			},
			want: kubeoneapi.ContainerRuntimeConfig{
				Docker: &kubeoneapi.ContainerRuntimeDocker{},
			},
		},
		{
			name:    "docker kubernetes 1.23.0",
			version: "1.23.0",
			input: kubeoneapi.ContainerRuntimeConfig{
				Docker: &kubeoneapi.ContainerRuntimeDocker{},
			},
			want: kubeoneapi.ContainerRuntimeConfig{
				Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
			},
		},
		{
			name:    "containerd kubernetes 1.22.0",
			version: "1.22.0",
			input: kubeoneapi.ContainerRuntimeConfig{
				Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
			},
			want: kubeoneapi.ContainerRuntimeConfig{
				Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
			},
		},
		{
			name:    "containerd kubernetes 1.23.0",
			version: "1.23.0",
			input: kubeoneapi.ContainerRuntimeConfig{
				Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
			},
			want: kubeoneapi.ContainerRuntimeConfig{
				Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
			},
		},
		{
			name:    "default kubernetes 1.22.0",
			version: "1.22.0",
			want: kubeoneapi.ContainerRuntimeConfig{
				Docker: &kubeoneapi.ContainerRuntimeDocker{},
			},
		},
		{
			name:    "default kubernetes 1.23.0",
			version: "1.23.0",
			want: kubeoneapi.ContainerRuntimeConfig{
				Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			version := semver.MustParse(tt.version)
			s := &State{
				Cluster: &kubeoneapi.KubeOneCluster{
					ContainerRuntime: tt.input,
				},
				LiveCluster: &Cluster{
					ExpectedVersion: version,
				},
			}

			if got := s.ContainerRuntimeConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerRuntimeConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
