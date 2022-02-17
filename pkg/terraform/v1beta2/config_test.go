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

package v1beta2

import (
	"reflect"
	"testing"

	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
)

func Test_mergeHosts(t *testing.T) {
	type args struct {
		dst []kubeonev1beta2.HostConfig
		src []kubeonev1beta2.HostConfig
	}
	tests := []struct {
		name    string
		wantErr bool
		args    args
		sample  []kubeonev1beta2.HostConfig
	}{
		{
			name: "simple",
			args: args{
				dst: []kubeonev1beta2.HostConfig{
					{
						PrivateAddress: "1.1.1.1",
						Hostname:       "ololo1",
					},
				},
				src: []kubeonev1beta2.HostConfig{
					{
						PrivateAddress: "1.1.1.1",
						SSHPort:        2222,
					},
				},
			},
			sample: []kubeonev1beta2.HostConfig{
				{
					PrivateAddress: "1.1.1.1",
					Hostname:       "ololo1",
					SSHPort:        2222,
				},
			},
		},
		{
			name: "completely new",
			args: args{
				dst: []kubeonev1beta2.HostConfig{},
				src: []kubeonev1beta2.HostConfig{
					{
						PrivateAddress: "2.2.2.2",
						Hostname:       "completely-new",
						SSHPort:        2222,
					},
				},
			},
			sample: []kubeonev1beta2.HostConfig{
				{
					PrivateAddress: "2.2.2.2",
					Hostname:       "completely-new",
					SSHPort:        2222,
				},
			},
		},
		{
			name: "empty",
			args: args{
				dst: []kubeonev1beta2.HostConfig{},
				src: []kubeonev1beta2.HostConfig{},
			},
			sample: []kubeonev1beta2.HostConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mergeHosts(tt.args.dst, tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeHosts() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(result, tt.sample) {
				t.Logf("%#v", result)
				t.Errorf("result doesn't match the sample")
			}
		})
	}
}
