/*
Copyright 2025 The KubeOne Authors.

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

package scripts

import "testing"

func TestParams_String(t *testing.T) {
	tests := []struct {
		params Params
		want   string
	}{
		{
			want: "installing nothing",
		},
		{
			want:   "upgrading nothing",
			params: Params{Upgrade: true},
		},
		{
			params: Params{Kubelet: true},
			want:   "installing kubelet",
		},
		{
			params: Params{Kubelet: true, Kubectl: true},
			want:   "installing kubectl and kubelet",
		},
		{
			params: Params{Upgrade: true, Kubelet: true, Kubeadm: true, Force: true},
			want:   "upgrading kubeadm and kubelet using force",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.params.String(); got != tt.want {
				t.Errorf("Params.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
