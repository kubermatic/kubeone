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

	"k8c.io/kubeone/pkg/fail"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/csaupgrade"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// clientSideApplyManagers are the field managers whose ownership is handed over to the server-side apply field
// manager: "kubectl apply" without --server-side, and the placeholder the API server assigns to objects created
// before managed fields existed.
var clientSideApplyManagers = sets.New("kubectl-client-side-apply", "before-first-apply")

// migrateClientSideApply transfers ownership of fields managed by client-side apply to the Applier's field
// manager, so that fields removed from the manifest get removed from the live object by the next apply.
// It reports whether a migration took place.
func (a *Applier) migrateClientSideApply(ctx context.Context, live *unstructured.Unstructured) (bool, error) {
	var migrated bool

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		patch, err := csaupgrade.UpgradeManagedFieldsPatch(live, clientSideApplyManagers, a.fieldManager)
		if err != nil {
			return err
		}

		if patch == nil {
			return nil
		}

		err = a.client.Patch(ctx, live, client.RawPatch(types.JSONPatchType, patch))
		if err == nil {
			migrated = true

			return nil
		}

		// The patch pins resourceVersion; refresh the object before computing a new patch.
		if getErr := a.client.Get(ctx, client.ObjectKeyFromObject(live), live); getErr != nil {
			return getErr
		}

		return err
	})

	return migrated, fail.KubeClient(err, "migrating managed fields of %s from client-side apply", RefOf(live))
}
