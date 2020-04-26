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

func TestCopyPKIHome(t *testing.T) {
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
		t.Run(tt.workdir, func(t *testing.T) {
			got, err := CopyPKIHome(tt.workdir)
			if err != tt.wantErr {
				t.Fatalf("CopyPKIHome() error = %v, wantErr %v", err, tt.wantErr)
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}
