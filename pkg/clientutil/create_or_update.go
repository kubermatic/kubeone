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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateOrUpdate makes it easy to "apply" objects to kubernetes API server
func CreateOrUpdate(ctx context.Context, c client.Client, obj runtime.Object) error {
	existing := obj.DeepCopyObject()
	existingMetaObj, ok := existing.(metav1.Object)
	if !ok {
		return errors.Errorf("%T does not implement metav1.Object interface", obj)
	}

	key := client.ObjectKey{
		Name:      existingMetaObj.GetName(),
		Namespace: existingMetaObj.GetNamespace(),
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
