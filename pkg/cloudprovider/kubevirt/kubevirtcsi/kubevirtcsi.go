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

// Package kubevirtcsi supports the csi-kubevirt addon, which deploys the KubeVirt CSI driver through
// kubevirt-csi-driver-operator. It only talks to the tenant cluster.
package kubevirtcsi

import (
	"context"
	"strconv"
	"strings"
	"time"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// driverName is the name of the KubeVirt CSI driver, and the provisioner of its StorageClasses.
	driverName = "csi.kubevirt.io"

	// tenantGroup and tenantName identify the Tenant reconciled by kubevirt-csi-driver-operator. The operator only
	// reconciles the Tenant with this name.
	tenantGroup = "csiprovisioner.kubevirt.io"
	tenantName  = "tenant"

	// defaultVolumeSnapshotClassAnnotation is set by kubevirt-csi-driver-operator on every VolumeSnapshotClass it
	// creates, to "true" or "false".
	defaultVolumeSnapshotClassAnnotation = "snapshot.storage.kubernetes.io/is-default-class"

	// classNamePrefix is prepended by kubevirt-csi-driver-operator to the infra class name to name the StorageClass or
	// VolumeSnapshotClass it creates.
	classNamePrefix = "kubevirt-"

	// nodePluginDeleteTimeout is how long to wait for the node plugin DaemonSet of the old csi-kubevirt addon, and
	// its pods, to be deleted.
	nodePluginDeleteTimeout = 5 * time.Minute
)

var (
	storageClassListGVK = storagev1.SchemeGroupVersion.WithKind("StorageClassList")

	volumeSnapshotClassGVK = schema.GroupVersionKind{
		Group:   "snapshot.storage.k8s.io",
		Version: "v1",
		Kind:    "VolumeSnapshotClass",
	}
	volumeSnapshotClassListGVK = volumeSnapshotClassGVK.GroupVersion().WithKind("VolumeSnapshotClassList")
)

// MigrateNodePlugin removes the node plugin of the csi-kubevirt addon from before kubevirt-csi-driver-operator. The
// operator runs its own node plugin in the kubevirt-csi-driver namespace, which uses the same host paths, so both must
// not run at the same time on a node. It waits until the old pods are gone. It must run before the addon is applied.
//
// The other objects are deleted by name, because the addon prune only removes them when the addon inventory exists,
// and clusters last applied with kubectl don't have one.
func MigrateNodePlugin(s *state.State) error {
	ctx, c := s.Context, s.DynamicClient

	nodeDaemonSet := &appsv1.DaemonSet{
		Name: "kubevirt-csi-node", Namespace: metav1.NamespaceSystem,
	}
	if err := clientutil.DeleteAndWait(ctx, c, nodeDaemonSet, nodePluginDeleteTimeout); err != nil {
		return err
	}

	leftovers := []client.Object{
		&corev1.ServiceAccount{Name: "kubevirt-csi-node-sa", Namespace: metav1.NamespaceSystem},
		&rbacv1.ClusterRole{Name: "kubevirt-csi-node-cr"},
		&rbacv1.ClusterRoleBinding{Name: "kubevirt-csi-node-binding"},
		&corev1.ConfigMap{Name: "driver-config", Namespace: metav1.NamespaceSystem},
	}
	for _, obj := range leftovers {
		if err := clientutil.DeleteIfExists(ctx, c, obj); err != nil {
			return err
		}
	}

	return nil
}

