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
	"os"
	"path/filepath"
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
		name     string
		in       kubeoneapi.DigitalOceanLoadBalancer
		nodeSets []kubeoneapi.NodeSet
		want     kubeoneapi.DigitalOceanLoadBalancer
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
			name: "existing mode only gets the region",
			in:   kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x"},
			want: kubeoneapi.DigitalOceanLoadBalancer{ID: "lb-x", Region: "fra1"},
		},
		{
			name:     "region given as ConfigVar object",
			in:       kubeoneapi.DigitalOceanLoadBalancer{},
			nodeSets: []kubeoneapi.NodeSet{{Name: "cp", Replicas: 1, CloudProviderSpec: []byte(`{"region":{"value":"ams3"},"size":"s-2vcpu-4gb"}`)}},
			want:     kubeoneapi.DigitalOceanLoadBalancer{Name: "demo-kubeapi", Region: "ams3", SizeUnit: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cp := &kubeoneapi.DigitalOceanControlPlane{LoadBalancer: tc.in}
			ns := nodeSets
			if tc.nodeSets != nil {
				ns = tc.nodeSets
			}
			setDefaultDigitalOceanControlPlane("demo", ns, cp)

			if !reflect.DeepEqual(cp.LoadBalancer, tc.want) {
				t.Errorf("got %+v, want %+v", cp.LoadBalancer, tc.want)
			}
		})
	}
}

func TestHetznerControlPlaneDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   kubeoneapi.HetznerLoadBalancer
		want kubeoneapi.HetznerLoadBalancer
	}{
		{
			name: "defaults",
			in:   kubeoneapi.HetznerLoadBalancer{},
			want: kubeoneapi.HetznerLoadBalancer{
				Name:     "demo-kubeapi",
				Type:     "lb11",
				Location: "nbg1",
				PublicIP: new(true),
				Labels:   map[string]string{"kubeone_cluster_name": "demo", "kubeone_role": "api"},
			},
		},
		{
			name: "cluster name label is the cluster name, not the load balancer name",
			in: kubeoneapi.HetznerLoadBalancer{
				Name:     "custom",
				Type:     "lb21",
				Location: "fsn1",
				PublicIP: new(false),
				Labels:   map[string]string{"kubeone_cluster_name": "other", "team": "infra"},
			},
			want: kubeoneapi.HetznerLoadBalancer{
				Name:     "custom",
				Type:     "lb21",
				Location: "fsn1",
				PublicIP: new(false),
				Labels:   map[string]string{"kubeone_cluster_name": "demo", "kubeone_role": "api", "team": "infra"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cp := &kubeoneapi.HetznerControlPlane{LoadBalancer: tc.in}
			setDefaultHetznerControlPlane("demo", cp)

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
			DigitalOcean: &kubeoneapi.DigitalOceanSpec{
				ControlPlane: &kubeoneapi.DigitalOceanControlPlane{},
			},
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

func TestGCEControlPlaneDefaults(t *testing.T) {
	nodeSets := []kubeoneapi.NodeSet{
		{Name: "cp", Replicas: 3, CloudProviderSpec: []byte(`{"zone":"europe-west3-a","machineType":"e2-standard-2"}`)},
	}

	tests := []struct {
		name     string
		in       kubeoneapi.GCELoadBalancer
		nodeSets []kubeoneapi.NodeSet
		want     kubeoneapi.GCELoadBalancer
	}{
		{
			name: "defaults",
			in:   kubeoneapi.GCELoadBalancer{},
			want: kubeoneapi.GCELoadBalancer{
				Name:                "demo-kubeapi",
				Region:              "europe-west3",
				AllowedSourceRanges: []string{"0.0.0.0/0"},
			},
		},
		{
			name: "user values are kept",
			in: kubeoneapi.GCELoadBalancer{
				ProjectID:           "proj",
				Name:                "custom",
				Region:              "us-central1",
				Network:             "global/networks/vpc",
				AllowedSourceRanges: []string{"10.0.0.0/8"},
			},
			want: kubeoneapi.GCELoadBalancer{
				ProjectID:           "proj",
				Name:                "custom",
				Region:              "us-central1",
				Network:             "global/networks/vpc",
				AllowedSourceRanges: []string{"10.0.0.0/8"},
			},
		},
		{
			name:     "zone given as ConfigVar object",
			in:       kubeoneapi.GCELoadBalancer{},
			nodeSets: []kubeoneapi.NodeSet{{Name: "cp", Replicas: 1, CloudProviderSpec: []byte(`{"zone":{"value":"us-east1-b"}}`)}},
			want: kubeoneapi.GCELoadBalancer{
				Name:                "demo-kubeapi",
				Region:              "us-east1",
				AllowedSourceRanges: []string{"0.0.0.0/0"},
			},
		},
		{
			// The network is resolved from all nodeSets when the load balancer is ensured
			name:     "network of the nodeSet isn't defaulted",
			in:       kubeoneapi.GCELoadBalancer{},
			nodeSets: []kubeoneapi.NodeSet{{Name: "cp", Replicas: 1, CloudProviderSpec: []byte(`{"zone":"europe-west3-a","network":"my-vpc"}`)}},
			want: kubeoneapi.GCELoadBalancer{
				Name:                "demo-kubeapi",
				Region:              "europe-west3",
				AllowedSourceRanges: []string{"0.0.0.0/0"},
			},
		},
		{
			name:     "only a subnetwork",
			in:       kubeoneapi.GCELoadBalancer{},
			nodeSets: []kubeoneapi.NodeSet{{Name: "cp", Replicas: 1, CloudProviderSpec: []byte(`{"zone":"europe-west3-a","subnetwork":"my-subnet"}`)}},
			want: kubeoneapi.GCELoadBalancer{
				Name:                "demo-kubeapi",
				Region:              "europe-west3",
				AllowedSourceRanges: []string{"0.0.0.0/0"},
			},
		},
		{
			name:     "no zone, no region",
			in:       kubeoneapi.GCELoadBalancer{},
			nodeSets: []kubeoneapi.NodeSet{{Name: "cp", Replicas: 1, CloudProviderSpec: []byte(`{}`)}},
			want: kubeoneapi.GCELoadBalancer{
				Name:                "demo-kubeapi",
				AllowedSourceRanges: []string{"0.0.0.0/0"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cp := &kubeoneapi.GCEControlPlane{LoadBalancer: tc.in}
			ns := nodeSets
			if tc.nodeSets != nil {
				ns = tc.nodeSets
			}
			setDefaultGCEControlPlane("demo", ns, cp)

			if !reflect.DeepEqual(cp.LoadBalancer, tc.want) {
				t.Errorf("got %+v, want %+v", cp.LoadBalancer, tc.want)
			}
		})
	}
}

func TestGCEManagedControlPlaneIsSupported(t *testing.T) {
	cluster := &kubeoneapi.KubeOneCluster{
		Name: "demo",
		CloudProvider: kubeoneapi.CloudProviderSpec{
			GCE: &kubeoneapi.GCESpec{
				ControlPlane: &kubeoneapi.GCEControlPlane{},
			},
		},
		ControlPlane: kubeoneapi.ControlPlaneConfig{
			NodeSets: []kubeoneapi.NodeSet{
				{Name: "cp", Replicas: 1, CloudProviderSpec: []byte(`{"zone":"europe-west3-a"}`)},
			},
		},
	}

	if err := SetKubeOneClusterDynamicDefaults(cluster, "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lb := cluster.CloudProvider.GCE.ControlPlane.LoadBalancer
	if lb.Name != "demo-kubeapi" || lb.Region != "europe-west3" {
		t.Fatalf("expected control plane load balancer to be defaulted, got %+v", lb)
	}
}

func TestValidateAuthenticationConfigurationFile(t *testing.T) {
	baseDir := t.TempDir()

	files := map[string]string{
		"valid.yaml": heredoc.Doc(`
			apiVersion: apiserver.config.k8s.io/v1
			kind: AuthenticationConfiguration
			jwt:
			- issuer:
			    url: https://issuer.example.com
			    audiences:
			    - kubernetes
			  claimMappings:
			    username:
			      claim: sub
			      prefix: "oidc:"
		`),
		"malformed.yaml": "apiVersion: apiserver.config.k8s.io/v1\nkind: AuthenticationConfiguration\njwt: [\n",
		"unknown-field.yaml": heredoc.Doc(`
			apiVersion: apiserver.config.k8s.io/v1
			kind: AuthenticationConfiguration
			jwts: []
		`),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(baseDir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name    string
		feature *kubeoneapi.AuthenticationConfiguration
		wantErr bool
	}{
		{
			name: "feature not configured",
		},
		{
			name:    "feature disabled",
			feature: &kubeoneapi.AuthenticationConfiguration{ConfigFilePath: "missing.yaml"},
		},
		{
			name:    "relative path",
			feature: &kubeoneapi.AuthenticationConfiguration{Enable: true, ConfigFilePath: "valid.yaml"},
		},
		{
			name:    "absolute path",
			feature: &kubeoneapi.AuthenticationConfiguration{Enable: true, ConfigFilePath: filepath.Join(baseDir, "valid.yaml")},
		},
		{
			name:    "missing file",
			feature: &kubeoneapi.AuthenticationConfiguration{Enable: true, ConfigFilePath: "missing.yaml"},
			wantErr: true,
		},
		{
			name:    "malformed yaml",
			feature: &kubeoneapi.AuthenticationConfiguration{Enable: true, ConfigFilePath: "malformed.yaml"},
			wantErr: true,
		},
		{
			name:    "unknown field",
			feature: &kubeoneapi.AuthenticationConfiguration{Enable: true, ConfigFilePath: "unknown-field.yaml"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := &kubeoneapi.KubeOneCluster{
				ClusterNetwork: kubeoneapi.ClusterNetworkConfig{ServiceDomainName: "cluster.local"},
				Features:       kubeoneapi.Features{AuthenticationConfiguration: tt.feature},
			}

			err := validateAuthenticationConfigurationFile(cluster, baseDir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateAuthenticationConfigurationFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
