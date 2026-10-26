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

package infra

import (
	"context"

	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/fail"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
)

// CleanupSelector selects every infra-side component object of the cluster.
// It requires LabelComponent, so objects that only carry LabelClusterName,
// such as the control plane API server Service, are not selected.
func CleanupSelector(clusterName string) labels.Selector {
	clusterReq, _ := labels.NewRequirement(LabelClusterName, selection.Equals, []string{ClusterLabelValue(clusterName)})
	componentReq, _ := labels.NewRequirement(LabelComponent, selection.Exists, nil)

	return labels.NewSelector().Add(*clusterReq, *componentReq)
}

// clusterCleanupSelector selects every cluster-scoped infra-side component
// object of the cluster in the namespace.
func clusterCleanupSelector(clusterName, namespace string) labels.Selector {
	nsReq, _ := labels.NewRequirement(LabelInfraNamespace, selection.Equals, []string{namespace})

	return CleanupSelector(clusterName).Add(*nsReq)
}

// Cleanup deletes every infra-side component object of the cluster in the
// namespace, and the cluster-scoped objects belonging to it. Workloads are
// deleted before the identities they use.
func Cleanup(ctx context.Context, infraClient ctrlrt.Client, clusterName, namespace string) error {
	lists := []ctrlrt.ObjectList{
		&appsv1.DeploymentList{},
		&corev1.ServiceList{},
		&corev1.ConfigMapList{},
		&corev1.SecretList{},
		&rbacv1.RoleBindingList{},
		&rbacv1.RoleList{},
		&corev1.ServiceAccountList{},
	}

	opts := []ctrlrt.ListOption{
		ctrlrt.InNamespace(namespace),
		ctrlrt.MatchingLabelsSelector{Selector: CleanupSelector(clusterName)},
	}

	if err := deleteListed(ctx, infraClient, lists, opts...); err != nil {
		return err
	}

	clusterLists := []ctrlrt.ObjectList{
		&rbacv1.ClusterRoleBindingList{},
		&rbacv1.ClusterRoleList{},
	}

	err := deleteListed(ctx, infraClient, clusterLists, ctrlrt.MatchingLabelsSelector{Selector: clusterCleanupSelector(clusterName, namespace)})
	// Without permission to list cluster-scoped RBAC, KubeOne could not have created any.
	if apierrors.IsForbidden(err) {
		return nil
	}

	return err
}

func deleteListed(ctx context.Context, infraClient ctrlrt.Client, lists []ctrlrt.ObjectList, opts ...ctrlrt.ListOption) error {
	for _, list := range lists {
		if err := infraClient.List(ctx, list, opts...); err != nil {
			return fail.KubeClient(err, "listing %T", list)
		}

		items, err := meta.ExtractList(list)
		if err != nil {
			return fail.Runtime(err, "extracting %T", list)
		}

		for _, item := range items {
			obj, ok := item.(ctrlrt.Object)
			if !ok {
				continue
			}

			if err := clientutil.DeleteIfExists(ctx, infraClient, obj); err != nil {
				return err
			}
		}
	}

	return nil
}
