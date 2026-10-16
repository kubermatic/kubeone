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

func TestDigitalOceanControlPlaneDefaults(t *testing.T) {
	nodeSets := []kubeoneapi.NodeSet{
		{Name: "cp", Replicas: 3, CloudProviderSpec: []byte(`{"region":"fra1","size":"s-2vcpu-4gb"}`)},
	}

	tests := []struct {
		name string
		in   kubeoneapi.DigitalOceanLoadBalancer
		want kubeoneapi.DigitalOceanLoadBalancer
	}{
		{
			name: "create mode defaults",
			in:   kubeoneapi.DigitalOceanLoadBalancer{},
			want: kubeoneapi.DigitalOceanLoadBalancer{Name: "demo-kubeapi", Region: "fra1", SizeUnit: 1},
		},
		{
			name: "user values are kept",
			in:   kubeoneapi.DigitalOceanLoadBalancer{Name: "custom", Region: "ams3", SizeUnit: 2},
			want: kubeoneapi.DigitalOceanLoadBalancer{Name: "custom", Region: "ams3", SizeUnit: 2},
		},
		{
			name: "existing mode is not defaulted",
			in:   kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x"},
			want: kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cp := &kubeoneapi.DigitalOceanControlPlane{LoadBalancer: tc.in}
			setDefaultDigitalOceanControlPlane("demo", nodeSets, cp)

			if !reflect.DeepEqual(cp.LoadBalancer, tc.want) {
				t.Errorf("got %+v, want %+v", cp.LoadBalancer, tc.want)
			}
		})
	}
}

func TestDigitalOceanManagedControlPlaneIsSupported(t *testing.T) {
	cluster := &kubeoneapi.KubeOneCluster{
		Name: "demo",
		CloudProvider: kubeoneapi.CloudProviderSpec{
			DigitalOcean: &kubeoneapi.DigitalOceanSpec{},
		},
		ControlPlane: kubeoneapi.ControlPlaneConfig{
			NodeSets: []kubeoneapi.NodeSet{
				{Name: "cp", Replicas: 1, CloudProviderSpec: []byte(`{"region":"fra1","size":"s-2vcpu-4gb"}`)},
			},
		},
	}

	if err := SetKubeOneClusterDynamicDefaults(cluster, "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cp := cluster.CloudProvider.DigitalOcean.ControlPlane
	if cp == nil || cp.LoadBalancer.Name != "demo-kubeapi" {
		t.Fatalf("expected control plane load balancer to be defaulted, got %+v", cp)
	}
}
