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

package externalccm

import (
	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	vSphereDeploymentName = "vsphere-cloud-controller-manager"
)

func migrateVsphereAddon(s *state.State) error {
	return clientutil.DeleteIfExists(s.Context, s.DynamicClient, vSphereService())
}

func vSphereService() *corev1.Service {
	// We're intentionally keeping only Service metadata, as it's enough for
	// deleting the object
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vSphereDeploymentName,
			Namespace: metav1.NamespaceSystem,
		},
	}
}
