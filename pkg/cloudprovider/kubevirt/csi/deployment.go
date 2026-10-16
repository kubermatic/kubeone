/*
Copyright 2026 The KubeOne Authors.

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

package csi

import (
	"fmt"
	"maps"

	kubevirtinfra "k8c.io/kubeone/pkg/cloudprovider/kubevirt/infra"
	"k8c.io/kubeone/pkg/templates/images"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	socketDir         = "/csi"
	csiEndpoint       = "unix://" + socketDir + "/csi.sock"
	csiAddress        = socketDir + "/csi.sock"
	tenantKubeconfig  = "/var/run/secrets/tenantcluster/" + kubevirtinfra.KubeconfigSecretKey
	tenantVolumeName  = "tenantcluster"
	socketVolumeName  = "socket-dir"
	healthPort        = 10301
	healthPortName    = "healthz"
	sidecarTimeout    = "--timeout=3m"
	sidecarKubeconfig = "--kubeconfig=" + tenantKubeconfig
)

// ImageGetter resolves an image reference, applying the overwrite registry.
type ImageGetter func(images.Resource, ...images.GetOpt) string

// Deployment returns the CSI controller Deployment running in the infra
// cluster. It accesses the infra cluster with its in-cluster ServiceAccount and
// the tenant cluster with the tenant kubeconfig Secret.
func Deployment(clusterName, namespace string, image ImageGetter) *appsv1.Deployment {
	name := kubevirtinfra.ObjectName(clusterName, Component.Name)
	labels := kubevirtinfra.Labels(clusterName, Component.Name)
	podLabels := maps.Clone(labels)
	podLabels["app.kubernetes.io/name"] = Component.Name

	socketMount := corev1.VolumeMount{Name: socketVolumeName, MountPath: socketDir}
	tenantMount := corev1.VolumeMount{Name: tenantVolumeName, MountPath: "/var/run/secrets/tenantcluster", ReadOnly: true}
	csiAddressArg := "--csi-address=" + csiAddress

	sidecar := func(name string, img images.Resource, args ...string) corev1.Container {
		return corev1.Container{
			Name:         name,
			Image:        image(img),
			Args:         append([]string{csiAddressArg, sidecarKubeconfig, sidecarTimeout}, args...),
			VolumeMounts: []corev1.VolumeMount{socketMount, tenantMount},
			Resources:    requests("10m", "50Mi"),
		}
	}

	return &appsv1.Deployment{
		Name:      name,
		Namespace: namespace,
		Labels:    labels,
		Spec: appsv1.DeploymentSpec{
			Replicas: new(int32(1)),
			Strategy: appsv1.DeploymentStrategy{
				// Never run two controllers at once, sidecars do not use leader election.
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{MatchLabels: podLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: podLabels},
				Spec: corev1.PodSpec{
					ServiceAccountName: Component.ServiceAccountName(clusterName),
					Containers: []corev1.Container{
						{
							Name:  "csi-driver",
							Image: image(images.KubeVirtCSI),
							Args: []string{
								"--endpoint=" + csiEndpoint,
								"--infra-cluster-namespace=" + namespace,
								fmt.Sprintf("--infra-cluster-labels=%s=%s", kubevirtinfra.LabelClusterName, clusterName),
								"--tenant-cluster-kubeconfig=" + tenantKubeconfig,
								"--run-node-service=false",
								"--run-controller-service=true",
								"--v=5",
							},
							Ports: []corev1.ContainerPort{
								{Name: healthPortName, ContainerPort: healthPort, Protocol: corev1.ProtocolTCP},
							},
							LivenessProbe: &corev1.Probe{
								HTTPGet:             &corev1.HTTPGetAction{Path: "/healthz", Port: intstr.FromString(healthPortName)},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      3,
								PeriodSeconds:       10,
								FailureThreshold:    5,
							},
							VolumeMounts: []corev1.VolumeMount{socketMount, tenantMount},
							Resources:    requests("10m", "50Mi"),
						},
						sidecar("csi-provisioner", images.KubeVirtCSIProvisioner, "--default-fstype=ext4", "--retry-interval-max=1m", "--v=5"),
						sidecar("csi-attacher", images.KubeVirtCSIAttacher, "--retry-interval-max=1m", "--v=5"),
						sidecar("csi-snapshotter", images.KubeVirtCSISnapshotter, "--v=5"),
						sidecar("csi-resizer", images.KubeVirtCSIResizer, "--handle-volume-inuse-error=false", "--v=5"),
						{
							Name:  "csi-liveness-probe",
							Image: image(images.KubeVirtCSILivenessProbe),
							Args: []string{
								csiAddressArg,
								"--probe-timeout=3s",
								fmt.Sprintf("--health-port=%d", healthPort),
							},
							VolumeMounts: []corev1.VolumeMount{socketMount},
							Resources:    requests("10m", "20Mi"),
						},
					},
					Volumes: []corev1.Volume{
						{
							Name:     socketVolumeName,
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
						{
							Name: tenantVolumeName,
							Secret: &corev1.SecretVolumeSource{
								SecretName: Component.TenantKubeconfigSecretName(clusterName),
							},
						},
					},
				},
			},
		},
	}
}

func requests(cpu, memory string) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpu),
			corev1.ResourceMemory: resource.MustParse(memory),
		},
	}
}
