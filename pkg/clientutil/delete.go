/*
Copyright 2021 The KubeOne Authors.

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
	"time"

	"k8c.io/kubeone/pkg/fail"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const deleteWaitInterval = 2 * time.Second

// DeleteIfExists makes it easy to "delete" Kubernetes objects if they exist
func DeleteIfExists(ctx context.Context, c client.Client, obj client.Object) error {
	err := c.Delete(ctx, obj)

	if !k8serrors.IsNotFound(err) {
		return fail.KubeClient(err, "deleting %T %s", obj, client.ObjectKeyFromObject(obj))
	}

	return nil
}

// DeleteAndWait deletes the object, if it exists, with foreground propagation
// and waits until it's gone. With foreground propagation the object is removed
// only after all its dependents, e.g. the pods of a DaemonSet, are deleted.
func DeleteAndWait(ctx context.Context, c client.Client, obj client.Object, timeout time.Duration) error {
	key := client.ObjectKeyFromObject(obj)

	err := c.Delete(ctx, obj, client.PropagationPolicy(metav1.DeletePropagationForeground))
	if k8serrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fail.KubeClient(err, "deleting %T %s", obj, key)
	}

	err = wait.PollUntilContextTimeout(ctx, deleteWaitInterval, timeout, true, func(ctx context.Context) (bool, error) {
		getErr := c.Get(ctx, key, obj)
		if k8serrors.IsNotFound(getErr) {
			return true, nil
		}

		return false, getErr
	})

	return fail.KubeClient(err, "waiting for %T %s to be deleted", obj, key)
}
