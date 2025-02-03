/*
Copyright 2025 The KubeOne Authors.

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

package tasks

import (
	"maps"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func cleanupStaleObjects(st *state.State) error {
	st.Logger.Infoln("Cleanup stale objects...")

	var cleanupObjects []crclient.Object
	cleanupObjects = append(cleanupObjects, kuredObjects()...)

NextObject:
	for _, obj := range cleanupObjects {
		originalLabels := maps.Clone(obj.GetLabels())
		obj.SetLabels(map[string]string{})

		err := st.DynamicClient.Get(st.Context, crclient.ObjectKeyFromObject(obj), obj)
		switch {
		case apierrors.IsNotFound(err):
			continue NextObject
		case err != nil:
			return fail.KubeClient(err, "checking stale object %s %q", obj.GetObjectKind().GroupVersionKind().String(), crclient.ObjectKeyFromObject(obj))
		}

		realLabels := obj.GetLabels()

		// compare requested labels to the real of, if not match -> let the object live
		for cleanupKey, cleanupValue := range originalLabels {
			if val, ok := realLabels[cleanupKey]; !ok || val != cleanupValue {
				st.Logger.Debugf("skip deleting object as labels are different: %s %q", obj.GetObjectKind().GroupVersionKind().String(), crclient.ObjectKeyFromObject(obj))

				continue NextObject
			}
		}

		if err := st.DynamicClient.Delete(st.Context, obj); crclient.IgnoreNotFound(err) != nil {
			return fail.KubeClient(err, "deleting stale object %s %q", obj.GetObjectKind().GroupVersionKind().String(), crclient.ObjectKeyFromObject(obj))
		}

		st.Logger.Debugf("deleted stale object %s %q", obj.GetObjectKind().GroupVersionKind().String(), crclient.ObjectKeyFromObject(obj))
	}

	return nil
}

func kuredObjects() []crclient.Object {
	labels := map[string]string{"kubeone.io/addon": "unattended-upgrades"}
	unstructuredLabels := withLabels(labels)

	cleanupObjects := []crclient.Object{
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "kured",
				Labels: labels,
			},
		},
		newUnstructured(
			"rbac.authorization.k8s.io/v1",
			"ClusterRoleBinding",
			crclient.ObjectKey{Name: "kured"},
			unstructuredLabels,
		),
		newUnstructured(
			"rbac.authorization.k8s.io/v1",
			"Role",
			crclient.ObjectKey{Name: "kured", Namespace: "kube-system"},
			unstructuredLabels,
		),
		newUnstructured(
			"rbac.authorization.k8s.io/v1",
			"RoleBinding",
			crclient.ObjectKey{Name: "kured", Namespace: "kube-system"},
			unstructuredLabels,
		),
		newUnstructured(
			"v1",
			"ServiceAccount",
			crclient.ObjectKey{Name: "kured", Namespace: "kube-system"},
			unstructuredLabels,
		),
	}

	return cleanupObjects
}

func newUnstructured(apiVersion string, kind string, identity crclient.ObjectKey, opts ...func(*metav1unstructured.Unstructured)) crclient.Object {
	obj := &metav1unstructured.Unstructured{}
	obj.SetAPIVersion(apiVersion)
	obj.SetKind(kind)
	obj.SetName(identity.Name)
	obj.SetNamespace(identity.Namespace)

	for _, opt := range opts {
		opt(obj)
	}

	return obj
}

func withLabels(labels map[string]string) func(*metav1unstructured.Unstructured) {
	return func(u *metav1unstructured.Unstructured) {
		u.SetLabels(labels)
	}
}
