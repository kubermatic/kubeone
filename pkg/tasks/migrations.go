/*
Copyright 2024 The KubeOne Authors.

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

package tasks

import (
	"fmt"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
)

func migratePVCAllocatedResourceStatus(s *state.State) error {
	s.Logger.Info("Removing AllocatedResourceStatus from PersistentVolumeClaims...")

	pvcList := corev1.PersistentVolumeClaimList{}

	if err := s.DynamicClient.List(s.Context, &pvcList); err != nil {
		return fail.KubeClient(err, "getting %T", pvcList)
	}

	for _, pvc := range pvcList.Items {
		log := s.Logger.WithField("pvc", fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name))
		log.Debug("Checking AllocatedResourceStatus for PVC...")

		var found bool
		for k, v := range pvc.Status.AllocatedResourceStatuses {
			if k == corev1.ResourceStorage &&
				(v == "ControllerResizeFailed" || v == "NodeResizeFailed") {
				log.Info("Removing AllocatedResourceStatus from PVC...")
				found = true

				pvc.Status.AllocatedResourceStatuses = map[corev1.ResourceName]corev1.ClaimResourceStatus{}

				if err := s.DynamicClient.Status().Update(s.Context, &pvc); err != nil {
					return fail.KubeClient(err, "updating pvc")
				}

				break
			}
		}
		if !found {
			log.Debug("No AllocatedResourceStatus found for PVC.")
		}
	}

	return nil
}
