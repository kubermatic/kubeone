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

	"k8c.io/kubeone/pkg/testhelper"
)

func TestCCMMigrationRegenerateControlPlaneManifests(t *testing.T) {
	t.Parallel()

	type args struct {
		workdir     string
		nodeID      int
		verboseFlag string
	}

	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "verbose",
			args: args{
				workdir:     "test-wd",
				nodeID:      0,
				verboseFlag: "--v=6",
			},
		},
		{
			name: "not-verbose",
			args: args{
				workdir: "test-wd",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := CCMMigrationRegenerateControlPlaneManifests(tt.args.workdir, tt.args.nodeID, tt.args.verboseFlag)
			if err != tt.err {
				t.Errorf("TestCCMMigrationRegenerateControlPlaneManifests() error = %v, wantErr %v", err, tt.err)
				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestCCMMigrationUpdateKubeletConfig(t *testing.T) {
	t.Parallel()

	type args struct {
		workdir     string
		nodeID      int
		verboseFlag string
	}

	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "verbose",
			args: args{
				workdir:     "test-wd",
				nodeID:      0,
				verboseFlag: "--v=6",
			},
		},
		{
			name: "not-verbose",
			args: args{
				workdir: "test-wd",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := CCMMigrationUpdateKubeletConfig(tt.args.workdir, tt.args.nodeID, tt.args.verboseFlag)
			if err != tt.err {
				t.Errorf("CCMMigrationUpdateKubeletConfig() error = %v, wantErr %v", err, tt.err)
				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}
