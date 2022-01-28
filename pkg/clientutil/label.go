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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	KubeoneComponentLabel = "kubeone.io/component"
)

func AddLabels(labels map[string]string, objects ...runtime.Object) []runtime.Object {
	for i := range objects {
		metaobj, _ := objects[i].(metav1.Object)
		existingLabels := metaobj.GetLabels()
		copyLabels := map[string]string{}

		for k, v := range existingLabels {
			copyLabels[k] = v
		}

		for k, v := range labels {
			copyLabels[k] = v
		}

		metaobj.SetLabels(copyLabels)
		objects[i], _ = metaobj.(runtime.Object)
	}

	return objects
}

func LabelComponent(component string, objects ...runtime.Object) []runtime.Object {
	return AddLabels(map[string]string{KubeoneComponentLabel: component}, objects...)
}