// ReconcileClasses makes the StorageClasses and VolumeSnapshotClasses created by kubevirt-csi-driver-operator match
// the manifest. It must run after the Tenant is applied.
//
// The operator updates only some fields of an existing StorageClass, and fails on every reconcile when an immutable
// field changed. It also never deletes classes removed from the Tenant. Every StorageClass owned by the Tenant that
// differs from the manifest is deleted, and the operator recreates it from the Tenant. All of them are deleted before
// waiting, because the operator stops at the first StorageClass it can't update.
//
// The operator updates all fields of an existing VolumeSnapshotClass, so only VolumeSnapshotClasses removed from the
// Tenant are deleted. It reconciles VolumeSnapshotClasses after all StorageClasses, so they are waited for last.
func ReconcileClasses(s *state.State) error {
	kubevirt := s.Cluster.CloudProvider.Kubevirt
	if kubevirt == nil {
		return nil
	}

	desiredSCs := desiredStorageClasses(kubevirt.StorageClasses)
	desiredVSCs := desiredVolumeSnapshotClasses(kubevirt.VolumeSnapshotClasses)

	if err := deleteStaleStorageClasses(s.Context, s.DynamicClient, desiredSCs); err != nil {
		return err
	}

	if err := deleteStaleVolumeSnapshotClasses(s.Context, s.DynamicClient, desiredVSCs); err != nil {
		return err
	}

	return nil
}

// desiredStorageClasses returns the StorageClasses kubevirt-csi-driver-operator creates for the given entries, keyed by
// name, with the defaults the API server applies.
func desiredStorageClasses(storageClasses []kubeoneapi.KubevirtStorageClass) map[string]*storagev1.StorageClass {
	desired := make(map[string]*storagev1.StorageClass, len(storageClasses))

	for _, sc := range storageClasses {
		obj := &storagev1.StorageClass{
			Name:        classNamePrefix + sc.InfraStorageClassName,
			Labels:      sc.Labels,
			Provisioner: driverName,
			Parameters: map[string]string{
				"infraStorageClassName": sc.InfraStorageClassName,
				"bus":                   sc.Bus,
			},
			ReclaimPolicy:        new(corev1.PersistentVolumeReclaimDelete),
			AllowVolumeExpansion: new(sc.AllowVolumeExpansion),
			VolumeBindingMode:    new(storagev1.VolumeBindingImmediate),
		}

		if sc.ReclaimPolicy != "" {
			obj.ReclaimPolicy = new(sc.ReclaimPolicy)
		}
		if sc.VolumeBindingMode != "" {
			obj.VolumeBindingMode = new(sc.VolumeBindingMode)
		}

		// Same order as kubevirt-csi-driver-operator: zones first, then regions.
		if len(sc.Zones) > 0 {
			obj.AllowedTopologies = append(obj.AllowedTopologies, topologyTerm(corev1.LabelTopologyZone, sc.Zones))
		}
		if len(sc.Regions) > 0 {
			obj.AllowedTopologies = append(obj.AllowedTopologies, topologyTerm(corev1.LabelTopologyRegion, sc.Regions))
		}

		desired[obj.Name] = obj
	}

	return desired
}

func topologyTerm(key string, values []string) corev1.TopologySelectorTerm {
	return corev1.TopologySelectorTerm{
		MatchLabelExpressions: []corev1.TopologySelectorLabelRequirement{
			{Key: key, Values: values},
		},
	}
}

// desiredVolumeSnapshotClasses returns the VolumeSnapshotClasses kubevirt-csi-driver-operator creates for the given
// entries, keyed by name.
func desiredVolumeSnapshotClasses(volumeSnapshotClasses []kubeoneapi.KubevirtVolumeSnapshotClass) map[string]*unstructured.Unstructured {
	desired := make(map[string]*unstructured.Unstructured, len(volumeSnapshotClasses))

	for _, vsc := range volumeSnapshotClasses {
		deletionPolicy := "Delete"
		if vsc.DeletionPolicy != "" {
			deletionPolicy = vsc.DeletionPolicy
		}

		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"driver":         driverName,
				"deletionPolicy": deletionPolicy,
				"parameters": map[string]any{
					"infraSnapshotClassName": vsc.InfraVolumeSnapshotClass,
				},
			},
		}
		obj.SetGroupVersionKind(volumeSnapshotClassGVK)
		obj.SetName(classNamePrefix + vsc.InfraVolumeSnapshotClass)
		obj.SetAnnotations(map[string]string{
			defaultVolumeSnapshotClassAnnotation: strconv.FormatBool(vsc.IsDefaultClass),
		})

		desired[obj.GetName()] = obj
	}

	return desired
}

