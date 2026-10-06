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

package machinecontroller

import (
	"encoding/json"
	"reflect"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

func TestMachineSpec(t *testing.T) {
	tests := []struct {
		name     string
		provider kubeoneapi.CloudProviderSpec
		spec     string
		want     string
		wantErr  bool
	}{
		{
			name:     "aws cluster tag is added and other fields are kept",
			provider: kubeoneapi.CloudProviderSpec{AWS: &kubeoneapi.AWSSpec{}},
			spec:     `{"region":"eu-west-1","accessKeyId":{"secretKeyRef":{"name":"aws","key":"id"}},"tags":{"team":"a"}}`,
			want:     `{"region":"eu-west-1","accessKeyId":{"secretKeyRef":{"name":"aws","key":"id"}},"tags":{"team":"a","kubernetes.io/cluster/test":"shared"}}`,
		},
		{
			name:     "aws cluster tag is added without existing tags",
			provider: kubeoneapi.CloudProviderSpec{AWS: &kubeoneapi.AWSSpec{}},
			spec:     `{"region":"eu-west-1"}`,
			want:     `{"region":"eu-west-1","tags":{"kubernetes.io/cluster/test":"shared"}}`,
		},
		{
			name:     "aws tags of a wrong type are rejected",
			provider: kubeoneapi.CloudProviderSpec{AWS: &kubeoneapi.AWSSpec{}},
			spec:     `{"tags":["team"]}`,
			wantErr:  true,
		},
		{
			name:     "aws tags under a differently cased key are merged into tags",
			provider: kubeoneapi.CloudProviderSpec{AWS: &kubeoneapi.AWSSpec{}},
			spec:     `{"Tags":{"team":"a"}}`,
			want:     `{"tags":{"team":"a","kubernetes.io/cluster/test":"shared"}}`,
		},
		{
			name:     "aws tags with non-string values are rejected",
			provider: kubeoneapi.CloudProviderSpec{AWS: &kubeoneapi.AWSSpec{}},
			spec:     `{"tags":{"team":1}}`,
			wantErr:  true,
		},
		{
			name:     "aws mistyped values are rejected",
			provider: kubeoneapi.CloudProviderSpec{AWS: &kubeoneapi.AWSSpec{}},
			spec:     `{"diskSize":"large"}`,
			wantErr:  true,
		},
		{
			name:     "aws unknown fields are rejected",
			provider: kubeoneapi.CloudProviderSpec{AWS: &kubeoneapi.AWSSpec{}},
			spec:     `{"diskSiz":50}`,
			wantErr:  true,
		},
		{
			name:     "non-aws spec is passed through",
			provider: kubeoneapi.CloudProviderSpec{Hetzner: &kubeoneapi.HetznerSpec{}},
			spec:     `{"serverType":"cx22","tags":{"team":"a"}}`,
			want:     `{"serverType":"cx22","tags":{"team":"a"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := &kubeoneapi.KubeOneCluster{Name: "test"}
			workerset := kubeoneapi.DynamicWorkerConfig{
				Config: kubeoneapi.ProviderSpec{CloudProviderSpec: json.RawMessage(tt.spec)},
			}

			got, err := machineSpec(cluster, workerset, tt.provider)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			var want map[string]any
			if err = json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}
