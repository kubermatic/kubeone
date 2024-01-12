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

package kubeone

import (
	"reflect"
	"testing"
)

func TestFeatureGatesString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		featureGates map[string]bool
		expected     string
	}{
		{
			name:         "one feature gate",
			featureGates: map[string]bool{"TestFeatureGate": true},
			expected:     "TestFeatureGate=true",
		},
		{
			name: "two feature gates",
			featureGates: map[string]bool{
				"TestFeatureGate":  true,
				"TestDisabledGate": false,
			},
			expected: "TestDisabledGate=false,TestFeatureGate=true",
		},
		{
			name: "three feature gates",
			featureGates: map[string]bool{
				"TestFeatureGate":  true,
				"TestDisabledGate": false,
				"TestThirdGate":    true,
			},
			expected: "TestDisabledGate=false,TestFeatureGate=true,TestThirdGate=true",
		},
		{
			name:         "no feature gates",
			featureGates: map[string]bool{},
			expected:     "",
		},
		{
			name:         "feature gates nil",
			featureGates: nil,
			expected:     "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := marshalFeatureGates(tc.featureGates)
			if got != tc.expected {
				t.Errorf("TestFeatureGatesString() got = %v, expected %v", got, tc.expected)
			}
		})
	}
}

func TestMapStringStringToString(t *testing.T) {
	tests := []struct {
		name string
		m1   map[string]string
		want string
	}{
		{
			m1: map[string]string{
				"k2": "v2",
				"k3": "v3",
				"k1": "v1",
			},
			want: "k1=v1,k2=v2,k3=v3",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := MapStringStringToString(tt.m1, "="); got != tt.want {
				t.Errorf("MapStringStringToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultAssetConfiguration(t *testing.T) {
	tests := []struct {
		name                       string
		cluster                    *KubeOneCluster
		expectedAssetConfiguration AssetConfiguration
	}{
		{
			name:                       "default options",
			cluster:                    &KubeOneCluster{},
			expectedAssetConfiguration: AssetConfiguration{},
		},
		{
			name: "overwriteRegistry",
			cluster: &KubeOneCluster{
				RegistryConfiguration: &RegistryConfiguration{
					OverwriteRegistry: "my.corp",
				},
			},
			expectedAssetConfiguration: AssetConfiguration{
				Kubernetes: ImageAsset{
					ImageRepository: "my.corp",
				},
				CoreDNS: ImageAsset{
					ImageRepository: "my.corp",
				},
				Etcd: ImageAsset{
					ImageRepository: "my.corp",
				},
				MetricsServer: ImageAsset{
					ImageRepository: "my.corp",
				},
			},
		},
		{
			name: "coredns over assetConfiguration (v1beta1)",
			cluster: &KubeOneCluster{
				RegistryConfiguration: &RegistryConfiguration{
					OverwriteRegistry: "my.corp",
				},
				AssetConfiguration: AssetConfiguration{
					CoreDNS: ImageAsset{
						ImageRepository: "my.corp/coredns",
					},
				},
			},
			expectedAssetConfiguration: AssetConfiguration{
				Kubernetes: ImageAsset{
					ImageRepository: "my.corp",
				},
				CoreDNS: ImageAsset{
					ImageRepository: "my.corp/coredns",
				},
				Etcd: ImageAsset{
					ImageRepository: "my.corp",
				},
				MetricsServer: ImageAsset{
					ImageRepository: "my.corp",
				},
			},
		},
		{
			name: "coredns over coreDNS Feature (v1beta2)",
			cluster: &KubeOneCluster{
				RegistryConfiguration: &RegistryConfiguration{
					OverwriteRegistry: "my.corp",
				},
				Features: Features{
					CoreDNS: &CoreDNS{
						ImageRepository: "my.corp/coredns",
					},
				},
			},
			expectedAssetConfiguration: AssetConfiguration{
				Kubernetes: ImageAsset{
					ImageRepository: "my.corp",
				},
				CoreDNS: ImageAsset{
					ImageRepository: "my.corp/coredns",
				},
				Etcd: ImageAsset{
					ImageRepository: "my.corp",
				},
				MetricsServer: ImageAsset{
					ImageRepository: "my.corp",
				},
			},
		},
		{
			name: "coredns over coreDNS Feature without overwriteRegistry (v1beta2)",
			cluster: &KubeOneCluster{
				Features: Features{
					CoreDNS: &CoreDNS{
						ImageRepository: "my.corp/coredns",
					},
				},
			},
			expectedAssetConfiguration: AssetConfiguration{
				CoreDNS: ImageAsset{
					ImageRepository: "my.corp/coredns",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.cluster.DefaultAssetConfiguration()
			if !reflect.DeepEqual(tt.cluster.AssetConfiguration, tt.expectedAssetConfiguration) {
				t.Errorf("Expected AssetConfiguration=%v, but got=%v", tt.expectedAssetConfiguration, tt.cluster.AssetConfiguration)
			}
		})
	}
}
