/*
Copyright 2020 The KubeOne Authors.

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

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// CRDsReadyCondition generate a k8s.io/apimachinery/pkg/util/wait.ConditionFunc function to be used in
// k8s.io/apimachinery/pkg/util/wait.Poll* family of functions. It will check all provided GKs (GroupKinds) to exists
// and have Established status
func CRDsReadyCondition(ctx context.Context, client dynclient.Client, names []string) func() (bool, error) {
	return func() (bool, error) {
		var establishedNum int

		for _, gk := range names {
			crd := apiextensions.CustomResourceDefinition{}
			key := dynclient.ObjectKey{Name: gk}

			if err := client.Get(ctx, key, &crd); err != nil {
				return false, err
			}

			for _, cond := range crd.Status.Conditions {
				if cond.Type == apiextensions.Established && cond.Status == apiextensions.ConditionTrue {
					establishedNum++
				}
			}
		}

		return establishedNum == len(names), nil
	}
}
