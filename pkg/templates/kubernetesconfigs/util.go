/*
Copyright 2022 The KubeOne Authors.

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

package kubernetesconfigs

import (
	"encoding/json"

	metav1unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func dropFields(obj runtime.Object, fields ...[]string) (runtime.Object, error) {
	buf, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var uObj metav1unstructured.Unstructured

	_, _, err = metav1unstructured.UnstructuredJSONScheme.Decode(buf, nil, &uObj)
	if err != nil {
		return nil, err
	}

	for _, fieldSet := range fields {
		metav1unstructured.RemoveNestedField(uObj.Object, fieldSet...)
	}

	return &uObj, nil
}
