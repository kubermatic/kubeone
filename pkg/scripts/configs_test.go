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

	"github.com/kubermatic/kubeone/pkg/testhelper"
)

func TestKubernetesAdminConfig(t *testing.T) {
	t.Parallel()

	got, err := KubernetesAdminConfig()
	if err != nil {
		t.Fatalf("KubernetesAdminConfig() error = %v", err)
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestSaveCloudConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		workdir string
		wantErr error
	}{
		{name: "kubeone1", workdir: "test-dir1"},
		{name: "kubeone2", workdir: "./subdir/test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SaveCloudConfig(tt.workdir)
			if err != tt.wantErr {
				t.Errorf("SaveCloudConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestSaveAuditPolicyConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		workdir string
		wantErr error
	}{
		{name: "kubeone1", workdir: "test-dir1"},
		{name: "kubeone2", workdir: "./subdir/test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SaveAuditPolicyConfig(tt.workdir)
			if err != tt.wantErr {
				t.Errorf("SaveAuditPolicyConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}
