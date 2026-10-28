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

package nativeclient

import (
	"context"
	"strings"

	"k8c.io/kubeone/pkg/fail"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// InventoryGroupKindsKey is the inventory ConfigMap key holding the newline-separated GroupKinds of the applied
	// objects, e.g. "ConfigMap" or "Deployment.apps".
	InventoryGroupKindsKey = "groupKinds"

	inventoryManagedByLabel = "app.kubernetes.io/managed-by"
)

func inventoryRef(key client.ObjectKey) ObjectRef {
	return ObjectRef{
		GroupKind: schema.GroupKind{Kind: "ConfigMap"},
		Namespace: key.Namespace,
		Name:      key.Name,
	}
}

// prepareInventory reads the kinds recorded by the previous Apply and, before anything is applied, records the
// union of those and the kinds of objs. If this Apply fails midway, the next one still prunes every kind that
// may have objects left behind.
func (a *Applier) prepareInventory(ctx context.Context, objs []*unstructured.Unstructured, opts ApplyOptions) (sets.Set[schema.GroupKind], error) {
	if opts.Inventory == nil {
		return sets.New[schema.GroupKind](), nil
	}

	previous, err := a.readInventory(ctx, *opts.Inventory)
	if err != nil {
		return nil, err
	}

	if opts.DryRun {
		return previous, nil
	}

	return previous, a.writeInventory(ctx, *opts.Inventory, previous.Union(groupKinds(objs)))
}

// finalizeInventory records only the kinds of objs, once objects of other kinds have been pruned.
func (a *Applier) finalizeInventory(ctx context.Context, objs []*unstructured.Unstructured, opts ApplyOptions) error {
	if opts.Inventory == nil || opts.DryRun || opts.PruneSelector == nil {
		return nil
	}

	return a.writeInventory(ctx, *opts.Inventory, groupKinds(objs))
}

func (a *Applier) readInventory(ctx context.Context, key client.ObjectKey) (sets.Set[schema.GroupKind], error) {
	kinds := sets.New[schema.GroupKind]()

	var cm corev1.ConfigMap
	if err := a.client.Get(ctx, key, &cm); err != nil {
		return kinds, fail.KubeClient(client.IgnoreNotFound(err), "reading inventory %s", key)
	}

	for line := range strings.Lines(cm.Data[InventoryGroupKindsKey]) {
		if line = strings.TrimSpace(line); line != "" {
			kinds.Insert(schema.ParseGroupKind(line))
		}
	}

	return kinds, nil
}

func (a *Applier) writeInventory(ctx context.Context, key client.ObjectKey, kinds sets.Set[schema.GroupKind]) error {
	lines := make([]string, 0, kinds.Len())
	for gk := range kinds {
		lines = append(lines, gk.String())
	}

	cm := corev1ac.ConfigMap(key.Name, key.Namespace).
		WithLabels(map[string]string{inventoryManagedByLabel: a.fieldManager}).
		WithData(map[string]string{InventoryGroupKindsKey: strings.Join(sets.List(sets.New(lines...)), "\n")})

	err := a.client.Apply(ctx, cm, client.FieldOwner(a.fieldManager), client.ForceOwnership)

	return fail.KubeClient(err, "writing inventory %s", key)
}

func (a *Applier) deleteInventory(ctx context.Context, key client.ObjectKey) error {
	cm := &corev1.ConfigMap{}
	cm.SetName(key.Name)
	cm.SetNamespace(key.Namespace)

	return fail.KubeClient(client.IgnoreNotFound(a.client.Delete(ctx, cm)), "deleting inventory %s", key)
}

func groupKinds(objs []*unstructured.Unstructured) sets.Set[schema.GroupKind] {
	kinds := sets.New[schema.GroupKind]()
	for _, obj := range objs {
		kinds.Insert(obj.GroupVersionKind().GroupKind())
	}

	return kinds
}
