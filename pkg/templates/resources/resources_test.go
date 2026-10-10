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

package resources

import (
	"slices"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

func TestClusterDNSIPs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		network      kubeoneapi.ClusterNetworkConfig
		nodeLocalDNS bool
		expected     []string
	}{
		{
			name:         "nodeLocalDNS enabled",
			network:      kubeoneapi.ClusterNetworkConfig{ServiceSubnet: "10.96.0.0/12", CNI: &kubeoneapi.CNI{}},
			nodeLocalDNS: true,
			expected:     []string{NodeLocalDNSVirtualIP},
		},
		{
			name:     "nodeLocalDNS disabled",
			network:  kubeoneapi.ClusterNetworkConfig{ServiceSubnet: "10.224.0.0/12", CNI: &kubeoneapi.CNI{}},
			expected: []string{"10.224.0.10"},
		},
		{
			name: "cilium local redirect policy enabled",
			network: kubeoneapi.ClusterNetworkConfig{
				ServiceSubnet: "10.96.0.0/12",
				CNI:           &kubeoneapi.CNI{Cilium: &kubeoneapi.CiliumSpec{EnableLocalRedirectPolicy: true}},
			},
			expected: []string{NodeLocalDNSVirtualIP, "10.96.0.10"},
		},
		{
			name: "clusterDNS override",
			network: kubeoneapi.ClusterNetworkConfig{
				ServiceSubnet: "10.96.0.0/12",
				ClusterDNS:    []string{NodeLocalDNSVirtualIP, "10.96.0.10"},
				CNI:           &kubeoneapi.CNI{},
			},
			expected: []string{NodeLocalDNSVirtualIP, "10.96.0.10"},
		},
		{
			name: "clusterDNS override takes precedence over cilium local redirect policy",
			network: kubeoneapi.ClusterNetworkConfig{
				ServiceSubnet: "10.96.0.0/12",
				ClusterDNS:    []string{"10.0.0.53"},
				CNI:           &kubeoneapi.CNI{Cilium: &kubeoneapi.CiliumSpec{EnableLocalRedirectPolicy: true}},
			},
			expected: []string{"10.0.0.53"},
		},
		{
			name: "IPv6-only",
			network: kubeoneapi.ClusterNetworkConfig{
				IPFamily:          kubeoneapi.IPFamilyIPv6,
				ServiceSubnetIPv6: "fd02::/120",
				CNI:               &kubeoneapi.CNI{},
			},
			expected: []string{"fd02::a"},
		},
		{
			name: "dual-stack IPv4 primary",
			network: kubeoneapi.ClusterNetworkConfig{
				IPFamily:          kubeoneapi.IPFamilyIPv4IPv6,
				ServiceSubnet:     "10.96.0.0/12",
				ServiceSubnetIPv6: "fd02::/120",
				CNI:               &kubeoneapi.CNI{},
			},
			expected: []string{"10.96.0.10"},
		},
		{
			name: "dual-stack IPv6 primary",
			network: kubeoneapi.ClusterNetworkConfig{
				IPFamily:          kubeoneapi.IPFamilyIPv6IPv4,
				ServiceSubnet:     "10.96.0.0/12",
				ServiceSubnetIPv6: "fd02::/120",
				CNI:               &kubeoneapi.CNI{},
			},
			expected: []string{"fd02::a"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cluster := &kubeoneapi.KubeOneCluster{
				ClusterNetwork: tc.network,
				Features: kubeoneapi.Features{
					NodeLocalDNS: &kubeoneapi.NodeLocalDNS{Deploy: tc.nodeLocalDNS},
				},
			}

			if got := ClusterDNSIPs(cluster); !slices.Equal(got, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
