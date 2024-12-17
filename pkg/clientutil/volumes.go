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
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/reconciler/pkg/reconciling"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	annotationKeyDescription = "description"

	// AnnDynamicallyProvisioned is added to a PV that is dynamically provisioned by kubernetes
	// Because the annotation is defined only at k8s.io/kubernetes, copying the content instead of vendoring
	// https://github.com/kubernetes/kubernetes/blob/v1.21.0/pkg/controller/volume/persistentvolume/util/util.go#L65
	AnnDynamicallyProvisioned = "pv.kubernetes.io/provisioned-by"
)

var VolumeResources = []string{"persistentvolumes", "persistentvolumeclaims"}

func CleanupUnretainedVolumes(ctx context.Context, logger logrus.FieldLogger, c client.Client) error {
	// We disable the PV & PVC creation so nothing creates new PV's while we delete them
	logger.Infoln("Creating ValidatingWebhookConfiguration to disable future PV & PVC creation...")
	if err := disablePVCreation(ctx, c); err != nil {
		return fail.KubeClient(err, "failed to disable future PV & PVC creation.")
	}

	pvcList, pvList, err := getDynamicallyProvisionedUnretainedPvs(ctx, c)
	if err != nil {
		return err
	}

	// Do not attempt to delete any pods when there are no PVs and PVCs
	if (pvcList != nil && pvList != nil) && len(pvcList.Items) == 0 && len(pvList.Items) == 0 {
		return nil
	}

	// Delete all Pods that use PVs. We must keep the remaining pods, otherwise
	// we end up in a deadlock when CSI is used
	if err := cleanupPVCUsingPods(ctx, c); err != nil {
		return fail.KubeClient(err, "failed to clean up PV using pod from user cluster.")
	}

	// Delete PVC's
	logger.Infoln("Deleting persistent volume claims...")
	for _, pvc := range pvcList.Items {
		if pvc.DeletionTimestamp == nil {
			identifier := fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)
			logger.Infoln("Deleting PVC...", identifier)

			if err := DeleteIfExists(ctx, c, &pvc); err != nil {
				return fail.KubeClient(err, "failed to delete PVC from user cluster.")
			}
		}
	}

	return nil
}

func disablePVCreation(ctx context.Context, c client.Client) error {
	// Prevent re-creation of PVs and PVCs by using an intentionally defunct admissionWebhook
	creatorGetters := []reconciling.NamedValidatingWebhookConfigurationReconcilerFactory{
		creationPreventingWebhook("", VolumeResources),
	}
	if err := reconciling.ReconcileValidatingWebhookConfigurations(ctx, creatorGetters, "", c); err != nil {
		return fail.KubeClient(err, "failed to create ValidatingWebhookConfiguration to prevent creation of PVs/PVCs.")
	}

	return nil
}

func cleanupPVCUsingPods(ctx context.Context, c client.Client) error {
	podList := &corev1.PodList{}
	if err := c.List(ctx, podList); err != nil {
		return fail.KubeClient(err, "failed to list Pods from user cluster.")
	}

	var pvUsingPods []*corev1.Pod
	for idx := range podList.Items {
		pod := &podList.Items[idx]
		if podUsesPV(pod) {
			pvUsingPods = append(pvUsingPods, pod)
		}
	}

	for _, pod := range pvUsingPods {
		if pod.DeletionTimestamp == nil {
			if err := DeleteIfExists(ctx, c, pod); err != nil {
				return fail.KubeClient(err, "failed to delete Pod.")
			}
		}
	}

	return nil
}

func podUsesPV(p *corev1.Pod) bool {
	for _, volume := range p.Spec.Volumes {
		if volume.VolumeSource.PersistentVolumeClaim != nil {
			return true
		}
	}

	return false
}

func getDynamicallyProvisionedUnretainedPvs(ctx context.Context, c client.Client) (*corev1.PersistentVolumeClaimList, *corev1.PersistentVolumeList, error) {
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := c.List(ctx, pvcList); err != nil {
		return nil, nil, fail.KubeClient(err, "failed to list PVCs from user cluster.")
	}
	allPVList := &corev1.PersistentVolumeList{}
	if err := c.List(ctx, allPVList); err != nil {
		return nil, nil, fail.KubeClient(err, "failed to list PVs from user cluster.")
	}
	pvList := &corev1.PersistentVolumeList{}
	for _, pv := range allPVList.Items {
		// Check only dynamically provisioned PVs with delete reclaim policy to verify provisioner has done the cleanup
		// this filters out everything else because we leave those be
		if pv.Annotations[AnnDynamicallyProvisioned] != "" && pv.Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimDelete {
			pvList.Items = append(pvList.Items, pv)
		}
	}

	return pvcList, pvList, nil
}

func WaitCleanUpVolumes(ctx context.Context, logger logrus.FieldLogger, c client.Client) error {
	logger.Infoln("Waiting for all dynamically provisioned and unretained volumes to get deleted...")

	return wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, false, func(ctx context.Context) (bool, error) {
		pvcList, pvList, err := getDynamicallyProvisionedUnretainedPvs(ctx, c)
		if err != nil {
			return false, nil
		}

		if (pvcList != nil && pvList != nil) && len(pvcList.Items) == 0 && len(pvList.Items) == 0 {
			return true, nil
		}

		return false, nil
	})
}
