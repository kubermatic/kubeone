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

package e2e

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func supportsStorageTests(provider string) bool {
	switch provider {
	case "aws", "azure", "digitalocean", "gce", "hetzner", "nutanix", "openstack", "vsphere":
		return true
	default:
		return false
	}
}

func supportsLoadBalancerTests(provider string) bool {
	switch provider {
	// our openstack provider where we run e2e does not support loadbalancers
	case
		"aws",
		"azure",
		"digitalocean",
		"gce",
		"hetzner",
		"openstack":
		return true
	default:
		return false
	}
}

func supportsSnapshotterTests(provider string) bool {
	switch provider {
	case "aws", "azure", "digitalocean", "gce", "nutanix", "openstack", "vsphere":
		return true
	default:
		return false
	}
}

const (
	cpTestNamespaceName   = "test-cloud-provider"
	cpTestStatefulSetName = "test-sts"
	cpTestServiceName     = "test-svc"
	cpTestSnapshotName    = "test-snapshot"

	cpTestPollPeriod = 10 * time.Second
	cpTestTimeout    = 10 * time.Minute
)

var (
	cpTestVolumeName = "data-test-cp-0"

	cloudProviderPodLabels = map[string]string{"app": "test-cp"}
)

type cloudProviderTests struct {
	ctx      context.Context
	client   ctrlruntimeclient.Client
	provider string
}

func newCloudProviderTests(client ctrlruntimeclient.Client, provider string) *cloudProviderTests {
	return &cloudProviderTests{
		ctx:      context.Background(),
		client:   client,
		provider: provider,
	}
}

func (c *cloudProviderTests) run(t *testing.T) {
	c.createStatefulSetWithStorage(t)
	c.createVolumeSnapshot(t)
	c.exposeStatefulSet(t)
}

func (c *cloudProviderTests) runWithCleanup(t *testing.T) {
	defer c.cleanUp(t)

	c.run(t)
}

func (c *cloudProviderTests) createStatefulSetWithStorage(t *testing.T) {
	if !supportsStorageTests(c.provider) {
		t.Logf("Skipping cloud provider storage tests because cloud provider %q is not supported.", c.provider)

		return
	}

	t.Log("Testing storage support...")

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: cpTestNamespaceName,
		},
	}
	err := retryFn(func() error {
		return c.client.Create(c.ctx, ns)
	})
	if err != nil {
		t.Fatalf("creating test namespace: %v", err)
	}

	t.Log("Creating a StatefulSet with PVC...")

	// Creating a simple StatefulSet with 1 replica which writes to the PV. That way we know if storage can be provisioned and consumed
	set := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cpTestStatefulSetName,
			Namespace: ns.Name,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: cloudProviderPodLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: cloudProviderPodLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "echoserver",
							Image: "registry.k8s.io/echoserver:1.10",
							Ports: []corev1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: 8080,
								},
							},
						},
						{
							Name:  "busybox",
							Image: "registry.k8s.io/busybox",
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
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Gi")},
						},
					},
				},
			},
		},
	}

	err = retryFn(func() error {
		return c.client.Create(c.ctx, set)
	})
	if err != nil {
		t.Fatalf("creating test statefulset: %v", err)
	}

	c.validateStatefulSetReadiness(t)
}

