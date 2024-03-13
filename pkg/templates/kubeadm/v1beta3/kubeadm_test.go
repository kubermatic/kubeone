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
	"maps"
	"reflect"
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestEtcdVersionCorruptCheckExtraArgs(t *testing.T) {
	etcdExtraArgs := map[string]string{
		"experimental-compact-hash-check-enabled": "true",
		"experimental-initial-corrupt-check":      "true",
		"experimental-corrupt-check-time":         "240m",
	}

	etcdExtraArgsWithCiphers := maps.Clone(etcdExtraArgs)
	etcdExtraArgsWithCiphers["cipher-suites"] = "cipher1,cipher2"

	tests := []struct {
		name                 string
		kubeVersion          *semver.Version
		etcdImageTag         string
		expectedEtcdImageTag string
		cipherSuites         []string
		expectedEtcdArgs     map[string]string
	}{
		{
			name:                 "any 1.24",
			kubeVersion:          semver.MustParse("1.24"),
			expectedEtcdImageTag: fixedEtcdVersion,
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "any 1.24 high",
			kubeVersion:          semver.MustParse("1.24.999"),
			expectedEtcdImageTag: fixedEtcdVersion,
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "any 1.25",
			kubeVersion:          semver.MustParse("1.25"),
			expectedEtcdImageTag: fixedEtcdVersion,
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "any 1.25 high",
			kubeVersion:          semver.MustParse("1.25.999"),
			expectedEtcdImageTag: fixedEtcdVersion,
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "fixed 1.26",
			kubeVersion:          semver.MustParse("1.26.13"),
			expectedEtcdImageTag: "",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "unfixed 1.26",
			kubeVersion:          semver.MustParse("1.26.12"),
			expectedEtcdImageTag: fixedEtcdVersion,
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "fixed 1.27",
			kubeVersion:          semver.MustParse("1.27.9"),
			expectedEtcdImageTag: "",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "unfixed 1.27",
			kubeVersion:          semver.MustParse("1.27.8"),
			expectedEtcdImageTag: fixedEtcdVersion,
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "fixed 1.28",
			kubeVersion:          semver.MustParse("1.28.6"),
			expectedEtcdImageTag: "",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "unfixed 1.28",
			kubeVersion:          semver.MustParse("1.28.5"),
			expectedEtcdImageTag: fixedEtcdVersion,
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "any 1.29",
			kubeVersion:          semver.MustParse("1.29"),
			expectedEtcdImageTag: "",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "any 1.29 high",
			kubeVersion:          semver.MustParse("1.29.999"),
			expectedEtcdImageTag: "",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "unfixed, but tag is overwritten",
			kubeVersion:          semver.MustParse("1.26.12"),
			etcdImageTag:         "9.9.9-0",
			expectedEtcdImageTag: "9.9.9-0",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "fixed, but tag is overwritten",
			kubeVersion:          semver.MustParse("1.26.13"),
			etcdImageTag:         "9.9.9-0",
			expectedEtcdImageTag: "9.9.9-0",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "tls cipher suites",
			kubeVersion:          semver.MustParse("1.26.13"),
			etcdImageTag:         "9.9.9-0",
			expectedEtcdImageTag: "9.9.9-0",
			cipherSuites:         []string{"cipher1", "cipher2"},
			expectedEtcdArgs:     etcdExtraArgsWithCiphers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, args := etcdVersionCorruptCheckExtraArgs(tt.kubeVersion, tt.etcdImageTag, tt.cipherSuites)
			if ver != tt.expectedEtcdImageTag {
				t.Errorf("got etcd image tag %q, but expected %q", ver, tt.expectedEtcdImageTag)
			}
			if !reflect.DeepEqual(args, tt.expectedEtcdArgs) {
				t.Errorf("got etcd tags %q, but expected %q", args, tt.expectedEtcdArgs)
			}
		})
	}
}
