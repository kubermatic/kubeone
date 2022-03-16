/*
Copyright 2019 The KubeOne Authors.

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

package clientutil

import (
	"context"

	"github.com/imdario/mergo"

	"k8c.io/kubeone/pkg/fail"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Updater func(client.Client, client.Object)

func WithComponentLabel(componentname string) Updater {
	return func(c client.Client, obj client.Object) {
		LabelComponent(componentname, obj)
	}
}

// CreateOrUpdate makes it easy to "apply" objects to kubernetes API server
func CreateOrUpdate(ctx context.Context, c client.Client, obj client.Object, updaters ...Updater) error {
	for _, update := range updaters {
		update(c, obj)
	}

	existing, _ := obj.DeepCopyObject().(client.Object)
	key := client.ObjectKey{
		Name:      existing.GetName(),
		Namespace: existing.GetNamespace(),
	}

	err := c.Get(ctx, key, existing)

	switch {
	case k8serrors.IsNotFound(err):
		err = c.Create(ctx, obj)

		return fail.KubeClient(err, "creating %T %s", obj, key)
	case err != nil:
		return fail.KubeClient(err, "getting %T %s", obj, key)
	}

	if err = mergo.Merge(obj, existing); err != nil {
		return fail.Runtime(err, "merging updated %T %s with existing", obj, key)
	}

	return fail.KubeClient(c.Update(ctx, obj), "updating %T %s", obj, key)
}

// CreateOrReplace makes it easy to "replace" objects
func CreateOrReplace(ctx context.Context, c client.Client, obj client.Object, updaters ...Updater) error {
	for _, update := range updaters {
		update(c, obj)
	}

	err := c.Create(ctx, obj)
	if err == nil {
		return nil // success!
	}

	key := client.ObjectKeyFromObject(obj)

	// Object does not exist already, but creating failed for another reason
	if !k8serrors.IsAlreadyExists(err) {
		return fail.KubeClient(err, "creating %T %s", obj, key)
	}

	// Object exists already, time to update it
	existingObj, _ := obj.DeepCopyObject().(client.Object)

	if err = c.Get(ctx, key, existingObj); err != nil {
		return fail.KubeClient(err, "getting %T %s", obj, key)
	}

	// do not use mergo to merge the existing into the new object,
	// because this would bring back the "kubectl apply" semantics;
	// we want "kubectl replace" semantics instead, so we only keep
	// a few fields from the metadata intact and overwrite everything else

	obj.SetResourceVersion(existingObj.GetResourceVersion())
	obj.SetGeneration(existingObj.GetGeneration())

	return fail.Runtime(c.Update(ctx, obj), "updating %T %s", obj, key)
}
