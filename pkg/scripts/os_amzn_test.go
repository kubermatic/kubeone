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

func TestAmazonLinuxScript(t *testing.T) {
	type args struct {
		cluster *kubeoneapi.KubeOneCluster
		params  Params
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AmazonLinuxScript(tt.args.cluster, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("AmazonLinuxScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AmazonLinuxScript() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveBinariesAmazonLinux(t *testing.T) {
	t.Parallel()

	got, err := RemoveBinariesAmazonLinux()
	if err != nil {
		t.Errorf("RemoveBinariesAmazonLinux() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}
