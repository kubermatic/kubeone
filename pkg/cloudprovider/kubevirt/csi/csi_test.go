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

package csi

import (
	"context"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubevirtinfra "k8c.io/kubeone/pkg/cloudprovider/kubevirt/infra"
	"k8c.io/kubeone/pkg/templates/images"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func fakeImage(res images.Resource, _ ...images.GetOpt) string {
	return "registry.example.com/" + res.String()
}

func TestDeployment(t *testing.T) {
	d := Deployment("c1", "infra-ns", fakeImage)

	if d.Name != "c1-kubevirt-csi-controller" || d.Namespace != "infra-ns" {
		t.Errorf("unexpected name %s/%s", d.Namespace, d.Name)
	}

	if d.Labels[kubevirtinfra.LabelClusterName] != "c1" || d.Labels[kubevirtinfra.LabelComponent] != Component.Name {
		t.Errorf("missing cleanup labels: %v", d.Labels)
	}

	for k, v := range d.Spec.Selector.MatchLabels {
		if d.Spec.Template.Labels[k] != v {
			t.Errorf("selector label %s=%s not on pod template", k, v)
		}
	}

	pod := d.Spec.Template.Spec
	if pod.ServiceAccountName != "c1-kubevirt-csi-controller" {
		t.Errorf("ServiceAccountName = %q", pod.ServiceAccountName)
	}

	var tenantSecret string
	for _, v := range pod.Volumes {
		if v.Secret != nil {
			tenantSecret = v.Secret.SecretName
		}
	}
	if tenantSecret != "c1-kubevirt-csi-controller-tenant-kubeconfig" {
		t.Errorf("tenant kubeconfig Secret = %q", tenantSecret)
	}

	containers := map[string]corev1.Container{}
	for _, c := range pod.Containers {
		containers[c.Name] = c

		if !strings.HasPrefix(c.Image, "registry.example.com/KubeVirtCSI") {
			t.Errorf("%s: image %q not resolved through the image getter", c.Name, c.Image)
		}
	}

	driver, ok := containers["csi-driver"]
	if !ok {
		t.Fatal("csi-driver container not found")
	}

	for _, want := range []string{
		"--infra-cluster-namespace=infra-ns",
		"--infra-cluster-labels=kubeone_cluster_name=c1",
		"--tenant-cluster-kubeconfig=/var/run/secrets/tenantcluster/kubeconfig",
		"--run-controller-service=true",
		"--run-node-service=false",
	} {
		if !slices.Contains(driver.Args, want) {
			t.Errorf("csi-driver args %v missing %q", driver.Args, want)
		}
	}

	for _, arg := range driver.Args {
		if strings.HasPrefix(arg, "--infra-cluster-kubeconfig") {
			t.Errorf("csi-driver must use in-cluster config for the infra cluster, got %q", arg)
		}
	}

	for _, name := range []string{"csi-provisioner", "csi-attacher", "csi-snapshotter", "csi-resizer"} {
		c, ok := containers[name]
		if !ok {
			t.Errorf("%s container not found", name)

			continue
		}

		if !slices.Contains(c.Args, "--kubeconfig=/var/run/secrets/tenantcluster/kubeconfig") {
			t.Errorf("%s does not use the tenant kubeconfig: %v", name, c.Args)
		}
	}

	if _, ok := containers["csi-liveness-probe"]; !ok {
		t.Error("csi-liveness-probe container not found")
	}
}

func TestComponent(t *testing.T) {
	if Component.NeedsKubeconfig {
		t.Error("the CSI driver uses in-cluster config, no infra kubeconfig must be generated")
	}

	has := func(group, resource, verb string) bool {
		for _, r := range Component.Rules {
			if slices.Contains(r.APIGroups, group) && slices.Contains(r.Resources, resource) && slices.Contains(r.Verbs, verb) {
				return true
			}
		}

		return false
	}

	for _, want := range [][3]string{
		{"cdi.kubevirt.io", "datavolumes", "create"},
		{"cdi.kubevirt.io", "datavolumes", "delete"},
		{"kubevirt.io", "virtualmachineinstances", "get"},
		{"subresources.kubevirt.io", "virtualmachineinstances/addvolume", "update"},
		{"subresources.kubevirt.io", "virtualmachineinstances/removevolume", "update"},
		{"snapshot.storage.k8s.io", "volumesnapshots", "create"},
		{"", "persistentvolumeclaims", "get"},
	} {
		if !has(want[0], want[1], want[2]) {
			t.Errorf("Role does not allow %s on %s.%s", want[2], want[1], want[0])
		}
	}
}

func TestTenant(t *testing.T) {
	wffc := storagev1.VolumeBindingWaitForFirstConsumer
	tenant := Tenant(&kubeoneapi.KubevirtSpec{
		StorageClasses: []kubeoneapi.KubevirtInfraStorageClass{
			{
				InfraStorageClassName: "standard",
				IsDefaultClass:        true,
				VolumeBindingMode:     &wffc,
				AllowVolumeExpansion:  true,
				ReclaimPolicy:         corev1.PersistentVolumeReclaimRetain,
				Zones:                 []string{"z1"},
				Labels:                map[string]string{"a": "b"},
			},
			{InfraStorageClassName: "fast", Bus: "virtio"},
		},
		VolumeSnapshotClasses: []kubeoneapi.KubevirtInfraVolumeSnapshotClass{
			{InfraVolumeSnapshotClass: "snap", DeletionPolicy: "Retain"},
		},
	})

	if tenant.GetName() != "tenant" || tenant.GetKind() != "Tenant" || tenant.GetAPIVersion() != "csiprovisioner.kubevirt.io/v1alpha1" {
		t.Fatalf("unexpected Tenant identity: %s %s %s", tenant.GetAPIVersion(), tenant.GetKind(), tenant.GetName())
	}

	scs, _, _ := unstructured.NestedSlice(tenant.Object, "spec", "storageClasses")
	if len(scs) != 2 {
		t.Fatalf("got %d storage classes, want 2", len(scs))
	}

	first, _ := scs[0].(map[string]any)
	for k, want := range map[string]any{
		"infraStorageClassName": "standard",
		"isDefaultClass":        true,
		"volumeBindingMode":     "WaitForFirstConsumer",
		"allowVolumeExpansion":  true,
		"reclaimPolicy":         "Retain",
		"bus":                   "scsi",
	} {
		if first[k] != want {
			t.Errorf("storageClasses[0].%s = %v, want %v", k, first[k], want)
		}
	}

	second, _ := scs[1].(map[string]any)
	if second["bus"] != "virtio" {
		t.Errorf("storageClasses[1].bus = %v, want virtio", second["bus"])
	}
	if _, ok := second["reclaimPolicy"]; ok {
		t.Error("unset reclaimPolicy must be omitted")
	}

	vscs, _, _ := unstructured.NestedSlice(tenant.Object, "spec", "volumeSnapshotClasses")
	if len(vscs) != 1 {
		t.Fatalf("got %d snapshot classes, want 1", len(vscs))
	}

	// Panics if the object contains non JSON-compatible types.
	_ = tenant.DeepCopy()
}

func TestTenantEmpty(t *testing.T) {
	tenant := Tenant(&kubeoneapi.KubevirtSpec{})

	scs, found, err := unstructured.NestedSlice(tenant.Object, "spec", "storageClasses")
	if err != nil || !found || len(scs) != 0 {
		t.Errorf("storageClasses = %v (found=%v, err=%v), want an empty list", scs, found, err)
	}
}

// operatorStorageClass returns the StorageClass kubevirt-csi-driver-operator
// creates for the configured class, with the API server defaults applied.
func operatorStorageClass(want kubeoneapi.KubevirtInfraStorageClass) *storagev1.StorageClass {
	bindingMode := storagev1.VolumeBindingImmediate
	if want.VolumeBindingMode != nil {
		bindingMode = *want.VolumeBindingMode
	}

	reclaimPolicy := corev1.PersistentVolumeReclaimDelete
	if want.ReclaimPolicy != "" {
		reclaimPolicy = want.ReclaimPolicy
	}

	return &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:   StorageClassName(want.InfraStorageClassName),
			Labels: maps.Clone(want.Labels),
		},
		Provisioner:          "csi.kubevirt.io",
		Parameters:           map[string]string{"infraStorageClassName": want.InfraStorageClassName, "bus": storageClassBus(want)},
		AllowVolumeExpansion: new(want.AllowVolumeExpansion),
		AllowedTopologies:    storageClassAllowedTopologies(want),
		VolumeBindingMode:    &bindingMode,
		ReclaimPolicy:        &reclaimPolicy,
	}
}

