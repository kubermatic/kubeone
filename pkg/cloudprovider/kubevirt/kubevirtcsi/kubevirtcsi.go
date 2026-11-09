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
	"maps"
	"slices"
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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
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

	// classNamePrefix is prepended by kubevirt-csi-driver-operator to the infra class name to name the StorageClass or
	// VolumeSnapshotClass it creates.
	classNamePrefix = "kubevirt-"

	// nodePluginDeleteTimeout is how long to wait for the node plugin DaemonSet of the old csi-kubevirt addon, and
	// its pods, to be deleted.
	nodePluginDeleteTimeout = 5 * time.Minute

	// classesTimeout is how long to wait for kubevirt-csi-driver-operator to create the StorageClasses. It covers a
	// fresh install, where the operator image has to be pulled first.
	classesTimeout = 5 * time.Minute

	classesPollInterval = 2 * time.Second
)

var volumeSnapshotClassListGVK = schema.GroupVersionKind{
	Group:   "snapshot.storage.k8s.io",
	Version: "v1",
	Kind:    "VolumeSnapshotClassList",
}

// MigrateNodePlugin removes the node plugin of the csi-kubevirt addon from before kubevirt-csi-driver-operator. The
// operator runs its own node plugin in the kubevirt-csi-driver namespace, which uses the same host paths, so both must
// not run at the same time on a node. It waits until the old pods are gone. It must run before the addon is applied.
//
// The other objects are deleted by name, because the addon prune only removes them when the addon inventory exists,
// and clusters last applied with kubectl don't have one.
func MigrateNodePlugin(s *state.State) error {
	ctx, c := s.Context, s.DynamicClient

	nodeDaemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "kubevirt-csi-node", Namespace: metav1.NamespaceSystem},
	}
	if err := clientutil.DeleteAndWait(ctx, c, nodeDaemonSet, nodePluginDeleteTimeout); err != nil {
		return err
	}

	leftovers := []client.Object{
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "kubevirt-csi-node-sa", Namespace: metav1.NamespaceSystem}},
		&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "kubevirt-csi-node-cr"}},
		&rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "kubevirt-csi-node-binding"}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "driver-config", Namespace: metav1.NamespaceSystem}},
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
func ReconcileClasses(s *state.State) error {
	kubevirt := s.Cluster.CloudProvider.Kubevirt
	if kubevirt == nil {
		return nil
	}

	desired := desiredStorageClasses(kubevirt.StorageClasses)

	if err := deleteStaleStorageClasses(s.Context, s.DynamicClient, desired); err != nil {
		return err
	}

	if err := deleteStaleVolumeSnapshotClasses(s.Context, s.DynamicClient, kubevirt.VolumeSnapshotClasses); err != nil {
		return err
	}

	return waitForStorageClasses(s.Context, s.DynamicClient, desired, classesTimeout)
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

// ownedByTenant reports whether kubevirt-csi-driver-operator created the object.
func ownedByTenant(obj metav1.Object) bool {
	ref := metav1.GetControllerOfNoCopy(obj)
	if ref == nil {
		return false
	}

	return ref.Kind == "Tenant" && ref.Name == tenantName && strings.HasPrefix(ref.APIVersion, tenantGroup+"/")
}

func deleteStaleStorageClasses(ctx context.Context, c client.Client, desired map[string]*storagev1.StorageClass) error {
	var list storagev1.StorageClassList
	if err := c.List(ctx, &list); err != nil {
		return fail.KubeClient(err, "listing StorageClasses")
	}

	for i := range list.Items {
		current := &list.Items[i]
		if !ownedByTenant(current) {
			continue
		}

		if want, ok := desired[current.Name]; ok && storageClassMatches(current, want) {
			continue
		}

		if err := clientutil.DeleteIfExists(ctx, c, current); err != nil {
			return err
		}
	}

	return nil
}

func deleteStaleVolumeSnapshotClasses(ctx context.Context, c client.Client, volumeSnapshotClasses []kubeoneapi.KubevirtVolumeSnapshotClass) error {
	desired := make(map[string]struct{}, len(volumeSnapshotClasses))
	for _, vsc := range volumeSnapshotClasses {
		desired[classNamePrefix+vsc.InfraVolumeSnapshotClass] = struct{}{}
	}

	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(volumeSnapshotClassListGVK)
	if err := c.List(ctx, list); err != nil {
		// Without the VolumeSnapshotClass CRD there is nothing to delete.
		if meta.IsNoMatchError(err) || k8serrors.IsNotFound(err) {
			return nil
		}

		return fail.KubeClient(err, "listing VolumeSnapshotClasses")
	}

	for i := range list.Items {
		current := &list.Items[i]
		if !ownedByTenant(current) {
			continue
		}

		if _, ok := desired[current.GetName()]; ok {
			continue
		}

		if err := clientutil.DeleteIfExists(ctx, c, current); err != nil {
			return err
		}
	}

	return nil
}

func waitForStorageClasses(ctx context.Context, c client.Client, desired map[string]*storagev1.StorageClass, timeout time.Duration) error {
	if len(desired) == 0 {
		return nil
	}

	names := slices.Sorted(maps.Keys(desired))
	var pending []string

	err := wait.PollUntilContextTimeout(ctx, classesPollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		pending = pending[:0]

		for _, name := range names {
			current := &storagev1.StorageClass{}
			err := c.Get(ctx, client.ObjectKey{Name: name}, current)
			switch {
			case k8serrors.IsNotFound(err):
				pending = append(pending, name)
			case err != nil:
				return false, err
			case !storageClassMatches(current, desired[name]):
				pending = append(pending, name)
			}
		}

		return len(pending) == 0, nil
	})

	return fail.KubeClient(err,
		"waiting for kubevirt-csi-driver-operator to create StorageClasses %v, check the logs of the kubevirt-csi-driver/kubevirt-csi-driver-operator Deployment",
		pending)
}
