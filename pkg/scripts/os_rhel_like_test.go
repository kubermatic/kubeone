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

import (
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/testhelper"
)

func TestRHELLikeScript(t *testing.T) {
	tests := []struct {
		name    string
		cluster kubeoneapi.KubeOneCluster
		params  Params
		wantErr bool
	}{
		{
			name:    "install all",
			cluster: genCluster(),
			params:  Params{Kubeadm: true, Kubectl: true, Kubelet: true},
		},
		{
			name:    "install all with force",
			cluster: genCluster(),
			params:  Params{Force: true, Kubeadm: true, Kubectl: true, Kubelet: true},
		},
		{
			name:    "install all with proxy",
			cluster: genCluster(withProxy("http://proxy.tld")),
			params:  Params{Kubeadm: true, Kubectl: true, Kubelet: true},
		},
		{
			name:    "upgrade kubeadm and kubectl",
			cluster: genCluster(),
			params:  Params{Upgrade: true, Kubeadm: true, Kubectl: true},
		},
		{
			name:    "upgrade kubeadm and kubectl with force",
			cluster: genCluster(),
			params:  Params{Force: true, Upgrade: true, Kubeadm: true, Kubectl: true},
		},
		{
			name:    "upgrade kubelet",
			cluster: genCluster(),
			params:  Params{Upgrade: true, Kubelet: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RHELLikeScript(&tt.cluster, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("RHELLikeScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestRemoveBinariesRHELLike(t *testing.T) {
	t.Parallel()

	got, err := RemoveBinariesRHELLike()
	if err != nil {
		t.Errorf("RemoveBinariesRHELLike() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}
