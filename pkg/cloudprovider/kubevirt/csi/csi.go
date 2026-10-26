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

// Package csi deploys the KubeVirt CSI driver: the Tenant CR consumed by
// kubevirt-csi-driver-operator in the tenant cluster, and the CSI controller
// running in the infra cluster.
package csi

import (
	"context"
	"maps"
	"time"

	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/clientutil"
	kubevirtinfra "k8c.io/kubeone/pkg/cloudprovider/kubevirt/infra"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// TenantServiceAccountName is the tenant cluster ServiceAccount of the CSI controller.
	TenantServiceAccountName = "kubevirt-csi-controller"
	// TenantTokenSecretName is the token Secret of TenantServiceAccountName.
	TenantTokenSecretName = "kubevirt-csi-controller-token"
	// TenantNamespace is the tenant cluster namespace of TenantServiceAccountName.
	TenantNamespace = "kube-system"

	tenantCRDName = "tenants.csiprovisioner.kubevirt.io"
	tenantName    = "tenant"

	crdWaitInterval = 2 * time.Second
	crdWaitTimeout  = 2 * time.Minute
)

// Component is the infra cluster identity of the CSI controller. The driver
// uses in-cluster configuration, so no infra kubeconfig is generated.
var Component = kubevirtinfra.Component{
	Name: "kubevirt-csi-controller",
	Rules: []rbacv1.PolicyRule{
		{
			APIGroups: []string{"cdi.kubevirt.io"},
			Resources: []string{"datavolumes"},
			Verbs:     []string{"get", "create", "delete"},
		},
		{
			APIGroups: []string{"kubevirt.io"},
			Resources: []string{"virtualmachines", "virtualmachineinstances"},
			Verbs:     []string{"get", "list"},
		},
		{
			APIGroups: []string{"subresources.kubevirt.io"},
			Resources: []string{
				"virtualmachines/addvolume",
				"virtualmachines/removevolume",
				"virtualmachineinstances/addvolume",
				"virtualmachineinstances/removevolume",
			},
			Verbs: []string{"update"},
		},
		{
			APIGroups: []string{"snapshot.storage.k8s.io"},
			Resources: []string{"volumesnapshots"},
			Verbs:     []string{"get", "create", "delete"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"persistentvolumeclaims"},
			Verbs:     []string{"get", "list", "watch", "update", "patch"},
		},
	},
}

// Ensure deploys everything the tenant addon cannot: the Tenant CR (its CRD
// is applied by the same addon and has to be established first), the CSI
// controller identities and the CSI controller Deployment in the infra cluster.
func Ensure(s *state.State) error {
	if s.DynamicClient == nil {
		return fail.NoKubeClient()
	}

	ctx := s.Context
	spec := s.Cluster.CloudProvider.Kubevirt

	if err := waitCRDEstablished(ctx, s.DynamicClient, tenantCRDName); err != nil {
		return err
	}

	if err := clientutil.CreateOrReplace(ctx, s.DynamicClient, Tenant(spec)); err != nil {
		return err
	}

	if err := reconcileStorageClasses(ctx, s.DynamicClient, s.Logger, spec); err != nil {
		return err
	}

	infraClient, namespace, err := kubevirtinfra.Client(s)
	if err != nil {
		return err
	}

	if err = Component.EnsureIdentity(ctx, infraClient, s.Cluster.Name, namespace); err != nil {
		return err
	}

	err = Component.EnsureTenantKubeconfig(
		ctx,
		s.DynamicClient,
		infraClient,
		s.Cluster.Name,
		namespace,
		kubevirtinfra.TenantServer(s.Cluster.APIEndpoint),
		ctrlrt.ObjectKey{Namespace: TenantNamespace, Name: TenantTokenSecretName},
	)
	if err != nil {
		return err
	}

	return clientutil.CreateOrReplace(ctx, infraClient, Deployment(s.Cluster.Name, namespace, s.Images.Get))
}

// StorageClassName returns the name of the tenant StorageClass that
// kubevirt-csi-driver-operator creates for the infra StorageClass.
func StorageClassName(infraStorageClassName string) string {
	return "kubevirt-" + infraStorageClassName
}

