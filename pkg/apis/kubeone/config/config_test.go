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

package config

import (
	"reflect"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

func Test_setRegistriesAuth(t *testing.T) {
	type args struct {
		cluster *kubeoneapi.KubeOneCluster
		buf     string
	}

	tests := []struct {
		name           string
		args           args
		exampleCluster *kubeoneapi.KubeOneCluster
		wantErr        bool
	}{
		{
			name: "no apiVersion",
			args: args{
				buf: `registries: {}`,
			},
			wantErr: true,
		},
		{
			name: "no kind",
			args: args{
				buf: heredoc.Doc(`
					apiVersion: kubeone.k8c.io/v1beta2
					registries: {}
				`),
			},
			wantErr: true,
		},
		{
			name: "wrong apiVersion",
			args: args{
				buf: heredoc.Doc(`
					apiVersion: kubeone.k8c.io/v1beta1
					kind: ContainerRuntimeContainerd
					registries: {}
				`),
			},
			wantErr: true,
		},
		{
			name: "wrong kind",
			args: args{
				buf: heredoc.Doc(`
					apiVersion: kubeone.k8c.io/v1beta2
					kind: KubeOneCluster
					registries: {}
				`),
			},
			wantErr: true,
		},
		{
			name: "no containerd runtime",
			args: args{
				cluster: &kubeoneapi.KubeOneCluster{},
				buf: heredoc.Doc(`
					apiVersion: kubeone.k8c.io/v1beta2
					kind: ContainerRuntimeContainerd
					registries: {}
				`),
			},
			exampleCluster: &kubeoneapi.KubeOneCluster{},
			wantErr:        true,
		},
		{
			name: "borked fields (strict)",
			args: args{
				cluster: &kubeoneapi.KubeOneCluster{
					ContainerRuntime: kubeoneapi.ContainerRuntimeConfig{
						Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
					},
				},
				buf: heredoc.Doc(`
					apiVersion: kubeone.k8c.io/v1beta2
					kind: ContainerRuntimeContainerd
					registries:
					  auth:
					    some.tld:
					      username: root
				`),
			},
			exampleCluster: &kubeoneapi.KubeOneCluster{
				ContainerRuntime: kubeoneapi.ContainerRuntimeConfig{
					Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
				},
			},
			wantErr: true,
		},
		{
			name: "simple",
			args: args{
				cluster: &kubeoneapi.KubeOneCluster{
					ContainerRuntime: kubeoneapi.ContainerRuntimeConfig{
						Containerd: &kubeoneapi.ContainerRuntimeContainerd{
							Registries: map[string]kubeoneapi.ContainerdRegistry{
								"some.tld": {
									Mirrors: []string{"https://mirror1"},
								},
							},
						},
					},
				},
				buf: heredoc.Doc(`
					apiVersion: kubeone.k8c.io/v1beta2
					kind: ContainerRuntimeContainerd
					registries:
					  some.tld:
					    auth:
					      username: root
				`),
			},
			wantErr: false,
			exampleCluster: &kubeoneapi.KubeOneCluster{
				ContainerRuntime: kubeoneapi.ContainerRuntimeConfig{
					Containerd: &kubeoneapi.ContainerRuntimeContainerd{
						Registries: map[string]kubeoneapi.ContainerdRegistry{
							"some.tld": {
								Mirrors: []string{"https://mirror1"},
								Auth: &kubeoneapi.ContainerdRegistryAuthConfig{
									Username: "root",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := setRegistriesAuth(tt.args.cluster, tt.args.buf); (err != nil) != tt.wantErr {
				t.Errorf("setRegistriesAuth() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.exampleCluster, tt.args.cluster) {
				t.Errorf("example cluster doesn't correspond to resulted")
			}
		})
	}
}
