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

	"k8c.io/kubeone/pkg/testhelper"
)

func TestKubeadmJoin(t *testing.T) {
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
			t.Parallel()
			got, err := KubeadmJoin(tt.args.workdir, tt.args.nodeID, tt.args.verboseFlag)
			if !errors.Is(err, tt.err) {
				t.Errorf("KubeadmJoin() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestKubeadmJoinWorker(t *testing.T) {
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
			t.Parallel()
			got, err := KubeadmJoinWorker(tt.args.workdir, tt.args.nodeID, tt.args.verboseFlag)
			if !errors.Is(err, tt.err) {
				t.Errorf("KubeadmJoinWorker() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestKubeadmCert(t *testing.T) {
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
			t.Parallel()
			got, err := KubeadmCert(tt.args.workdir, tt.args.nodeID, tt.args.verboseFlag)
			if !errors.Is(err, tt.err) {
				t.Errorf("KubeadmCert() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestKubeadmInit(t *testing.T) {
	t.Parallel()

	type args struct {
		workdir     string
		nodeID      int
		verboseFlag string
		token       string
		tokenTTL    string
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
				token:       "123098",
				tokenTTL:    "1h",
			},
		},
		{
			name: "not-verbose",
			args: args{
				workdir:  "test-wd",
				token:    "123098",
				tokenTTL: "1h",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := KubeadmInit(tt.args.workdir, tt.args.nodeID, tt.args.verboseFlag, tt.args.token, tt.args.tokenTTL, "")
			if !errors.Is(err, tt.err) {
				t.Errorf("KubeadmInit() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestKubeadmReset(t *testing.T) {
	t.Parallel()

	type args struct {
		verboseFlag string
		workdir     string
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
			t.Parallel()
			got, err := KubeadmReset(tt.args.verboseFlag, tt.args.workdir)
			if !errors.Is(err, tt.err) {
				t.Errorf("KubeadmReset() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestKubeadmUpgrade(t *testing.T) {
	t.Parallel()

	type args struct {
		kubeadmCmd string
		workdir    string
		leader     bool
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "v1beta2",
			args: args{
				workdir:    "test-wd",
				kubeadmCmd: "kubeadm upgrade node",
			},
		},
		{
			name: "leader",
			args: args{
				workdir:    "some",
				kubeadmCmd: "kubeadm upgrade apply -y --certificate-renewal=true v1.1.1",
				leader:     true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := KubeadmUpgrade(tt.args.kubeadmCmd, tt.args.workdir, tt.args.leader, 0)
			if !errors.Is(err, tt.err) {
				t.Errorf("KubeadmUpgradeLeader() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}
