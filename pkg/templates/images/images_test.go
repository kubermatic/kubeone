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
	"testing"
)

func TestValidateProviders(t *testing.T) {
	tests := []struct {
		name      string
		providers []string
		wantErr   bool
	}{
		{name: "empty", providers: nil},
		{name: "none", providers: []string{ProviderNone}},
		{name: "single", providers: []string{"aws"}},
		{name: "multiple", providers: []string{"aws", "vmwareCloudDirector"}},
		{name: "unknown", providers: []string{"aws", "foo"}, wantErr: true},
		{name: "none combined", providers: []string{ProviderNone, "aws"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProviders(tt.providers)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProviders(%v) error = %v, wantErr %v", tt.providers, err, tt.wantErr)
			}
		})
	}
}

func TestResolverListWithProviders(t *testing.T) {
	r := NewResolver(WithKubernetesVersionGetter(func() string { return "1.36.0" }))

	img := func(res Resource) string { return r.Get(res) }

	tests := []struct {
		name      string
		filter    ListFilter
		providers []string
		want      []string
		notWant   []string
	}{
		{
			name:    "no providers lists all",
			filter:  ListFilterNone,
			want:    []string{img(AwsCCM), img(AzureCCM), img(HetznerCSI), img(CalicoNode)},
			notWant: nil,
		},
		{
			name:      "none lists all",
			filter:    ListFilterNone,
			providers: []string{ProviderNone},
			want:      []string{img(AwsCCM), img(AzureCCM), img(HetznerCSI), img(CalicoNode)},
		},
		{
			name:      "aws only",
			filter:    ListFilterNone,
			providers: []string{"aws"},
			want:      []string{img(AwsCCM), img(AwsEbsCSI), img(CalicoNode), img(CSISnapshotController), img(MachineController)},
			notWant:   []string{img(AzureCCM), img(HetznerCSI), img(VsphereCSIDriver)},
		},
		{
			name:      "multiple providers",
			filter:    ListFilterOptional,
			providers: []string{"hetzner", "openstack"},
			want:      []string{img(HetznerCCM), img(HetznerCSI), img(OpenstackCCM), img(OpenstackCSI), img(Cilium)},
			notWant:   []string{img(AwsCCM), img(GCPCCM), img(MachineController)},
		},
		{
			// DigitalOcean and Hetzner share identical sidecar images, which must
			// not be dropped when only one of them is selected.
			name:      "shared images kept",
			filter:    ListFilterNone,
			providers: []string{"hetzner"},
			want:      []string{img(HetznerCSIAttacher), img(DigitalOceanCSIAttacher)},
			notWant:   []string{img(DigitaloceanCCM)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.List(tt.filter, WithProviders(tt.providers...))

			for _, w := range tt.want {
				if !slices.Contains(got, w) {
					t.Errorf("expected image %q in the list", w)
				}
			}

			for _, nw := range tt.notWant {
				if slices.Contains(got, nw) {
					t.Errorf("unexpected image %q in the list", nw)
				}
			}
		})
	}

	t.Run("base unaffected by providers", func(t *testing.T) {
		if !slices.Equal(r.List(ListFilterBase), r.List(ListFilterBase, WithProviders("aws"))) {
			t.Error("base images list must not depend on providers")
		}
	})
}

func TestResolverListAllWithProviders(t *testing.T) {
	r := NewResolver()

	all := r.ListAll()
	aws := r.ListAll(WithProviders("aws"))

	if len(aws) >= len(all) {
		t.Fatalf("expected filtered list (%d) to be shorter than full list (%d)", len(aws), len(all))
	}

	for _, img := range allResources()[AzureCCM] {
		if slices.Contains(aws, img) {
			t.Errorf("unexpected azure image %q in aws list", img)
		}
	}

	for _, img := range allResources()[AwsCCM] {
		if !slices.Contains(aws, img) {
			t.Errorf("expected aws image %q in aws list", img)
		}
	}
}
