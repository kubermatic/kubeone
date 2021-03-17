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
	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/state"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	vSphereSAName           = "cloud-controller-manager"
	vSphereDeploymentName   = "vsphere-cloud-controller-manager"
	vSphereConfigSecretName = "cloud-config" //nolint:gosec
	vSphereImageRegistry    = "gcr.io"
	vSphereImage            = "/cloud-provider-vsphere/cpi/release/manager:v1.2.1"
)

func ensureVsphere(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	image := s.Cluster.RegistryConfiguration.ImageRegistry(vSphereImageRegistry) + vSphereImage

	k8sobjects := []client.Object{
		vSphereServiceAccount(),
		vSphereSecret(s.Cluster.CloudProvider.CloudConfig),
		vSphereClusterRole(),
		vSphereClusterRoleBinding(),
		vSphereRoleBinding(),
		vSphereDaemonSet(image),
		vSphereService(),
	}

	withLabel := clientutil.WithComponentLabel(ccmComponentLabel)
	for _, obj := range k8sobjects {
		if err := clientutil.CreateOrUpdate(s.Context, s.DynamicClient, obj, withLabel); err != nil {
			return errors.Wrapf(err, "failed to ensure vSphere CCM %T", obj)
		}
	}

	return nil
}

func vSphereServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vSphereSAName,
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func vSphereSecret(cloudConfig string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vSphereConfigSecretName,
			Namespace: metav1.NamespaceSystem,
		},
		StringData: map[string]string{
			"vsphere.conf": cloudConfig,
		},
	}
}

func vSphereClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:cloud-controller-manager",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes/status"},
				Verbs:     []string{"patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"list", "patch", "update", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services/status"},
				Verbs:     []string{"patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"serviceaccounts"},
				Verbs:     []string{"create", "get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "update", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
				Verbs:     []string{"create", "get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
}

func vSphereRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "servicecatalog.k8s.io:apiserver-authentication-reader",
			Namespace: metav1.NamespaceSystem,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     "extension-apiserver-authentication-reader",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      vSphereSAName,
				Namespace: metav1.NamespaceSystem,
			},
			{
				Kind: "User",
				Name: vSphereSAName,
			},
		},
	}
}

func vSphereClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:cloud-controller-manager",
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "system:cloud-controller-manager",
			Kind:     "ClusterRole",
			APIGroup: rbacv1.GroupName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      vSphereSAName,
				Namespace: metav1.NamespaceSystem,
			},
			{
				Kind: "User",
				Name: vSphereSAName,
			},
		},
	}
}
func vSphereService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vSphereDeploymentName,
			Namespace: metav1.NamespaceSystem,
			Labels:    map[string]string{"component": vSphereDeploymentName},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"component": vSphereDeploymentName},
			Type:     corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Port:       43001,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(43001),
				},
			},
		},
	}
}

func vSphereDaemonSet(image string) *appsv1.DaemonSet {
	var (
		runAsUser int64 = 1001
		vslabels        = map[string]string{"k8s-app": vSphereDeploymentName}
	)

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vSphereDeploymentName,
			Namespace: metav1.NamespaceSystem,
			Labels:    vslabels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: vslabels,
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
					Labels: vslabels,
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"node-role.kubernetes.io/master": "",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &runAsUser,
					},
					Tolerations: []corev1.Toleration{
						{
							Key:    "node-role.kubernetes.io/master",
							Effect: corev1.TaintEffectNoSchedule,
						},
						{
							Key:    "node.cloudprovider.kubernetes.io/uninitialized",
							Value:  "true",
							Effect: corev1.TaintEffectNoSchedule,
						},
						{
							Key:      "node.kubernetes.io/not-ready",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
					ServiceAccountName: vSphereSAName,
					Containers: []corev1.Container{
						{
							Name:  "vsphere-cloud-controller-manager",
							Image: image,
							Args: []string{
								"--v=2",
								"--cloud-provider=vsphere",
								"--cloud-config=/etc/cloud/vsphere.conf",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/etc/cloud",
									Name:      "vsphere-config-volume",
									ReadOnly:  true,
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("200m"),
								},
							},
						},
					},
					HostNetwork: true,
					Volumes: []corev1.Volume{
						{
							Name: "vsphere-config-volume",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: vSphereConfigSecretName,
								},
							},
						},
					},
				},
			},
		},
	}
}
