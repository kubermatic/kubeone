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

	"github.com/pkg/errors"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrCRDNotEstablished = errors.New("crd is not established")
)

func VerifyCRD(ctx context.Context, client dynclient.Client, crdName string) (bool, error) {
	crd := &apiextensions.CustomResourceDefinition{}
	key := dynclient.ObjectKey{Name: crdName}
	if err := client.Get(ctx, key, crd); err != nil {
		return false, err
	}

	for _, cond := range crd.Status.Conditions {
		if cond.Type == apiextensions.Established && cond.Status == apiextensions.ConditionTrue {
			return true, nil
		}
	}

	return false, ErrCRDNotEstablished
}