// storageClassMatches reports whether the fields of the StorageClass that kubevirt-csi-driver-operator doesn't update
// match the desired ones. The default class annotation is not compared, because the operator updates it.
func storageClassMatches(current, desired *storagev1.StorageClass) bool {
	return current.Provisioner == desired.Provisioner &&
		equality.Semantic.DeepEqual(current.Parameters, desired.Parameters) &&
		equality.Semantic.DeepEqual(current.Labels, desired.Labels) &&
		equality.Semantic.DeepEqual(current.AllowedTopologies, desired.AllowedTopologies) &&
		ptr.Deref(current.ReclaimPolicy, corev1.PersistentVolumeReclaimDelete) == *desired.ReclaimPolicy &&
		ptr.Deref(current.VolumeBindingMode, storagev1.VolumeBindingImmediate) == *desired.VolumeBindingMode &&
		ptr.Deref(current.AllowVolumeExpansion, false) == *desired.AllowVolumeExpansion
}

// unstructuredStorageClassMatches is storageClassMatches for a StorageClass read as unstructured.
func unstructuredStorageClassMatches(obj *unstructured.Unstructured, desired *storagev1.StorageClass) (bool, error) {
	current := &storagev1.StorageClass{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, current); err != nil {
		return false, fail.Runtime(err, "converting StorageClass %s", obj.GetName())
	}

	return storageClassMatches(current, desired), nil
}

// ownedByTenant reports whether kubevirt-csi-driver-operator created the object.
func ownedByTenant(obj metav1.Object) bool {
	ref := metav1.GetControllerOfNoCopy(obj)
	if ref == nil {
		return false
	}

	return ref.Kind == "Tenant" && ref.Name == tenantName && strings.HasPrefix(ref.APIVersion, tenantGroup+"/")
}

// deleteStaleStorageClasses deletes the StorageClasses owned by the Tenant that are not desired, or that differ from
// the desired ones.
func deleteStaleStorageClasses(ctx context.Context, dynClient client.Client, desired map[string]*storagev1.StorageClass) error {
	return deleteStaleClasses(ctx, dynClient, storageClassListGVK, func(obj *unstructured.Unstructured) (bool, error) {
		want, ok := desired[obj.GetName()]
		if !ok {
			return false, nil
		}

		return unstructuredStorageClassMatches(obj, want)
	})
}

// deleteStaleVolumeSnapshotClasses deletes the VolumeSnapshotClasses owned by the Tenant that are not desired.
// kubevirt-csi-driver-operator updates all fields of the desired ones itself.
func deleteStaleVolumeSnapshotClasses(ctx context.Context, dynClient client.Client, desired map[string]*unstructured.Unstructured) error {
	return deleteStaleClasses(ctx, dynClient, volumeSnapshotClassListGVK, func(obj *unstructured.Unstructured) (bool, error) {
		_, ok := desired[obj.GetName()]

		return ok, nil
	})
}

// deleteStaleClasses deletes every object of the given list kind that is owned by the Tenant and that keep doesn't
// keep.
func deleteStaleClasses(ctx context.Context, dynClient client.Client, listGVK schema.GroupVersionKind, keep func(*unstructured.Unstructured) (bool, error)) error {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(listGVK)
	if err := dynClient.List(ctx, list); err != nil {
		// Without the CRD, e.g. the VolumeSnapshotClass one, there is nothing to delete.
		if meta.IsNoMatchError(err) || k8serrors.IsNotFound(err) {
			return nil
		}

		return fail.KubeClient(err, "listing %s", listGVK.Kind)
	}

	for i := range list.Items {
		current := &list.Items[i]
		if !ownedByTenant(current) {
			continue
		}

		ok, err := keep(current)
		if err != nil {
			return err
		}
		if ok {
			continue
		}

		if err := clientutil.DeleteIfExists(ctx, dynClient, current); err != nil {
			return err
		}
	}

	return nil
}
