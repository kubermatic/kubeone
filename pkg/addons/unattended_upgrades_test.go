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
	"context"
	"testing"

	embeddedaddons "k8c.io/kubeone/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/images"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"

	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

func machineDeployment(name, operatingSystem string) *clusterv1alpha1.MachineDeployment {
	md := &clusterv1alpha1.MachineDeployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: metav1.NamespaceSystem},
	}
	md.Spec.Template.Spec.ProviderSpec.Value = &runtime.RawExtension{
		Raw: []byte(`{"operatingSystem":"` + operatingSystem + `"}`),
	}

	return md
}

func newFakeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("unable to build scheme: %v", err)
	}

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func Test_flatcarNodesPresent(t *testing.T) {
	t.Parallel()

	host := func(os kubeoneapi.OperatingSystemName) kubeoneapi.HostConfig {
		return kubeoneapi.HostConfig{OperatingSystem: os}
	}

	ubuntuCluster := func() *kubeoneapi.KubeOneCluster {
		return &kubeoneapi.KubeOneCluster{
			ControlPlane:      kubeoneapi.ControlPlaneConfig{Hosts: []kubeoneapi.HostConfig{host(kubeoneapi.OperatingSystemNameUbuntu)}},
			MachineController: &kubeoneapi.MachineControllerConfig{Deploy: true},
		}
	}

	tests := []struct {
		name     string
		cluster  func() *kubeoneapi.KubeOneCluster
		objs     []client.Object
		expected bool
	}{
		{
			name:    "no flatcar nodes",
			cluster: ubuntuCluster,
			objs:    []client.Object{machineDeployment("md-ubuntu", "ubuntu")},
		},
		{
			name: "flatcar control plane",
			cluster: func() *kubeoneapi.KubeOneCluster {
				c := ubuntuCluster()
				c.ControlPlane.Hosts = append(c.ControlPlane.Hosts, host(kubeoneapi.OperatingSystemNameFlatcar))

				return c
			},
			expected: true,
		},
		{
			name: "flatcar static worker",
			cluster: func() *kubeoneapi.KubeOneCluster {
				c := ubuntuCluster()
				c.StaticWorkers.Hosts = []kubeoneapi.HostConfig{host(kubeoneapi.OperatingSystemNameFlatcar)}

				return c
			},
			expected: true,
		},
		{
			name: "flatcar dynamic worker in the manifest, not created yet",
			cluster: func() *kubeoneapi.KubeOneCluster {
				c := ubuntuCluster()
				c.DynamicWorkers = []kubeoneapi.DynamicWorkerConfig{{Config: kubeoneapi.ProviderSpec{OperatingSystem: "flatcar"}}}

				return c
			},
			expected: true,
		},
		{
			name:     "flatcar MachineDeployment that only exists in the cluster",
			cluster:  ubuntuCluster,
			objs:     []client.Object{machineDeployment("md-ubuntu", "ubuntu"), machineDeployment("md-flatcar", "flatcar")},
			expected: true,
		},
		{
			name: "flatcar MachineDeployment is ignored if machine-controller is not deployed",
			cluster: func() *kubeoneapi.KubeOneCluster {
				c := ubuntuCluster()
				c.MachineController.Deploy = false

				return c
			},
			objs: []client.Object{machineDeployment("md-flatcar", "flatcar")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := flatcarNodesPresent(context.Background(), tc.cluster(), newFakeClient(t, tc.objs...))
			if err != nil {
				t.Fatalf("flatcarNodesPresent() returned an error: %v", err)
			}

			if got != tc.expected {
				t.Errorf("flatcarNodesPresent() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func Test_withFlatcarNodesPresentParam(t *testing.T) {
	t.Parallel()

	newState := func(objs ...client.Object) *state.State {
		return &state.State{
			Context: context.Background(),
			Cluster: &kubeoneapi.KubeOneCluster{
				ControlPlane:      kubeoneapi.ControlPlaneConfig{Hosts: []kubeoneapi.HostConfig{{OperatingSystem: kubeoneapi.OperatingSystemNameUbuntu}}},
				MachineController: &kubeoneapi.MachineControllerConfig{Deploy: true},
			},
			DynamicClient: newFakeClient(t, objs...),
		}
	}

	t.Run("param is derived and the passed params are not modified", func(t *testing.T) {
		t.Parallel()

		in := map[string]string{"other": "value"}

		got, err := withFlatcarNodesPresentParam(newState(machineDeployment("md-flatcar", "flatcar")), in)
		if err != nil {
			t.Fatal(err)
		}

		if got[paramFlatcarNodesPresent] != "true" || got["other"] != "value" {
			t.Errorf("unexpected params: %v", got)
		}

		if _, ok := in[paramFlatcarNodesPresent]; ok {
			t.Errorf("the passed params were modified: %v", in)
		}
	})

	t.Run("nil params without flatcar nodes", func(t *testing.T) {
		t.Parallel()

		got, err := withFlatcarNodesPresentParam(newState(), nil)
		if err != nil {
			t.Fatal(err)
		}

		if got[paramFlatcarNodesPresent] != "false" {
			t.Errorf("unexpected params: %v", got)
		}
	})

	t.Run("param set by the user wins", func(t *testing.T) {
		t.Parallel()

		got, err := withFlatcarNodesPresentParam(newState(), map[string]string{paramFlatcarNodesPresent: "true"})
		if err != nil {
			t.Fatal(err)
		}

		if got[paramFlatcarNodesPresent] != "true" {
			t.Errorf("unexpected params: %v", got)
		}
	})
}

func TestUnattendedUpgradesFluoOnlyWithFlatcarNodes(t *testing.T) {
	t.Parallel()

	const fluoNamespace = "reboot-coordinator"

	tests := []struct {
		name       string
		params     map[string]string
		expectFluo bool
	}{
		{
			name:       "fluo is not rendered without flatcar nodes",
			params:     map[string]string{paramFlatcarNodesPresent: "false"},
			expectFluo: false,
		},
		{
			name:       "fluo is not rendered if the param is missing",
			params:     nil,
			expectFluo: false,
		},
		{
			name:       "fluo is rendered with flatcar nodes",
			params:     map[string]string{paramFlatcarNodesPresent: "true"},
			expectFluo: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := &applier{
				TemplateData: templateData{
					Config: &kubeoneapi.KubeOneCluster{Name: "kubeone-test"},
					InternalImages: &internalImages{
						resolver: func(images.Resource, ...images.GetOpt) string { return "registry.example.com/image:tag" },
					},
					Params: map[string]string{},
				},
				EmbeddedFS: embeddedaddons.FS,
			}

			manifests, err := a.loadAddonsManifests(a.EmbeddedFS, "unattended-upgrades", tc.params, logrus.New(), false, &kubeoneapi.KubeOneCluster{}, false)
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
