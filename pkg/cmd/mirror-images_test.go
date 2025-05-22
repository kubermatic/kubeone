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

package cmd

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestRetagImage(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		registry string
		want     string
		wantErr  bool
	}{
		{
			name:     "coredns special case",
			source:   "registry.k8s.io/coredns/coredns:v1.8.6",
			registry: "myregistry",
			want:     "myregistry/coredns:v1.8.6",
		},
		{
			name:     "regular image",
			source:   "nginx:latest",
			registry: "myregistry",
			want:     "myregistry/library/nginx:latest",
		},
		{
			name:     "Default kube-api-server image",
			source:   "registry.k8s.io/api-server:tag",
			registry: "myregistry",
			want:     "myregistry/api-server:tag",
		},
		{
			name:     "invalid image",
			source:   "invalid_image%%%_ref",
			registry: "myregistry",
			wantErr:  true,
		},
	}

	log := logrus.New()
	log.Out = io.Discard

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := retagImage(log, tt.source, tt.registry)

			if (err != nil) != tt.wantErr {
				t.Errorf("retagImage() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("retagImage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
