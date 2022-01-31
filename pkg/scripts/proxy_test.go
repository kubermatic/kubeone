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

package scripts

import (
	"errors"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/testhelper"
)

func TestEnvironmentFile(t *testing.T) {
	t.Parallel()

	type args struct {
		cluster *kubeoneapi.KubeOneCluster
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "empty-proxy",
			args: args{cluster: &kubeoneapi.KubeOneCluster{}},
		},
		{
			name: "http-proxy",
			args: args{cluster: &kubeoneapi.KubeOneCluster{
				Proxy: kubeoneapi.ProxyConfig{
					HTTP: "http://http.proxy",
				},
			}},
		},
		{
			name: "http-https-proxy",
			args: args{cluster: &kubeoneapi.KubeOneCluster{
				Proxy: kubeoneapi.ProxyConfig{
					HTTP:  "http://http.proxy",
					HTTPS: "http://https.proxy",
				},
			}},
		},
		{
			name: "http-https-no-proxy",
			args: args{cluster: &kubeoneapi.KubeOneCluster{
				Proxy: kubeoneapi.ProxyConfig{
					HTTP:    "http://http.proxy",
					HTTPS:   "http://https.proxy",
					NoProxy: ".local",
				},
			}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := EnvironmentFile(tt.args.cluster)
			if !errors.Is(err, tt.err) {
				t.Errorf("EnvironmentFile() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestTestDaemonsProxy(t *testing.T) {
	t.Parallel()

	got, err := DaemonsEnvironmentDropIn("docker", "containerd", "kubelet")
	if err != nil {
		t.Errorf("DaemonsProxy() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}