func (c *cloudProviderTests) validateStatefulSetReadiness(t *testing.T) {
	t.Log("Waiting until the StatefulSet is ready...")

	err := wait.PollUntilContextTimeout(c.ctx, cpTestPollPeriod, cpTestTimeout, false, func(ctx context.Context) (done bool, err error) {
		currentSet := &appsv1.StatefulSet{}
		name := types.NamespacedName{Namespace: cpTestNamespaceName, Name: cpTestStatefulSetName}

		if err := c.client.Get(ctx, name, currentSet); err != nil {
			t.Logf("Failed to fetch StatefulSet %s/%s: %v", cpTestNamespaceName, cpTestStatefulSetName, err)

			return false, nil
		}

		if currentSet.Status.ReadyReplicas == 1 {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		t.Fatalf("waiting for statefulset: %v", err)
	}

	t.Log("Successfully validated storage support")
}

func (c *cloudProviderTests) exposeStatefulSet(t *testing.T) {
	if !supportsLoadBalancerTests(c.provider) {
		t.Logf("Skipping cloud provider load balancer tests because cloud provider %q is not supported.", c.provider)

		return
	}

	t.Log("Testing Load Balancer support...")

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cpTestServiceName,
			Namespace: cpTestNamespaceName,
			Annotations: map[string]string{
				"load-balancer.hetzner.cloud/location": "nbg1",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Selector: cloudProviderPodLabels,
			Ports: []corev1.ServicePort{
				{
					Name:       "web",
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(8080),
					Port:       80,
				},
			},
		},
	}

	err := retryFn(func() error {
		return c.client.Create(c.ctx, svc)
	})
	if err != nil {
		t.Fatalf("creating test service: %v", err)
	}

	c.validateLoadBalancerReadiness(t)
}

func (c *cloudProviderTests) validateLoadBalancerReadiness(t *testing.T) {
	t.Log("Waiting for Load Balancer got the external IP address...")

	var svcAddr string

	err := wait.PollUntilContextTimeout(c.ctx, cpTestPollPeriod, cpTestTimeout, false, func(ctx context.Context) (done bool, err error) {
		currentSvc := &corev1.Service{}
		name := types.NamespacedName{Namespace: cpTestNamespaceName, Name: cpTestServiceName}

		if err := c.client.Get(ctx, name, currentSvc); err != nil {
			t.Logf("Failed to fetch Service %s/%s: %v", cpTestNamespaceName, cpTestServiceName, err)

			return false, nil
		}

		if len(currentSvc.Status.LoadBalancer.Ingress) > 0 {
			if currentSvc.Status.LoadBalancer.Ingress[0].Hostname != "" {
				svcAddr = currentSvc.Status.LoadBalancer.Ingress[0].Hostname
			} else if currentSvc.Status.LoadBalancer.Ingress[0].IP != "" {
				svcAddr = currentSvc.Status.LoadBalancer.Ingress[0].IP
			}

			return true, nil
		}

		return false, nil
	})
	if err != nil {
		t.Fatalf("waiting for statefulset to become exposed: %v", err)
	}

	t.Log("Waiting for Load Balancer to become reachable...")

	if !strings.HasPrefix(svcAddr, "http://") {
		svcAddr = "http://" + svcAddr
	}

	err = wait.PollUntilContextTimeout(c.ctx, cpTestPollPeriod, cpTestTimeout, false, func(_ context.Context) (done bool, err error) {
		resp, err := http.Get(svcAddr) //nolint:gosec,noctx
		if err != nil {
			t.Logf("error testing service endpoint: %v", err)

			return false, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		t.Fatalf("waiting for service to become reachable: %v", err)
	}

	t.Log("Successfully validated the Load Balancer support")
}

func (c *cloudProviderTests) createVolumeSnapshot(t *testing.T) {
	if !supportsSnapshotterTests(c.provider) {
		t.Logf("Skipping snapshotter tests because cloud provider %q is not supported.", c.provider)

		return
	}

	t.Log("Testing CSI snapshotter support...")

	snapshot := &snapshotv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cpTestSnapshotName,
			Namespace: cpTestNamespaceName,
		},
		Spec: snapshotv1.VolumeSnapshotSpec{
			Source: snapshotv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &cpTestVolumeName,
			},
		},
	}

	err := retryFn(func() error {
		return c.client.Create(c.ctx, snapshot)
	})
	if err != nil {
		t.Fatalf("creating test volumesnapshot: %v", err)
	}

	c.validateVolumeSnapshot(t)
}

func (c *cloudProviderTests) validateVolumeSnapshot(t *testing.T) {
	t.Log("Waiting for VolumeSnapshot to become ready to use...")

	err := wait.PollUntilContextTimeout(c.ctx, cpTestPollPeriod, cpTestTimeout, false, func(ctx context.Context) (done bool, err error) {
		currentSnap := &snapshotv1.VolumeSnapshot{}
		name := types.NamespacedName{Namespace: cpTestNamespaceName, Name: cpTestSnapshotName}

		if err := c.client.Get(ctx, name, currentSnap); err != nil {
			t.Logf("Failed to fetch VolumeSnapshot %s/%s: %v", cpTestNamespaceName, cpTestSnapshotName, err)

			return false, nil
		}

		if currentSnap.Status != nil && currentSnap.Status.ReadyToUse != nil {
			return *currentSnap.Status.ReadyToUse, nil
		}

		return false, nil
	})
	if err != nil {
		t.Fatalf("waiting for statefulset to become exposed: %v", err)
	}

	t.Log("Successfully validated the CSI snapshotter support")
}

