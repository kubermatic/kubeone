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
	"slices"
	"strings"
	"testing"
)

func TestSupportedProviders(t *testing.T) {
	providers := SupportedProviders()
	if len(providers) == 0 {
		t.Fatal("SupportedProviders returned empty list")
	}

	// must be sorted
	for i := 1; i < len(providers); i++ {
		if providers[i] < providers[i-1] {
			t.Errorf("SupportedProviders not sorted: %q before %q", providers[i-1], providers[i])
		}
	}

	// spot-check well-known providers are present
	for _, want := range []string{"aws", "azure", "gce", "hetzner", "openstack", "vsphere"} {
		if !slices.Contains(providers, want) {
			t.Errorf("SupportedProviders missing expected provider %q", want)
		}
	}
}

func TestListForProvider_UnknownProvider(t *testing.T) {
	r := NewResolver()
	_, err := r.ListForProvider("totally-unknown")
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
	if !strings.Contains(err.Error(), "totally-unknown") {
		t.Errorf("error message should mention the unknown provider; got: %v", err)
	}
}

func TestListForProvider_NoneProvider(t *testing.T) {
	r := NewResolver()
	imgs, err := r.ListForProvider("none")
	if err != nil {
		t.Fatalf("unexpected error for 'none' provider: %v", err)
	}
	if len(imgs) != 0 {
		t.Errorf("expected empty list for 'none' provider, got %d images", len(imgs))
	}
}

func TestListForProvider_AWS(t *testing.T) {
	r := NewResolver(WithKubernetesVersionGetter(func() string { return "1.34.0" }))
	imgs, err := r.ListForProvider("aws")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(imgs) == 0 {
		t.Fatal("expected images for 'aws' provider, got none")
	}

	// All returned images should be non-empty strings.
	for _, img := range imgs {
		if img == "" {
			t.Error("ListForProvider returned an empty image string")
		}
	}

	// Result must be sorted.
	for i := 1; i < len(imgs); i++ {
		if imgs[i] < imgs[i-1] {
			t.Errorf("result not sorted: %q before %q", imgs[i-1], imgs[i])
		}
	}

	// AWS CCM image must be present.
	found := false
	for _, img := range imgs {
		if strings.Contains(img, "provider-aws") {
			found = true

			break
		}
	}
	if !found {
		t.Errorf("AWS CCM image (provider-aws) not found in result: %v", imgs)
	}
}

func TestListForProvider_SharedImagesIncluded(t *testing.T) {
	r := NewResolver(WithKubernetesVersionGetter(func() string { return "1.34.0" }))

	// CSISnapshotController should appear in both AWS and vSphere results.
	for _, provider := range []string{"aws", "vsphere"} {
		imgs, err := r.ListForProvider(provider)
		if err != nil {
			t.Fatalf("provider %q: unexpected error: %v", provider, err)
		}

		found := false
		for _, img := range imgs {
			if strings.Contains(img, "snapshot-controller") {
				found = true

				break
			}
		}
		if !found {
			t.Errorf("provider %q: CSISnapshotController not found in result: %v", provider, imgs)
		}
	}
}

func TestListForProvider_AllProvidersReturnNonNil(t *testing.T) {
	r := NewResolver(WithKubernetesVersionGetter(func() string { return "1.34.0" }))
	for _, provider := range SupportedProviders() {
		imgs, err := r.ListForProvider(provider)
		if err != nil {
			t.Errorf("provider %q: unexpected error: %v", provider, err)
		}
		if imgs == nil {
			t.Errorf("provider %q: expected non-nil slice, got nil", provider)
		}
	}
}
