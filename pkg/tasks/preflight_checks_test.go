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

package tasks

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestParseContainerImageVersionValid(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name            string
		image           string
		expectedVersion *semver.Version
	}{
		{
			name:            "docker image",
			image:           "test/test-image:v1.13.3",
			expectedVersion: semver.MustParse("v1.13.3"),
		},
		{
			name:            "docker.io image",
			image:           "docker.io/test/test-image:v1.13.2",
			expectedVersion: semver.MustParse("v1.13.2"),
		},
		{
			name:            "gcr.io image",
			image:           "gcr.io/kubernetes/kubernetes:v1.14.0",
			expectedVersion: semver.MustParse("v1.14.0"),
		},
		{
			name:            "gcr.io image without v prefix",
			image:           "gcr.io/kubernetes:1.14.0",
			expectedVersion: semver.MustParse("1.14.0"),
		},
		{
			name:            "gcr.io image without v prefix and patch version",
			image:           "gcr.io/kubernetes:1.14",
			expectedVersion: semver.MustParse("v1.14.0"),
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ver, err := parseContainerImageVersion(tc.image)
			if err != nil {
				t.Fatal(err)
			}
			if !ver.Equal(tc.expectedVersion) {
				t.Fatalf("expected version %s, but got version %s", tc.expectedVersion.String(), ver.String())
			}
		})
	}
}

func TestParseContainerImageVersionInvalid(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name          string
		image         string
		expectedError error
	}{
		{
			name:          "docker image without version",
			image:         "test/test-image",
			expectedError: fmt.Errorf("invalid container image format: test/test-image"),
		},
		{
			name:          "docker image with the latest tag",
			image:         "test/test-image:latest",
			expectedError: fmt.Errorf("Invalid Semantic Version"),
		},
		{
			name:          "gcr.io image without version",
			image:         "gcr.io/kubernetes/kube-apiserver",
			expectedError: fmt.Errorf("invalid container image format: gcr.io/kubernetes/kube-apiserver"),
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseContainerImageVersion(tc.image)
			if err.Error() != tc.expectedError.Error() {
				t.Fatal(err)
			}
		})
	}
}

func TestCheckVersionSkewValid(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name           string
		currentVersion *semver.Version
		desiredVersion *semver.Version
		diff           uint64
	}{
		{
			name:           "upgrade 1.13.3 to 1.14.0 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.14.0"),
			diff:           1,
		},
		{
			name:           "upgrade 1.13.3 to 1.14.1 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.14.1"),
			diff:           1,
		},
		{
			name:           "upgrade 1.13.3 to 1.14.3 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.14.3"),
			diff:           1,
		},
		{
			name:           "upgrade 1.13.3 to 1.14.5 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.14.5"),
			diff:           1,
		},
		{
			name:           "upgrade 1.13.0 to 1.14.0 with diff of 1",
			currentVersion: semver.MustParse("1.13.0"),
			desiredVersion: semver.MustParse("1.14.0"),
			diff:           1,
		},
		{
			name:           "upgrade 1.13.0 to 1.14.1 with diff of 1",
			currentVersion: semver.MustParse("1.13.0"),
			desiredVersion: semver.MustParse("1.14.1"),
			diff:           1,
		},
		{
			name:           "upgrade 1.13.0 to 1.15.1 with diff of 2",
			currentVersion: semver.MustParse("1.13.0"),
			desiredVersion: semver.MustParse("1.15.1"),
			diff:           2,
		},
		{
			name:           "upgrade 1.13.0 to 1.13.3 with diff of 1",
			currentVersion: semver.MustParse("1.13.0"),
			desiredVersion: semver.MustParse("1.13.3"),
			diff:           1,
		},
		{
			name:           "upgrade 1.13.3 to 1.13.4 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.13.4"),
			diff:           1,
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := checkVersionSkew(tc.desiredVersion, tc.currentVersion, tc.diff)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCheckVersionSkewInvalid(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name           string
		currentVersion *semver.Version
		desiredVersion *semver.Version
		diff           uint64
	}{
		{
			name:           "upgrade 1.13.3 to 1.15.0 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.15.0"),
			diff:           1,
		},
		{
			name:           "upgrade 1.13.3 to 1.15.3 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.15.3"),
			diff:           1,
		},
		{
			name:           "upgrade 1.13.3 to 1.15.5 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.15.5"),
			diff:           1,
		},
		{
			name:           "upgrade 1.13.3 to 1.16.2 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.16.2"),
			diff:           1,
		},
		{
			name:           "upgrade 1.13.3 to 1.16.2 with diff of 2",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.16.2"),
			diff:           2,
		},
		{
			name:           "downgrade 1.13.0 to 1.12.1 with diff of 1",
			currentVersion: semver.MustParse("1.13.0"),
			desiredVersion: semver.MustParse("1.12.1"),
			diff:           1,
		},
		{
			name:           "downgrade 1.13.3 to 1.12.1 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.12.1"),
			diff:           1,
		},
		{
			name:           "downgrade 1.13.3 to 1.12.3 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.12.3"),
			diff:           1,
		},
		{
			name:           "downgrade 1.13.3 to 1.12.4 with diff of 1",
			currentVersion: semver.MustParse("1.13.3"),
			desiredVersion: semver.MustParse("1.12.4"),
			diff:           1,
		},
		{
			name:           "downgrade 1.13.3 to 1.13.2 with diff of 1",
			currentVersion: semver.MustParse("v1.13.3"),
			desiredVersion: semver.MustParse("v1.13.2"),
			diff:           1,
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := checkVersionSkew(tc.desiredVersion, tc.currentVersion, tc.diff)
			if err == nil {
				t.Fatalf("expected error but test succeed instead")
			}
		})
	}
}
