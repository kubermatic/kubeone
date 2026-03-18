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

package tasks

import "testing"

func Test_kernelSemver(t *testing.T) {
	tests := []struct {
		name     string
		krelease string
		want     string
		wantErr  bool
	}{
		{
			name:     "simple kernel version",
			krelease: "6.12.74",
			want:     "6.12.74",
		},
		{
			name:     "kernel version with revision and arch",
			krelease: "6.12.74-13-amd64",
			want:     "6.12.74",
		},
		{
			name:     "kernel version with deb plus suffix and arch",
			krelease: "6.12.74+deb13+1-amd64",
			want:     "6.12.74",
		},
		{
			name:     "kernel version with deb plus suffix no arch",
			krelease: "6.12.74+deb13+1",
			want:     "6.12.74",
		},
		{
			name:     "kernel version with single plus suffix and arch",
			krelease: "6.12.74+deb13-amd64",
			want:     "6.12.74",
		},
		{
			name:     "empty string",
			krelease: "",
			wantErr:  true,
		},
		{
			name:     "not a version",
			krelease: "bad-kernel",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kernelSemver(tt.krelease)
			if tt.wantErr {
				if err == nil {
					t.Errorf("kernelSemver(%q) expected error, got nil", tt.krelease)
				}

				return
			}
			if err != nil {
				t.Errorf("kernelSemver(%q) unexpected error: %v", tt.krelease, err)

				return
			}
			if got.String() != tt.want {
				t.Errorf("kernelSemver(%q) = %q, want %q", tt.krelease, got.String(), tt.want)
			}
		})
	}
}
