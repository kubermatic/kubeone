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
	"github.com/pkg/errors"

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

	existing := obj.DeepCopyObject().(client.Object)
	key := client.ObjectKey{
		Name:      existing.GetName(),
		Namespace: existing.GetNamespace(),
	}

	err := c.Get(ctx, key, existing)

	switch {
	case k8serrors.IsNotFound(err):
		return errors.Wrapf(c.Create(ctx, obj), "failed to create %T", obj)
	case err != nil:
		return errors.Wrapf(err, "failed to get %T object", obj)
	}

	if err = mergo.Merge(obj, existing); err != nil {
		return errors.Wrap(err, "failed to merge objects")
	}

	return errors.Wrapf(c.Update(ctx, obj), "failed to update %T object", obj)
}
