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

package images

import (
	"io"
	"reflect"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sirupsen/logrus"
)

func TestRetagImage(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		registry string
		want     []string
		wantErr  bool
	}{
		{
			name:     "coredns special case",
			source:   "registry.k8s.io/coredns/coredns:v1.8.6",
			registry: "myregistry",
			want:     []string{"myregistry/coredns/coredns:v1.8.6", "myregistry/coredns:v1.8.6"},
		},
		{
			name:     "regular image",
			source:   "nginx:latest",
			registry: "myregistry",
			want:     []string{"myregistry/library/nginx:latest"},
		},
		{
			name:     "Default kube-api-server image",
			source:   "registry.k8s.io/api-server:tag",
			registry: "myregistry",
			want:     []string{"myregistry/api-server:tag"},
		},
		{
			// Identifier() returns the digest for a digest-pinned reference. Joined with
			// ":" it produces "<registry>/<repo>:sha256:..." which is not a parseable
			// reference, so the copy fails before any bytes move.
			name:     "digest-pinned image keeps the @ separator",
			source:   "quay.io/operator-framework/olm@sha256:e74b2ac57963c7f3ba19122a8c31c9f2a0deb3c0c5cac9e5323ccffd0ca198ed",
			registry: "myregistry",
			want:     []string{"myregistry/operator-framework/olm@sha256:e74b2ac57963c7f3ba19122a8c31c9f2a0deb3c0c5cac9e5323ccffd0ca198ed"},
		},
		{
			name:     "coredns pinned by digest",
			source:   "registry.k8s.io/coredns/coredns@sha256:e74b2ac57963c7f3ba19122a8c31c9f2a0deb3c0c5cac9e5323ccffd0ca198ed",
			registry: "myregistry",
			want: []string{
				"myregistry/coredns/coredns@sha256:e74b2ac57963c7f3ba19122a8c31c9f2a0deb3c0c5cac9e5323ccffd0ca198ed",
				"myregistry/coredns@sha256:e74b2ac57963c7f3ba19122a8c31c9f2a0deb3c0c5cac9e5323ccffd0ca198ed",
			},
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

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("retagImage() got = %v, want %v", got, tt.want)
			}

			// A destination that does not parse fails the copy before it starts, which
			// is how digest-pinned sources used to break.
			for _, dest := range got {
				if _, err := name.ParseReference(dest); err != nil {
					t.Errorf("retagImage() produced unparseable reference %q: %v", dest, err)
				}
			}
		})
	}
}
