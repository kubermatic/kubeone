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

	"k8c.io/kubeone/pkg/fail"

	corev1 "k8s.io/api/core/v1"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// PodsReadyCondition generate a k8s.io/apimachinery/pkg/util/wait.ConditionFunc function to be used in
// k8s.io/apimachinery/pkg/util/wait.Poll* family of functions. It will check all selected pods to have PodReady
// condition and phase PodRunning.
func PodsReadyCondition(ctx context.Context, c dynclient.Client, listOpts dynclient.ListOptions) func() (bool, error) {
	return func() (bool, error) {
		podsList := corev1.PodList{}

		if err := c.List(ctx, &podsList, &listOpts); err != nil {
			return false, fail.KubeClient(err, "listing pods")
		}

		return allPodsAreRunningAndReady(&podsList), nil
	}
}

func allPodsAreRunningAndReady(podsList *corev1.PodList) bool {
	if len(podsList.Items) == 0 {
		return false
	}

	var readyNum int

	for _, pod := range podsList.Items {
		if pod.Status.Phase != corev1.PodRunning {
			return false
		}

		for _, podcond := range pod.Status.Conditions {
			if podcond.Type == corev1.PodReady && podcond.Status == corev1.ConditionTrue {
				readyNum++
			}
		}
	}

	return readyNum == len(podsList.Items)
}
