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
	kubevirtccm "k8c.io/kubeone/pkg/cloudprovider/kubevirt/ccm"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/images"
	"k8c.io/kubeone/pkg/templates/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

func TestKubeVirtCCMAddon(t *testing.T) {
	t.Parallel()

	cluster := &kubeoneapi.KubeOneCluster{
		Name: "kubeone-test",
		CloudProvider: kubeoneapi.CloudProviderSpec{
			External: true,
			Kubevirt: &kubeoneapi.KubevirtSpec{InfraNamespace: "infra-ns"},
		},
		RegistryConfiguration: &kubeoneapi.RegistryConfiguration{},
	}

	applier := &applier{
		TemplateData: templateData{
			Config: cluster,
			InternalImages: &internalImages{
				resolver: func(images.Resource, ...images.GetOpt) string { return "img:test" },
			},
		},
	}

	manifests, err := applier.loadAddonsManifests(embeddedaddons.FS, resources.AddonCCMKubeVirt, nil, nil, false, cluster, false)
	if err != nil {
		t.Fatalf("unable to render manifests: %v", err)
	}

	var token *corev1.Secret
	for _, m := range manifests {
		var obj metav1.PartialObjectMetadata
		if err := yaml.Unmarshal(m.Raw, &obj); err != nil {
			t.Fatal(err)
		}

		switch obj.Kind {
		case "Deployment", "ConfigMap", "Role":
			// The CCM, its cloud-config and its infra permissions live in the infra cluster
			t.Errorf("unexpected %s %s/%s in the tenant addon", obj.Kind, obj.Namespace, obj.Name)
		case "Secret":
			token = &corev1.Secret{}
			if err := yaml.Unmarshal(m.Raw, token); err != nil {
				t.Fatal(err)
			}
		}
	}

	if token == nil {
		t.Fatal("token Secret not found")
	}

	if token.Name != kubevirtccm.TenantTokenSecretName ||
		token.Namespace != kubevirtccm.TenantNamespace ||
		token.Type != corev1.SecretTypeServiceAccountToken ||
		token.Annotations[corev1.ServiceAccountNameKey] != kubevirtccm.TenantServiceAccountName {
		t.Errorf("token Secret is not bound to %s/%s: %+v", kubevirtccm.TenantNamespace, kubevirtccm.TenantServiceAccountName, token)
	}
}

func TestMigrateKubeVirtCCM(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	ns := metav1.NamespaceSystem
	legacy := []client.Object{
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "kubevirt-cloud-controller-manager", Namespace: ns}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cloud-config", Namespace: ns}},
		&rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "kccm", Namespace: ns}},
		&rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "kccm-sa", Namespace: ns}},
	}
	kept := []client.Object{
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "cloud-controller-manager", Namespace: ns}},
		&rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "kccm", Namespace: ns}},
		&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "kccm"}},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(append(legacy, kept...)...).Build()
	s := &state.State{Context: context.Background(), DynamicClient: c}

	for range 2 { // idempotent
		if err := migrateKubeVirtCCM(s); err != nil {
			t.Fatal(err)
		}
	}

	for _, obj := range legacy {
		if err := c.Get(s.Context, client.ObjectKeyFromObject(obj), obj); !apierrors.IsNotFound(err) {
			t.Errorf("%T %s: expected NotFound, got %v", obj, obj.GetName(), err)
		}
	}

	for _, obj := range kept {
		if err := c.Get(s.Context, client.ObjectKeyFromObject(obj), obj); err != nil {
			t.Errorf("%T %s: expected to be kept, got %v", obj, obj.GetName(), err)
		}
	}
}

func TestKubeVirtCCMAddonSelection(t *testing.T) {
	t.Parallel()

	s := &state.State{
		Cluster: &kubeoneapi.KubeOneCluster{
			CloudProvider: kubeoneapi.CloudProviderSpec{
				External: true,
				Kubevirt: &kubeoneapi.KubevirtSpec{InfraNamespace: "infra-ns"},
			},
		},
	}

	actions := ensureCCMAddons(s, nil)
	if len(actions) != 1 || actions[0].name != resources.AddonCCMKubeVirt {
		t.Fatalf("got %+v, want only %s", actions, resources.AddonCCMKubeVirt)
	}

	if actions[0].postFn == nil {
		t.Error("ccm-kubevirt must deploy the infra side in a postFn")
	}
}
