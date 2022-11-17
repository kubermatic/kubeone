/*
Copyright 2022 The KubeOne Authors.

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

package v1beta3

import (
	"reflect"
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestEtcdVersionCorruptCheckExtraArgs(t *testing.T) {
	etcdExtraArgs := map[string]string{
		"experimental-initial-corrupt-check": "true",
		"experimental-corrupt-check-time":    "240m",
	}

	tests := []struct {
		name                 string
		kubeVersion          *semver.Version
		etcdImageTag         string
		expectedEtcdImageTag string
		expectedEtcdArgs     map[string]string
	}{
		{
			name:                 "unfixed 1.23",
			kubeVersion:          semver.MustParse("1.23.13"),
			expectedEtcdImageTag: fixedEtcdVersion,
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "unfixed 1.24",
			kubeVersion:          semver.MustParse("1.24.7"),
			expectedEtcdImageTag: fixedEtcdVersion,
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "unfixed 1.25",
			kubeVersion:          semver.MustParse("1.25.3"),
			expectedEtcdImageTag: fixedEtcdVersion,
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "fixed 1.23",
			kubeVersion:          semver.MustParse("1.23.14"),
			expectedEtcdImageTag: "",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "fixed 1.24",
			kubeVersion:          semver.MustParse("1.24.8"),
			expectedEtcdImageTag: "",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "fixed 1.25",
			kubeVersion:          semver.MustParse("1.25.4"),
			expectedEtcdImageTag: "",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "fixed 1.26",
			kubeVersion:          semver.MustParse("1.26.0"),
			expectedEtcdImageTag: "",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "unfixed 1.25, but tag is overwritten",
			kubeVersion:          semver.MustParse("1.25.3"),
			etcdImageTag:         "9.9.9-0",
			expectedEtcdImageTag: "9.9.9-0",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "fixed 1.25, but tag is overwritten",
			kubeVersion:          semver.MustParse("1.25.4"),
			etcdImageTag:         "9.9.9-0",
			expectedEtcdImageTag: "9.9.9-0",
			expectedEtcdArgs:     etcdExtraArgs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, args := etcdVersionCorruptCheckExtraArgs(tt.kubeVersion, tt.etcdImageTag)
			if ver != tt.expectedEtcdImageTag {
				t.Errorf("got etcd image tag %q, but expected %q", ver, tt.expectedEtcdImageTag)
			}
			if !reflect.DeepEqual(args, tt.expectedEtcdArgs) {
				t.Errorf("got etcd tags %q, but expected %q", args, tt.expectedEtcdArgs)
			}
		})
	}
}