// reconcileStorageClasses updates the mutable fields of existing tenant
// StorageClasses that kubevirt-csi-driver-operator sets only on creation:
// allowVolumeExpansion and labels. Labels are added or updated, but never
// removed, since other labels might be set by someone else. StorageClasses
// that don't exist yet are left to the operator.
//
// Bus, volumeBindingMode and reclaimPolicy can't be changed on an existing
// StorageClass; changes to them are only reported.
func reconcileStorageClasses(ctx context.Context, c ctrlrt.Client, logger logrus.FieldLogger, spec *kubeoneapi.KubevirtSpec) error {
	for _, want := range spec.StorageClasses {
		sc := &storagev1.StorageClass{}
		key := ctrlrt.ObjectKey{Name: StorageClassName(want.InfraStorageClassName)}

		if err := c.Get(ctx, key, sc); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}

			return fail.KubeClient(err, "getting StorageClass %s", key.Name)
		}

		for _, field := range immutableStorageClassChanges(sc, want) {
			logger.Warnf("StorageClass %q: %s can't be changed on an existing StorageClass, delete it to have it recreated by the operator", sc.Name, field)
		}

		patch := ctrlrt.MergeFrom(sc.DeepCopy())
		sc.AllowVolumeExpansion = new(want.AllowVolumeExpansion)
		if len(want.Labels) > 0 && sc.Labels == nil {
			sc.Labels = map[string]string{}
		}
		maps.Copy(sc.Labels, want.Labels)

		if err := c.Patch(ctx, sc, patch); err != nil {
			return fail.KubeClient(err, "patching StorageClass %s", sc.Name)
		}
	}

	return nil
}

// immutableStorageClassChanges returns the immutable fields of the existing
// StorageClass that differ from the configuration.
func immutableStorageClassChanges(sc *storagev1.StorageClass, want kubeoneapi.KubevirtInfraStorageClass) []string {
	var changed []string

	if sc.Parameters["bus"] != storageClassBus(want) {
		changed = append(changed, "bus")
	}

	wantBindingMode := storagev1.VolumeBindingImmediate
	if want.VolumeBindingMode != nil {
		wantBindingMode = *want.VolumeBindingMode
	}
	if sc.VolumeBindingMode != nil && *sc.VolumeBindingMode != wantBindingMode {
		changed = append(changed, "volumeBindingMode")
	}

	wantReclaimPolicy := corev1.PersistentVolumeReclaimDelete
	if want.ReclaimPolicy != "" {
		wantReclaimPolicy = want.ReclaimPolicy
	}
	if sc.ReclaimPolicy != nil && *sc.ReclaimPolicy != wantReclaimPolicy {
		changed = append(changed, "reclaimPolicy")
	}

	return changed
}

func storageClassBus(sc kubeoneapi.KubevirtInfraStorageClass) string {
	if sc.Bus == "" {
		return "scsi"
	}

	return sc.Bus
}

func waitCRDEstablished(ctx context.Context, c ctrlrt.Client, name string) error {
	condFn := clientutil.CRDsReadyCondition(ctx, c, []string{name})

	err := wait.PollUntilContextTimeout(ctx, crdWaitInterval, crdWaitTimeout, true, func(context.Context) (bool, error) {
		ok, err := condFn()

		return ok, ctrlrt.IgnoreNotFound(err)
	})

	return fail.KubeClient(err, "waiting for CRD %s to be established", name)
}

// Tenant returns the Tenant CR reconciled by kubevirt-csi-driver-operator.
func Tenant(spec *kubeoneapi.KubevirtSpec) *unstructured.Unstructured {
	storageClasses := []any{}
	for _, sc := range spec.StorageClasses {
		obj := map[string]any{
			"infraStorageClassName": sc.InfraStorageClassName,
			"isDefaultClass":        sc.IsDefaultClass,
			"bus":                   storageClassBus(sc),
		}
		if sc.VolumeBindingMode != nil {
			obj["volumeBindingMode"] = string(*sc.VolumeBindingMode)
		}
		if sc.AllowVolumeExpansion {
			obj["allowVolumeExpansion"] = true
		}
		if sc.ReclaimPolicy != "" {
			obj["reclaimPolicy"] = string(sc.ReclaimPolicy)
		}
		if len(sc.Zones) > 0 {
			obj["zones"] = toAnySlice(sc.Zones)
		}
		if len(sc.Regions) > 0 {
			obj["regions"] = toAnySlice(sc.Regions)
		}
		if len(sc.Labels) > 0 {
			labels := map[string]any{}
			for k, v := range sc.Labels {
				labels[k] = v
			}
			obj["labels"] = labels
		}

		storageClasses = append(storageClasses, obj)
	}

	snapshotClasses := []any{}
	for _, vsc := range spec.VolumeSnapshotClasses {
		obj := map[string]any{
			"infraVolumeSnapshotClass": vsc.InfraVolumeSnapshotClass,
			"isDefaultClass":           vsc.IsDefaultClass,
		}
		if vsc.DeletionPolicy != "" {
			obj["deletionPolicy"] = vsc.DeletionPolicy
		}

		snapshotClasses = append(snapshotClasses, obj)
	}

	tenant := &unstructured.Unstructured{
		Object: map[string]any{
			"spec": map[string]any{
				"storageClasses":        storageClasses,
				"volumeSnapshotClasses": snapshotClasses,
			},
		},
	}
	tenant.SetAPIVersion("csiprovisioner.kubevirt.io/v1alpha1")
	tenant.SetKind("Tenant")
	tenant.SetName(tenantName)

	return tenant
}

func toAnySlice(in []string) []any {
	out := make([]any, 0, len(in))
	for _, s := range in {
		out = append(out, s)
	}

	return out
}
