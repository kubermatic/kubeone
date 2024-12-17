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
	"time"

	"github.com/sirupsen/logrus"

	"k8c.io/kubeone/pkg/fail"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CleanupLBs(ctx context.Context, logger logrus.FieldLogger, c client.Client) error {
	serviceList := &corev1.ServiceList{}
	if err := c.List(ctx, serviceList); err != nil {
		return fail.KubeClient(err, "failed to list Service.")
	}

	for _, service := range serviceList.Items {
		// This service is already in deletion, nothing further needs to happen.
		if service.DeletionTimestamp != nil {
			continue
		}
		// Only LoadBalancer services incur charges on cloud providers
		if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
			logger.Infof("Deleting SVC : %s/%s\n", service.Namespace, service.Name)
			if err := DeleteIfExists(ctx, c, &service); err != nil {
				return err
			}
		}
	}

	return nil
}

func WaitCleanupLbs(ctx context.Context, logger logrus.FieldLogger, c client.Client) error {
	logger.Infoln("Waiting for all load balancer services to get deleted...")

	return wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, false, func(ctx context.Context) (bool, error) {
		serviceList := &corev1.ServiceList{}
		if err := c.List(ctx, serviceList); err != nil {
			return false, nil
		}
		for _, service := range serviceList.Items {
			// Only LoadBalancer services incur charges on cloud providers
			if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
				return false, nil
			}
		}

		return true, nil
	})
}
