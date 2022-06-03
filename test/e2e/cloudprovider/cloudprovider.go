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

package cloudprovider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"k8c.io/kubeone/test/e2e/provisioner"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func supportsStorage(cloudProvider string) bool {
	switch cloudProvider {
	case provisioner.AWS, provisioner.DigitalOcean, provisioner.Hetzner, provisioner.GCE, provisioner.OpenStack:
		return true
	default:
		return false
	}
}

func supportsLoadBalancer(cloudProvider string) bool {
	switch cloudProvider {
	case provisioner.AWS, provisioner.DigitalOcean, provisioner.Hetzner, provisioner.GCE, provisioner.OpenStack:
		return true
	default:
		return false
	}
}

const (
	namespaceName   = "test-cloud-provider"
	statefulsetName = "test-sts"
	serviceName     = "test-svc"

	testPollPeriod = 10 * time.Second
	testTimeout    = 10 * time.Minute
)

var (
	labels = map[string]string{"app": "test-cp"}
)

func RunCloudProviderTests(ctx context.Context, t *testing.T, client ctrlruntimeclient.Client, cloudProvider string) error {
	t.Helper()
	t.Log("Testing cloud provider support...")

	if err := createPVCWithStorage(ctx, t, client, cloudProvider); err != nil {
		return err
	}
	if err := exposePVC(ctx, t, client, cloudProvider); err != nil {
		return err
	}

	return cleanUp(ctx, t, client)
}

func createPVCWithStorage(ctx context.Context, t *testing.T, client ctrlruntimeclient.Client, cloudProvider string) error {
	t.Helper()

	if !supportsStorage(cloudProvider) {
		t.Logf("Skipping storage tests because cloud provider %q is not supported.", cloudProvider)

		return nil
	}

	t.Log("Testing storage support...")

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	if err := client.Create(ctx, ns); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	t.Log("Creating a StatefulSet with PVC...")

	// Creating a simple StatefulSet with 1 replica which writes to the PV. That way we know if storage can be provisioned and consumed
	set := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      statefulsetName,
			Namespace: ns.Name,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
							Ports: []corev1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: 80,
								},
							},
						},
						{
							Name:  "busybox",
							Image: "k8s.gcr.io/busybox",
							Args: []string{
								"/bin/sh",
								"-c",
								`echo "alive" > /data/healthy; sleep 3600`,
							},
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: 0,
								SuccessThreshold:    3,
								PeriodSeconds:       5,
								TimeoutSeconds:      1,
								FailureThreshold:    1,
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"cat",
											"/data/healthy",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Gi")},
						},
					},
				},
			},
		},
	}

	if err := client.Create(ctx, set); err != nil {
		return fmt.Errorf("failed to create statefulset: %w", err)
	}

	t.Log("Waiting until the StatefulSet is ready...")

	err := wait.Poll(testPollPeriod, testTimeout, func() (done bool, err error) {
		currentSet := &appsv1.StatefulSet{}
		name := types.NamespacedName{Namespace: ns.Name, Name: set.Name}

		if err := client.Get(ctx, name, currentSet); err != nil {
			t.Logf("Failed to fetch StatefulSet %s/%s: %v", ns.Name, set.Name, err)

			return false, nil
		}

		if currentSet.Status.ReadyReplicas == 1 {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return fmt.Errorf("failed to check if StatefulSet is running: %w", err)
	}

	t.Log("Successfully validated storage support")

	return nil
}

func exposePVC(ctx context.Context, t *testing.T, client ctrlruntimeclient.Client, cloudProvider string) error {
	t.Helper()

	if !supportsLoadBalancer(cloudProvider) {
		t.Logf("Skipping load balancer tests because cloud provider %q is not supported.", cloudProvider)

		return nil
	}

	t.Log("Testing Load Balancer support...")

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespaceName,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "web",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
				},
			},
		},
	}
	if cloudProvider == provisioner.Hetzner {
		svc.Annotations = map[string]string{
			"load-balancer.hetzner.cloud/location": "nbg1",
		}
	}

	if err := client.Create(ctx, svc); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	t.Log("Waiting for Load Balancer got the external IP address...")

	err := wait.Poll(testPollPeriod, testTimeout, func() (done bool, err error) {
		currentSvc := &corev1.Service{}
		name := types.NamespacedName{Namespace: namespaceName, Name: svc.Name}

		if err := client.Get(ctx, name, currentSvc); err != nil {
			t.Logf("Failed to fetch Service %s/%s: %v", namespaceName, svc.Name, err)

			return false, nil
		}

		if len(currentSvc.Status.LoadBalancer.Ingress) > 0 {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return fmt.Errorf("failed to check if StatefulSet is exposed via Service: %w", err)
	}

	t.Log("Successfully validated the Load Balancer support")

	return nil
}

func cleanUp(ctx context.Context, t *testing.T, client ctrlruntimeclient.Client) error {
	t.Helper()

	t.Log("Cleaning up Load Balancer...")

	err := wait.Poll(testPollPeriod, testTimeout, func() (done bool, err error) {
		currentSvc := &corev1.Service{}
		name := types.NamespacedName{Namespace: namespaceName, Name: serviceName}

		if err := client.Get(ctx, name, currentSvc); err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}

			t.Logf("Failed to fetch Service %s/%s: %v", namespaceName, serviceName, err)

			return false, nil
		}

		if currentSvc.ObjectMeta.DeletionTimestamp == nil {
			if err := client.Delete(ctx, currentSvc); err != nil {
				t.Logf("Failed to delete Service: %v", err)

				return false, nil
			}
		}

		return false, nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete Service: %w", err)
	}

	t.Log("Cleaning up StatefulSet and PVC...")

	err = wait.Poll(testPollPeriod, testTimeout, func() (done bool, err error) {
		currentSts := &appsv1.StatefulSet{}
		name := types.NamespacedName{Namespace: namespaceName, Name: statefulsetName}

		if err := client.Get(ctx, name, currentSts); err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}

			t.Logf("Failed to fetch Statefulset: %v", err)

			return false, nil
		}

		if currentSts.ObjectMeta.DeletionTimestamp == nil {
			if err := client.Delete(ctx, currentSts); err != nil {
				t.Logf("Failed to delete Statefulset: %v", err)

				return false, nil
			}
		}

		return false, nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete Statefulset: %w", err)
	}

	return nil
}
