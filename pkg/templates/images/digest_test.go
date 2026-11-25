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

package images

import (
	"maps"
	"slices"
	"testing"

	"github.com/distribution/reference"
	"github.com/google/go-containerregistry/pkg/crane"
)

// TestImageDigests resolves the tag of every defined image against its real
// registry and checks that the pinned digest matches.
func TestImageDigests(t *testing.T) {
	t.Parallel()

	seen := map[string]struct{}{}
	for _, versions := range allResources() {
		for _, img := range versions {
			seen[img] = struct{}{}
		}
	}

	for _, img := range slices.Sorted(maps.Keys(seen)) {
		t.Run(img, func(t *testing.T) {
			t.Parallel()

			named, err := reference.ParseNormalizedNamed(img)
			if err != nil {
				t.Fatalf("parsing image reference: %v", err)
			}

			tagged, ok := named.(reference.Tagged)
			if !ok {
				t.Fatalf("image has no tag")
			}

			canonical, ok := named.(reference.Canonical)
			if !ok {
				t.Fatalf("image is not pinned by digest")
			}

			tagRef := named.Name() + ":" + tagged.Tag()
			want := canonical.Digest().String()

			opts := []crane.Option{crane.WithContext(t.Context())}

			got, err := crane.Digest(tagRef, opts...)
			if err != nil {
				t.Fatalf("resolving digest of %s: %v", tagRef, err)
			}

			if got == want {
				return
			}

			// Tell apart a tag that moved from a digest that never belonged to
			// this repository.
			digestRef := named.Name() + "@" + want
			if _, err := crane.Head(digestRef, opts...); err != nil {
				t.Errorf("digest mismatch for %s: pinned %s does not exist in the repository (%v), tag resolves to %s", tagRef, want, err, got)
			} else {
				t.Errorf("digest mismatch for %s: pinned %s, tag now resolves to %s", tagRef, want, got)
			}
		})
	}
}
