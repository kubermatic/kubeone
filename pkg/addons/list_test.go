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

package addons

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"
)

func TestResolveAddons(t *testing.T) {
	t.Parallel()

	addonsDir := t.TempDir()
	for _, dir := range []string{resources.AddonMetricsServer, "my-custom-ingress"} {
		if err := os.Mkdir(filepath.Join(addonsDir, dir), 0o750); err != nil {
			t.Fatalf("unable to create addon directory: %v", err)
		}
	}

	s := &state.State{
		ManifestFilePath: filepath.Join(addonsDir, "kubeone.yaml"),
		Cluster: &kubeoneapi.KubeOneCluster{
			Features: kubeoneapi.Features{
				CoreDNS:       &kubeoneapi.CoreDNS{DeployPodDisruptionBudget: new(bool)},
				MetricsServer: &kubeoneapi.MetricsServer{Enable: true},
				NodeLocalDNS:  &kubeoneapi.NodeLocalDNS{},
			},
			CloudProvider:          kubeoneapi.CloudProviderSpec{DisableBundledCSIDrivers: true},
			ClusterNetwork:         kubeoneapi.ClusterNetworkConfig{CNI: &kubeoneapi.CNI{}},
			MachineController:      &kubeoneapi.MachineControllerConfig{},
			OperatingSystemManager: &kubeoneapi.OperatingSystemManagerConfig{},
			Addons: &kubeoneapi.Addons{
				Path: addonsDir,
				Addons: []kubeoneapi.AddonRef{
					{Addon: &kubeoneapi.Addon{Name: "unattended-upgrades"}},
					{Addon: &kubeoneapi.Addon{Name: resources.AddonCCMAws, Delete: true}},
					{Addon: &kubeoneapi.Addon{Name: "not-existing"}},
				},
			},
		},
	}

	items, err := ResolveAddons(s)
	if err != nil {
		t.Fatalf("ResolveAddons() returned an error: %v", err)
	}

	if !slices.IsSortedFunc(items, func(a, b AddonItem) int {
		return strings.Compare(a.Name, b.Name)
	}) {
		t.Errorf("ResolveAddons() result is not sorted by name: %v", items)
	}

	expected := map[string]AddonItem{
		// embedded addon, enabled by the cluster configuration, overridden by a directory
		resources.AddonMetricsServer: {Name: resources.AddonMetricsServer, Status: AddonStatusInstall, Source: AddonSourceOverridden},
		// only in the addons directory
		"my-custom-ingress": {Name: "my-custom-ingress", Status: AddonStatusInstall, Source: AddonSourceCustom},
		// embedded addon, declared in the manifest
		"unattended-upgrades": {Name: "unattended-upgrades", Status: AddonStatusInstall, Source: AddonSourceEmbedded},
		// embedded addon, declared in the manifest with delete: true
		resources.AddonCCMAws: {Name: resources.AddonCCMAws, Status: AddonStatusDelete, Source: AddonSourceEmbedded},
		// embedded addon, not used
		resources.AddonCCMAzure: {Name: resources.AddonCCMAzure, Status: AddonStatusInactive, Source: AddonSourceEmbedded},
		// declared in the manifest, but neither embedded nor in the addons directory
		"not-existing": {Name: "not-existing", Status: AddonStatusInstall},
	}

	got := map[string]AddonItem{}
	for _, item := range items {
		got[item.Name] = item
	}

	for name, want := range expected {
		if got[name] != want {
			t.Errorf("addon %q: got %+v, want %+v", name, got[name], want)
		}
	}
}
