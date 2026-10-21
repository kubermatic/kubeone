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
	"testing"

	embeddedaddons "k8c.io/kubeone/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/images"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

func Test_flatcarNodesPresent(t *testing.T) {
	t.Parallel()

	host := func(os kubeoneapi.OperatingSystemName) kubeoneapi.HostConfig {
		return kubeoneapi.HostConfig{OperatingSystem: os}
	}

	tests := []struct {
		name     string
		cluster  *kubeoneapi.KubeOneCluster
		live     *state.Cluster
		expected bool
	}{
		{
			name:     "nil cluster and nil live cluster",
			expected: false,
		},
		{
			name: "no flatcar nodes",
			cluster: &kubeoneapi.KubeOneCluster{
				ControlPlane:  kubeoneapi.ControlPlaneConfig{Hosts: []kubeoneapi.HostConfig{host(kubeoneapi.OperatingSystemNameUbuntu)}},
				StaticWorkers: kubeoneapi.StaticWorkersConfig{Hosts: []kubeoneapi.HostConfig{host(kubeoneapi.OperatingSystemNameRockyLinux)}},
				DynamicWorkers: []kubeoneapi.DynamicWorkerConfig{
					{Config: kubeoneapi.ProviderSpec{OperatingSystem: "ubuntu"}},
				},
			},
			expected: false,
		},
		{
			name: "flatcar control plane",
			cluster: &kubeoneapi.KubeOneCluster{
				ControlPlane: kubeoneapi.ControlPlaneConfig{Hosts: []kubeoneapi.HostConfig{host(kubeoneapi.OperatingSystemNameFlatcar)}},
			},
			expected: true,
		},
		{
			name: "flatcar static worker",
			cluster: &kubeoneapi.KubeOneCluster{
				ControlPlane:  kubeoneapi.ControlPlaneConfig{Hosts: []kubeoneapi.HostConfig{host(kubeoneapi.OperatingSystemNameUbuntu)}},
				StaticWorkers: kubeoneapi.StaticWorkersConfig{Hosts: []kubeoneapi.HostConfig{host(kubeoneapi.OperatingSystemNameFlatcar)}},
			},
			expected: true,
		},
		{
			name: "flatcar dynamic worker",
			cluster: &kubeoneapi.KubeOneCluster{
				ControlPlane: kubeoneapi.ControlPlaneConfig{Hosts: []kubeoneapi.HostConfig{host(kubeoneapi.OperatingSystemNameUbuntu)}},
				DynamicWorkers: []kubeoneapi.DynamicWorkerConfig{
					{Config: kubeoneapi.ProviderSpec{OperatingSystem: "ubuntu"}},
					{Config: kubeoneapi.ProviderSpec{OperatingSystem: "flatcar"}},
				},
			},
			expected: true,
		},
		{
			name: "flatcar only known from the live cluster",
			cluster: &kubeoneapi.KubeOneCluster{
				ControlPlane: kubeoneapi.ControlPlaneConfig{Hosts: []kubeoneapi.HostConfig{host(kubeoneapi.OperatingSystemNameUbuntu)}},
			},
			live: &state.Cluster{
				ControlPlane:  []state.Host{{Config: &kubeoneapi.HostConfig{OperatingSystem: kubeoneapi.OperatingSystemNameUbuntu}}},
				StaticWorkers: []state.Host{{Config: &kubeoneapi.HostConfig{OperatingSystem: kubeoneapi.OperatingSystemNameFlatcar}}},
			},
			expected: true,
		},
		{
			name:     "live cluster hosts without config",
			cluster:  &kubeoneapi.KubeOneCluster{},
			live:     &state.Cluster{ControlPlane: []state.Host{{}}},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := flatcarNodesPresent(tc.cluster, tc.live); got != tc.expected {
				t.Errorf("flatcarNodesPresent() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestUnattendedUpgradesFluoOnlyWithFlatcarNodes(t *testing.T) {
	t.Parallel()

	const fluoNamespace = "reboot-coordinator"

	tests := []struct {
		name         string
		flatcarNodes bool
		expectFluo   bool
	}{
		{
			name:         "without flatcar nodes fluo is not rendered",
			flatcarNodes: false,
			expectFluo:   false,
		},
		{
			name:         "with flatcar nodes fluo is rendered",
			flatcarNodes: true,
			expectFluo:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := &applier{
				TemplateData: templateData{
					Config:              &kubeoneapi.KubeOneCluster{Name: "kubeone-test"},
					FlatcarNodesPresent: tc.flatcarNodes,
					InternalImages: &internalImages{
						resolver: func(images.Resource, ...images.GetOpt) string { return "registry.example.com/image:tag" },
					},
					Params: map[string]string{},
				},
				EmbeddedFS: embeddedaddons.FS,
			}

			manifests, err := a.loadAddonsManifests(a.EmbeddedFS, "unattended-upgrades", nil, logrus.New(), false, &kubeoneapi.KubeOneCluster{}, false)
			if err != nil {
				t.Fatalf("unable to load manifests: %v", err)
			}

			if len(manifests) == 0 {
				t.Fatal("expected the addon to render at least the distro specific manifests")
			}

			var hasFluo bool
			for _, raw := range manifests {
				hasFluo = hasFluo || referencesFluo(t, raw, fluoNamespace)
			}

			if hasFluo != tc.expectFluo {
				t.Errorf("fluo rendered = %v, want %v", hasFluo, tc.expectFluo)
			}
		})
	}
}

// referencesFluo returns true if the object is part of the Flatcar Linux Update
// Operator: it lives in the reboot-coordinator namespace or is a cluster-scoped
// RBAC object named after the operator or its agent.
func referencesFluo(t *testing.T, raw runtime.RawExtension, namespace string) bool {
	t.Helper()

	var obj struct {
		Kind     string `json:"kind"`
		Metadata struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
	}

	if err := yaml.Unmarshal(raw.Raw, &obj); err != nil {
		t.Fatalf("unable to unmarshal rendered manifest: %v", err)
	}

	switch obj.Kind {
	case "Namespace":
		return obj.Metadata.Name == namespace
	case "ClusterRole", "ClusterRoleBinding":
		return obj.Metadata.Name == "flatcar-linux-update-operator" || obj.Metadata.Name == "flatcar-linux-update-agent"
	default:
		return obj.Metadata.Namespace == namespace
	}
}