func (c *cloudProviderTests) cleanUp(t *testing.T) {
	t.Log("Cleaning up Load Balancer...")

	err := wait.PollUntilContextTimeout(c.ctx, cpTestPollPeriod, cpTestTimeout, false, func(ctx context.Context) (done bool, err error) {
		currentSvc := &corev1.Service{}
		name := types.NamespacedName{Namespace: cpTestNamespaceName, Name: cpTestServiceName}

		if err := c.client.Get(ctx, name, currentSvc); err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}

			// Make error transient so that we try to remove it again and
			// not leak any resources
			t.Logf("error getting load balancer service: %v", err)

			return false, nil
		}

		if currentSvc.ObjectMeta.DeletionTimestamp == nil {
			if err := c.client.Delete(ctx, currentSvc); err != nil {
				// Make error transient so that we try to remove it again and
				// not leak any resources
				t.Logf("error removing load balancer service: %v", err)

				return false, nil
			}
		}

		return false, nil
	})
	if err != nil {
		t.Fatalf("error waiting for load balancer service to get removed: %v", err)
	}

	t.Log("Cleaning up StatefulSet...")

	err = wait.PollUntilContextTimeout(c.ctx, cpTestPollPeriod, cpTestTimeout, false, func(ctx context.Context) (done bool, err error) {
		currentSts := &appsv1.StatefulSet{}
		name := types.NamespacedName{Namespace: cpTestNamespaceName, Name: cpTestStatefulSetName}

		if err := c.client.Get(ctx, name, currentSts); err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}

			// Make error transient so that we try to remove it again and
			// not leak any resources
			t.Logf("error getting statefulset: %v", err)

			return false, nil
		}

		if currentSts.ObjectMeta.DeletionTimestamp == nil {
			if err := c.client.Delete(ctx, currentSts); err != nil {
				// Make error transient so that we try to remove it again and
				// not leak any resources
				t.Logf("error removing statefulset: %v", err)

				return false, nil
			}
		}

		return false, nil
	})
	if err != nil {
		t.Fatalf("error waiting for statefulset to get removed: %v", err)
	}

	t.Log("Cleaning up VolumeSnapshot...")

	err = wait.PollUntilContextTimeout(c.ctx, cpTestPollPeriod, cpTestTimeout, false, func(ctx context.Context) (done bool, err error) {
		var snaps snapshotv1.VolumeSnapshotList
		if err := c.client.List(ctx, &snaps, ctrlruntimeclient.InNamespace(cpTestNamespaceName)); err != nil {
			t.Error(err)
		}

		if len(snaps.Items) == 0 {
			return true, nil
		}

		for _, snap := range snaps.Items {
			s := snap
			if s.ObjectMeta.DeletionTimestamp != nil {
				continue
			}

			if err := c.client.Delete(ctx, &s); err != nil {
				// Make error transient so that we try to remove it again and
				// not leak any resources
				t.Logf("error removing volumesnapshot %q: %v", s.Name, err)

				return false, nil
			}
		}

		return false, nil
	})
	if err != nil {
		t.Fatalf("error waiting for pvc to get removed: %v", err)
	}

	t.Log("Cleaning up PVC...")

	err = wait.PollUntilContextTimeout(c.ctx, cpTestPollPeriod, cpTestTimeout, false, func(ctx context.Context) (done bool, err error) {
		var pvcs corev1.PersistentVolumeClaimList
		if err := c.client.List(ctx, &pvcs, ctrlruntimeclient.InNamespace(cpTestNamespaceName)); err != nil {
			t.Error(err)
		}

		if len(pvcs.Items) == 0 {
			return true, nil
		}

		for _, pvc := range pvcs.Items {
			p := pvc
			if p.ObjectMeta.DeletionTimestamp != nil {
				continue
			}

			if err := c.client.Delete(ctx, &p); err != nil {
				// Make error transient so that we try to remove it again and
				// not leak any resources
				t.Logf("error removing pvc %q: %v", p.Name, err)

				return false, nil
			}
		}

		return false, nil
	})
	if err != nil {
		t.Fatalf("error waiting for pvc to get removed: %v", err)
	}

	t.Log("Cleaning up successful...")
}
