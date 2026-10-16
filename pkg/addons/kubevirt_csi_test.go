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
	"slices"
	"testing"

	embeddedaddons "k8c.io/kubeone/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubevirtcsi "k8c.io/kubeone/pkg/cloudprovider/kubevirt/csi"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/images"
	"k8c.io/kubeone/pkg/templates/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

func TestKubeVirtCSIDriverOperatorAddon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		overwriteRegistry string
		wantArgs          []string
	}{
		{
			name:     "default registry",
			wantArgs: []string{"--leader-elect"},
		},
		{
			name:              "overwrite registry",
			overwriteRegistry: "registry.example.com",
			wantArgs:          []string{"--leader-elect", "--overwrite-registry=registry.example.com"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cluster := &kubeoneapi.KubeOneCluster{
				Name: "kubeone-test",
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Kubevirt: &kubeoneapi.KubevirtSpec{InfraNamespace: "infra-ns"},
				},
				RegistryConfiguration: &kubeoneapi.RegistryConfiguration{OverwriteRegistry: tc.overwriteRegistry},
			}

			applier := &applier{
				TemplateData: templateData{
					Config: cluster,
					InternalImages: &internalImages{
						resolver: func(images.Resource, ...images.GetOpt) string {
							return "quay.io/kubermatic/kubevirt-csi-driver-operator:test"
						},
					},
				},
			}

			manifests, err := applier.loadAddonsManifests(embeddedaddons.FS, resources.AddonKubeVirtCSIDriverOperator, nil, nil, false, cluster, false)
			if err != nil {
				t.Fatalf("unable to render manifests: %v", err)
			}

			kinds := map[string]int{}
			var (
				deployment *appsv1.Deployment
				token      *corev1.Secret
			)

			for _, m := range manifests {
				var obj metav1.PartialObjectMetadata
				if err := yaml.Unmarshal(m.Raw, &obj); err != nil {
					t.Fatalf("unable to unmarshal manifest: %v", err)
				}
				kinds[obj.Kind]++

				switch {
				case obj.Kind == "Deployment":
					deployment = &appsv1.Deployment{}
					if err := yaml.Unmarshal(m.Raw, deployment); err != nil {
						t.Fatal(err)
					}
				case obj.Kind == "Secret" && obj.Name == kubevirtcsi.TenantTokenSecretName:
					token = &corev1.Secret{}
					if err := yaml.Unmarshal(m.Raw, token); err != nil {
						t.Fatal(err)
					}
				}
			}

			wantKinds := map[string]int{
				"CustomResourceDefinition": 1,
				"Namespace":                1,
				"ServiceAccount":           2,
				"Secret":                   1,
				"ClusterRole":              2,
				"ClusterRoleBinding":       2,
				"Deployment":               1,
			}
			for kind, n := range wantKinds {
				if kinds[kind] != n {
					t.Errorf("got %d %s objects, want %d", kinds[kind], kind, n)
				}
			}

			if deployment == nil {
				t.Fatal("operator Deployment not found")
			}

			if got := deployment.Spec.Template.Spec.Containers[0].Args; !slices.Equal(got, tc.wantArgs) {
				t.Errorf("operator args = %v, want %v", got, tc.wantArgs)
			}

			if token == nil {
				t.Fatal("controller token Secret not found")
			}

			if token.Namespace != kubevirtcsi.TenantNamespace ||
				token.Type != corev1.SecretTypeServiceAccountToken ||
				token.Annotations[corev1.ServiceAccountNameKey] != kubevirtcsi.TenantServiceAccountName {
				t.Errorf("token Secret is not bound to %s/%s: %+v", kubevirtcsi.TenantNamespace, kubevirtcsi.TenantServiceAccountName, token)
			}
		})
	}
}

func TestMigrateKubeVirtCSI(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	ns := metav1.NamespaceSystem
	legacy := []client.Object{
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "kubevirt-csi-controller", Namespace: ns}},
		&appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: "kubevirt-csi-node", Namespace: ns}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "driver-config", Namespace: ns}},
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "kubevirt-csi-controller-sa", Namespace: ns}},
		&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "kubevirt-csi-node-cr"}},
	}
	// The CSIDriver is managed by the operator and must not be deleted
	csiDriver := &storagev1.CSIDriver{ObjectMeta: metav1.ObjectMeta{Name: "csi.kubevirt.io"}}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(append(legacy, csiDriver)...).Build()
	s := &state.State{Context: context.Background(), DynamicClient: c}

	if err := migrateKubeVirtCSI(s); err != nil {
		t.Fatal(err)
	}

	// idempotent
	if err := migrateKubeVirtCSI(s); err != nil {
		t.Fatal(err)
	}

	for _, obj := range legacy {
		if err := c.Get(s.Context, client.ObjectKeyFromObject(obj), obj); !apierrors.IsNotFound(err) {
			t.Errorf("%T %s: expected NotFound, got %v", obj, obj.GetName(), err)
		}
	}

	if err := c.Get(s.Context, client.ObjectKeyFromObject(csiDriver), csiDriver); err != nil {
		t.Errorf("CSIDriver should be kept: %v", err)
	}
}

func TestKubeVirtCSIAddonSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		disableCSI  bool
		wantPresent bool
	}{
		{name: "bundled CSI drivers enabled", wantPresent: true},
		{name: "bundled CSI drivers disabled", disableCSI: true, wantPresent: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &state.State{
				Cluster: &kubeoneapi.KubeOneCluster{
					CloudProvider: kubeoneapi.CloudProviderSpec{
						Kubevirt:                 &kubeoneapi.KubevirtSpec{InfraNamespace: "infra-ns"},
						DisableBundledCSIDrivers: tc.disableCSI,
					},
					ClusterNetwork: kubeoneapi.ClusterNetworkConfig{CNI: &kubeoneapi.CNI{}},
					Features: kubeoneapi.Features{
						CoreDNS:       &kubeoneapi.CoreDNS{DeployPodDisruptionBudget: new(false)},
						MetricsServer: &kubeoneapi.MetricsServer{},
						NodeLocalDNS:  &kubeoneapi.NodeLocalDNS{},
					},
					MachineController:      &kubeoneapi.MachineControllerConfig{},
					OperatingSystemManager: &kubeoneapi.OperatingSystemManagerConfig{},
				},
			}

			var found *addonAction
			for _, a := range collectAddons(s) {
				if a.name == resources.AddonKubeVirtCSIDriverOperator {
					found = &a
				}
				if a.name == "csi-kubevirt" {
					t.Error("legacy csi-kubevirt addon must not be deployed")
				}
			}

			if (found != nil) != tc.wantPresent {
				t.Fatalf("addon present = %v, want %v", found != nil, tc.wantPresent)
			}

			if found != nil && (found.supportFn == nil || found.postFn == nil) {
				t.Error("addon must have both a migration supportFn and a postFn")
			}
		})
	}
}
