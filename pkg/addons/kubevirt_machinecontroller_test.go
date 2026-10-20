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

	"github.com/sirupsen/logrus"

	embeddedaddons "k8c.io/kubeone/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubevirtmc "k8c.io/kubeone/pkg/cloudprovider/kubevirt/machinecontroller"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/images"
	"k8c.io/kubeone/pkg/templates/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

func renderMachineControllerAddon(t *testing.T, cloudProvider kubeoneapi.CloudProviderSpec) map[string][]string {
	t.Helper()

	cluster := &kubeoneapi.KubeOneCluster{
		Name:                  "kubeone-test",
		CloudProvider:         cloudProvider,
		RegistryConfiguration: &kubeoneapi.RegistryConfiguration{},
	}

	applier := &applier{
		TemplateData: templateData{
			Config: cluster,
			Certificates: map[string]string{
				"KubernetesCA":                 "ca",
				"MachineControllerWebhookCert": "cert",
				"MachineControllerWebhookKey":  "key",
			},
			InternalImages: &internalImages{
				resolver: func(images.Resource, ...images.GetOpt) string { return "img:test" },
			},
			Resources: map[string]string{"CABundleSSLCertFilePath": "/etc/kubeone/certs/ca-certificates.crt"},
		},
	}

	manifests, err := applier.loadAddonsManifests(embeddedaddons.FS, resources.AddonMachineController, nil, logrus.New(), false, cluster, false)
	if err != nil {
		t.Fatalf("unable to render manifests: %v", err)
	}

	objects := map[string][]string{}
	for _, m := range manifests {
		var obj metav1.PartialObjectMetadata
		if err := yaml.Unmarshal(m.Raw, &obj); err != nil {
			t.Fatal(err)
		}

		objects[obj.Kind] = append(objects[obj.Kind], obj.Name)
	}

	return objects
}

func TestMachineControllerAddonKubeVirt(t *testing.T) {
	t.Parallel()

	objects := renderMachineControllerAddon(t, kubeoneapi.CloudProviderSpec{
		External: true,
		Kubevirt: &kubeoneapi.KubevirtSpec{InfraNamespace: "infra-ns"},
	})

	for _, kind := range []string{"Deployment", "Service", "PodDisruptionBudget", "MutatingWebhookConfiguration"} {
		if len(objects[kind]) != 0 {
			t.Errorf("on KubeVirt, %s %v must not be in the tenant addon", kind, objects[kind])
		}
	}

	wantSecrets := map[string]bool{
		kubevirtmc.ControllerTenantTokenSecretName: false,
		kubevirtmc.WebhookTenantTokenSecretName:    false,
	}
	for _, name := range objects["Secret"] {
		if _, ok := wantSecrets[name]; !ok {
			t.Errorf("unexpected Secret %s", name)
		}
		wantSecrets[name] = true
	}

	for name, found := range wantSecrets {
		if !found {
			t.Errorf("token Secret %s not found", name)
		}
	}

	if len(objects["CustomResourceDefinition"]) != 3 || len(objects["ServiceAccount"]) != 2 {
		t.Errorf("CRDs and ServiceAccounts must stay in the tenant cluster: %v", objects)
	}
}

func TestMachineControllerAddonOtherProviders(t *testing.T) {
	t.Parallel()

	objects := renderMachineControllerAddon(t, kubeoneapi.CloudProviderSpec{
		External: true,
		Hetzner:  &kubeoneapi.HetznerSpec{},
	})

	if len(objects["Deployment"]) != 2 || len(objects["Service"]) != 1 || len(objects["MutatingWebhookConfiguration"]) != 1 || len(objects["PodDisruptionBudget"]) != 2 {
		t.Errorf("machine-controller must still run in the cluster on other providers: %v", objects)
	}

	for _, name := range objects["Secret"] {
		if name == kubevirtmc.ControllerTenantTokenSecretName || name == kubevirtmc.WebhookTenantTokenSecretName {
			t.Errorf("token Secret %s must only exist on KubeVirt", name)
		}
	}
}

func TestMigrateKubeVirtMachineController(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	ns := metav1.NamespaceSystem
	legacy := []client.Object{
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "machine-controller", Namespace: ns}},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "machine-controller-webhook", Namespace: ns}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "machine-controller-webhook", Namespace: ns}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "machinecontroller-webhook-serving-cert", Namespace: ns}},
		&policyv1.PodDisruptionBudget{ObjectMeta: metav1.ObjectMeta{Name: "machine-controller", Namespace: ns}},
		&policyv1.PodDisruptionBudget{ObjectMeta: metav1.ObjectMeta{Name: "machine-controller-webhook", Namespace: ns}},
	}
	kept := []client.Object{
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "machine-controller", Namespace: ns}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: kubevirtmc.ControllerTenantTokenSecretName, Namespace: ns}},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(append(legacy, kept...)...).Build()
	s := &state.State{Context: context.Background(), DynamicClient: c}

	for range 2 { // idempotent
		if err := migrateKubeVirtMachineController(s); err != nil {
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

func TestMachineControllerAddonPostFn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		kubevirt   bool
		wantPostFn bool
	}{
		{name: "kubevirt", kubevirt: true, wantPostFn: true},
		{name: "other provider"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cluster := &kubeoneapi.KubeOneCluster{
				CloudProvider:  kubeoneapi.CloudProviderSpec{DisableBundledCSIDrivers: true},
				ClusterNetwork: kubeoneapi.ClusterNetworkConfig{CNI: &kubeoneapi.CNI{}},
				Features: kubeoneapi.Features{
					CoreDNS:       &kubeoneapi.CoreDNS{DeployPodDisruptionBudget: new(false)},
					MetricsServer: &kubeoneapi.MetricsServer{},
					NodeLocalDNS:  &kubeoneapi.NodeLocalDNS{},
				},
				MachineController:      &kubeoneapi.MachineControllerConfig{Deploy: true},
				OperatingSystemManager: &kubeoneapi.OperatingSystemManagerConfig{},
			}
			if tc.kubevirt {
				cluster.CloudProvider.Kubevirt = &kubeoneapi.KubevirtSpec{InfraNamespace: "infra-ns"}
			} else {
				cluster.CloudProvider.Hetzner = &kubeoneapi.HetznerSpec{}
			}

			var found bool
			for _, a := range collectAddons(&state.State{Cluster: cluster}) {
				if a.name != resources.AddonMachineController {
					continue
				}
				found = true

				if (a.postFn != nil) != tc.wantPostFn {
					t.Errorf("postFn set = %v, want %v", a.postFn != nil, tc.wantPostFn)
				}
			}

			if !found {
				t.Fatal("machinecontroller addon not collected")
			}
		})
	}
}
