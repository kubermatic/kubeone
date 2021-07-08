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

	"github.com/pkg/errors"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeleteIfExists makes it easy to "delete" Kubernetes objects if they exist
func DeleteIfExists(ctx context.Context, c client.Client, obj client.Object) error {
	err := c.Delete(ctx, obj)

	switch {
	case k8serrors.IsNotFound(err):
		return nil
	case err != nil:
		return errors.Wrapf(err, "failed to delete %T object", obj)
	default:
		return nil
	}
}
