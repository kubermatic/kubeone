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

package provisioner

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	cloud "k8c.io/machine-controller/pkg/cloudprovider/instance"
)

// fakeInstance implements cloud.Instance for testing.
type fakeInstance struct {
	addresses map[string]corev1.NodeAddressType
}

func (f *fakeInstance) Name() string                                 { return "fake" }
func (f *fakeInstance) ID() string                                   { return "fake-id" }
func (f *fakeInstance) ProviderID() string                           { return "fake://fake-id" }
func (f *fakeInstance) Addresses() map[string]corev1.NodeAddressType { return f.addresses }
func (f *fakeInstance) Status() cloud.Status                         { return cloud.StatusRunning }

func Test_GetMachineInfo(t *testing.T) {
	tests := []struct {
		name     string
		instance cloud.Instance
		want     Machine
	}{
		{
			name: "ipv4 public and private",
			instance: &fakeInstance{addresses: map[string]corev1.NodeAddressType{
				"1.2.3.4":  corev1.NodeExternalIP,
				"10.0.0.1": corev1.NodeInternalIP,
				"my-host":  corev1.NodeHostName,
			}},
			want: Machine{PublicAddress: "1.2.3.4", PrivateAddress: "10.0.0.1", Hostname: "my-host"},
		},
		{
			name: "ipv6 only falls back correctly",
			instance: &fakeInstance{addresses: map[string]corev1.NodeAddressType{
				"2001:db8::1": corev1.NodeExternalIP,
				"fd00::1":     corev1.NodeInternalIP,
			}},
			want: Machine{PublicAddress: "2001:db8::1", PrivateAddress: "fd00::1"},
		},
		{
			name: "ipv4 preferred over ipv6",
			instance: &fakeInstance{addresses: map[string]corev1.NodeAddressType{
				"2001:db8::1": corev1.NodeExternalIP,
				"5.6.7.8":     corev1.NodeExternalIP,
				"fd00::1":     corev1.NodeInternalIP,
				"192.168.1.1": corev1.NodeInternalIP,
			}},
			want: Machine{PublicAddress: "5.6.7.8", PrivateAddress: "192.168.1.1"},
		},
		{
			name: "hostname from NodeInternalDNS when NodeHostName absent",
			instance: &fakeInstance{addresses: map[string]corev1.NodeAddressType{
				"10.0.0.5":         corev1.NodeInternalIP,
				"node.example.com": corev1.NodeInternalDNS,
			}},
			want: Machine{PrivateAddress: "10.0.0.5", Hostname: "node.example.com"},
		},
		{
			name: "hostname from NodeExternalDNS when others absent",
			instance: &fakeInstance{addresses: map[string]corev1.NodeAddressType{
				"ext.example.com": corev1.NodeExternalDNS,
			}},
			want: Machine{Hostname: "ext.example.com"},
		},
		{
			name:     "empty addresses",
			instance: &fakeInstance{addresses: map[string]corev1.NodeAddressType{}},
			want:     Machine{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMachineInfo(tt.instance)
			if got != tt.want {
				t.Errorf("getMachineInfo() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
