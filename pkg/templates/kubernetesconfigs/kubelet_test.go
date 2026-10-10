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

package kubernetesconfigs

import (
	"slices"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/templates/resources"

	metav1unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewKubeletConfigurationClusterDNS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		nodeLocalDNS bool
		ciliumLRP    bool
		expected     []string
	}{
		{
			name:         "nodeLocalDNS enabled",
			nodeLocalDNS: true,
			expected:     []string{resources.NodeLocalDNSVirtualIP},
		},
		{
			name:     "nodeLocalDNS disabled",
			expected: []string{"10.96.0.10"},
		},
		{
			name:      "cilium local redirect policy enabled",
			ciliumLRP: true,
			expected:  []string{resources.NodeLocalDNSVirtualIP, "10.96.0.10"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cluster := &kubeoneapi.KubeOneCluster{
				ClusterNetwork: kubeoneapi.ClusterNetworkConfig{
					ServiceSubnet: "10.96.0.0/12",
					CNI:           &kubeoneapi.CNI{},
				},
				Features: kubeoneapi.Features{
					NodeLocalDNS: &kubeoneapi.NodeLocalDNS{Deploy: tc.nodeLocalDNS},
				},
			}

			if tc.ciliumLRP {
				cluster.ClusterNetwork.CNI.Cilium = &kubeoneapi.CiliumSpec{EnableLocalRedirectPolicy: true}
			}

			obj, err := NewKubeletConfiguration(cluster, nil)
			if err != nil {
				t.Fatalf("unable to build kubelet configuration: %v", err)
			}

			uObj, ok := obj.(*metav1unstructured.Unstructured)
			if !ok {
				t.Fatalf("expected *Unstructured, got %T", obj)
			}

			clusterDNS, _, err := metav1unstructured.NestedStringSlice(uObj.Object, "clusterDNS")
			if err != nil {
				t.Fatalf("unable to read clusterDNS: %v", err)
			}

			if !slices.Equal(clusterDNS, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, clusterDNS)
			}
		})
	}
}
