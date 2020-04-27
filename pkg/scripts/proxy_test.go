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
	"testing"

	"github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/testhelper"
)

func TestEnvironmentFile(t *testing.T) {
	t.Parallel()

	type args struct {
		cluster *kubeone.KubeOneCluster
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "empty-proxy",
			args: args{cluster: &kubeone.KubeOneCluster{}},
		},
		{
			name: "http-proxy",
			args: args{cluster: &kubeone.KubeOneCluster{
				Proxy: kubeone.ProxyConfig{
					HTTP: "http://http.proxy",
				},
			}},
		},
		{
			name: "http-https-proxy",
			args: args{cluster: &kubeone.KubeOneCluster{
				Proxy: kubeone.ProxyConfig{
					HTTP:  "http://http.proxy",
					HTTPS: "http://https.proxy",
				},
			}},
		},
		{
			name: "http-https-no-proxy",
			args: args{cluster: &kubeone.KubeOneCluster{
				Proxy: kubeone.ProxyConfig{
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
			got, err := EnvironmentFile(tt.args.cluster)
			if err != tt.err {
				t.Errorf("EnvironmentFile() error = %v, wantErr %v", err, tt.err)
				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestDaemonsProxy(t *testing.T) {
	t.Parallel()

	got, err := DaemonsProxy()
	if err != nil {
		t.Errorf("DaemonsProxy() error = %v", err)
		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}