func TestReconcileStorageClasses(t *testing.T) {
	ctx := context.Background()
	wffc := storagev1.VolumeBindingWaitForFirstConsumer

	ssd := kubeoneapi.KubevirtInfraStorageClass{InfraStorageClassName: "ssd", ReclaimPolicy: corev1.PersistentVolumeReclaimRetain}
	standard := kubeoneapi.KubevirtInfraStorageClass{
		InfraStorageClassName: "standard",
		AllowVolumeExpansion:  true,
		Labels:                map[string]string{"a": "new", "b": "added"},
		Bus:                   "virtio",
		VolumeBindingMode:     &wffc,
	}
	slow := kubeoneapi.KubevirtInfraStorageClass{
		InfraStorageClassName: "slow",
		AllowVolumeExpansion:  true,
		Labels:                map[string]string{"a": "new"},
		Zones:                 []string{"z1", "z2"},
		Regions:               []string{"r1"},
	}
	unzoned := kubeoneapi.KubevirtInfraStorageClass{InfraStorageClassName: "unzoned"}
	spec := &kubeoneapi.KubevirtSpec{
		StorageClasses: []kubeoneapi.KubevirtInfraStorageClass{
			ssd,
			standard,
			slow,
			unzoned,
			// not created by the operator yet
			{InfraStorageClassName: "fast"},
		},
	}

	// The StorageClasses as created by the operator from the previous configuration.
	existingSSD := operatorStorageClass(kubeoneapi.KubevirtInfraStorageClass{InfraStorageClassName: "ssd"})
	existingStandard := operatorStorageClass(kubeoneapi.KubevirtInfraStorageClass{InfraStorageClassName: "standard"})
	existingStandard.Labels = map[string]string{"a": "old", "foreign": "keep"}
	existingSlow := operatorStorageClass(kubeoneapi.KubevirtInfraStorageClass{InfraStorageClassName: "slow", Zones: []string{"z1"}})
	existingSlow.Labels = map[string]string{"a": "old", "foreign": "keep"}
	existingUnzoned := operatorStorageClass(kubeoneapi.KubevirtInfraStorageClass{InfraStorageClassName: "unzoned", Regions: []string{"r1"}})

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	var deleted []string
	recreate := []kubeoneapi.KubevirtInfraStorageClass{ssd, standard}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(existingSSD, existingStandard, existingSlow, existingUnzoned).
		WithInterceptorFuncs(interceptor.Funcs{
			Delete: func(ctx context.Context, cl ctrlrt.WithWatch, obj ctrlrt.Object, opts ...ctrlrt.DeleteOption) error {
				if err := cl.Delete(ctx, obj, opts...); err != nil {
					return err
				}
				deleted = append(deleted, obj.GetName())

				// The operator stops at the first StorageClass whose update is
				// rejected, so it only recreates them once all are deleted.
				if len(deleted) < len(recreate) {
					return nil
				}
				for _, want := range recreate {
					if err := cl.Create(ctx, operatorStorageClass(want)); err != nil {
						return err
					}
				}

				return nil
			},
		}).
		Build()
	logger, hook := logrustest.NewNullLogger()

	if err := reconcileStorageClasses(ctx, c, logger, spec); err != nil {
		t.Fatal(err)
	}

	if want := []string{"kubevirt-ssd", "kubevirt-standard"}; !slices.Equal(deleted, want) {
		t.Errorf("deleted StorageClasses = %v, want %v", deleted, want)
	}

	got := &storagev1.StorageClass{}
	if err := c.Get(ctx, ctrlrt.ObjectKey{Name: "kubevirt-standard"}, got); err != nil {
		t.Fatal(err)
	}
	if got.Parameters["bus"] != "virtio" || *got.VolumeBindingMode != wffc {
		t.Errorf("bus = %q, volumeBindingMode = %q, want the configured values", got.Parameters["bus"], *got.VolumeBindingMode)
	}
	if got.AllowVolumeExpansion == nil || !*got.AllowVolumeExpansion {
		t.Errorf("allowVolumeExpansion = %v, want true", got.AllowVolumeExpansion)
	}
	if want := map[string]string{"a": "new", "b": "added"}; !maps.Equal(got.Labels, want) {
		t.Errorf("labels = %v, want %v", got.Labels, want)
	}

	if err := c.Get(ctx, ctrlrt.ObjectKey{Name: "kubevirt-ssd"}, got); err != nil {
		t.Fatal(err)
	}
	if *got.ReclaimPolicy != corev1.PersistentVolumeReclaimRetain {
		t.Errorf("reclaimPolicy = %q, want %q", *got.ReclaimPolicy, corev1.PersistentVolumeReclaimRetain)
	}

	if err := c.Get(ctx, ctrlrt.ObjectKey{Name: "kubevirt-slow"}, got); err != nil {
		t.Fatal(err)
	}
	if got.AllowVolumeExpansion == nil || !*got.AllowVolumeExpansion {
		t.Errorf("allowVolumeExpansion = %v, want true", got.AllowVolumeExpansion)
	}
	if want := map[string]string{"a": "new", "foreign": "keep"}; !maps.Equal(got.Labels, want) {
		t.Errorf("labels = %v, want %v", got.Labels, want)
	}
	wantTopologies := []corev1.TopologySelectorTerm{
		{MatchLabelExpressions: []corev1.TopologySelectorLabelRequirement{{Key: corev1.LabelTopologyZone, Values: []string{"z1", "z2"}}}},
		{MatchLabelExpressions: []corev1.TopologySelectorLabelRequirement{{Key: corev1.LabelTopologyRegion, Values: []string{"r1"}}}},
	}
	if !equality.Semantic.DeepEqual(got.AllowedTopologies, wantTopologies) {
		t.Errorf("allowedTopologies = %v, want %v", got.AllowedTopologies, wantTopologies)
	}

	if err := c.Get(ctx, ctrlrt.ObjectKey{Name: "kubevirt-unzoned"}, got); err != nil {
		t.Fatal(err)
	}
	if got.AllowedTopologies != nil {
		t.Errorf("allowedTopologies = %v, want none", got.AllowedTopologies)
	}

	if err := c.Get(ctx, ctrlrt.ObjectKey{Name: "kubevirt-fast"}, &storagev1.StorageClass{}); !apierrors.IsNotFound(err) {
		t.Errorf("missing StorageClass must be left to the operator, got err %v", err)
	}

	var recreated []string
	for _, e := range hook.AllEntries() {
		if e.Level == logrus.InfoLevel && strings.HasPrefix(e.Message, "Recreating") {
			recreated = append(recreated, e.Message)
		}
	}
	if len(recreated) != 2 || !strings.Contains(recreated[0], "reclaimPolicy") ||
		!strings.Contains(recreated[1], "bus, volumeBindingMode") {
		t.Errorf("recreate messages = %q, want reclaimPolicy for ssd and bus, volumeBindingMode for standard", recreated)
	}
}
