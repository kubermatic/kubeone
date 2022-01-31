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

package tasks

import (
	"reflect"
	"testing"
)

func Test_marshalKubeletFlags(t *testing.T) {
	type args struct {
		kubeletflags map[string]string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "simple",
			args: args{
				kubeletflags: map[string]string{
					"--key1": "val1",
					"--key2": "val2",
				},
			},
			want: []byte(`KUBELET_KUBEADM_ARGS="--key1=val1 --key2=val2"`),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := marshalKubeletFlags(tt.args.kubeletflags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("marshalKubeletFlags() = \n%s, want\n%s", got, tt.want)
			}
		})
	}
}

func Test_unmarshalKubeletFlags(t *testing.T) {
	tests := []struct {
		name    string
		buf     []byte
		want    map[string]string
		wantErr bool
	}{
		{
			name: "simple",
			buf:  []byte(`KUBELET_KUBEADM_ARGS="--key1=val1 --key2=val2"`),
			want: map[string]string{
				"--key1": "val1",
				"--key2": "val2",
			},
			wantErr: false,
		},
		{
			name: "key-values in a flag",
			buf:  []byte(`KUBELET_KUBEADM_ARGS="--key1=val1=test1,val2=test2 --key2=val2"`),
			want: map[string]string{
				"--key1": "val1=test1,val2=test2",
				"--key2": "val2",
			},
			wantErr: false,
		},
		{
			name:    "error1",
			buf:     []byte{},
			wantErr: true,
		},
		{
			name:    "error2",
			buf:     []byte(`some="key1 key2"`),
			wantErr: true,
		},
		{
			name:    "error3",
			buf:     []byte(`some="key1 key2=val2"`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalKubeletFlags(tt.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalKubeletFlags() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshalKubeletFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}
