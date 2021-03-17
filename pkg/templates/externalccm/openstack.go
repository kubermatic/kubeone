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
	"context"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/state"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	openstackSAName           = "cloud-controller-manager"
	openstackDeploymentName   = "openstack-cloud-controller-manager"
	openstackConfigSecretName = "cloud-config" //nolint:gosec
	openstackImageRegistry    = "docker.io"
	openstackImage            = "/k8scloudprovider/openstack-cloud-controller-manager:v1.17.0"
)

func ensureOpenStack(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes client not initialised")
	}

	if s.Cluster.CloudProvider.CloudConfig == "" {
		return errors.New("cloudConfig not defined")
	}

	ctx := context.Background()

	sa := osServiceAccount()
	ccmRole := osCCMClusterRole()

	image := s.Cluster.RegistryConfiguration.ImageRegistry(openstackImageRegistry) + openstackImage

	k8sobjects := []client.Object{
		sa,
		osSecret(s.Cluster.CloudProvider.CloudConfig),
		ccmRole,
		genClusterRoleBinding("system:cloud-controller-manager", ccmRole, sa),
		osDaemonSet(image),
	}

	withLabel := clientutil.WithComponentLabel(ccmComponentLabel)
	for _, obj := range k8sobjects {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj, withLabel); err != nil {
			return errors.Wrapf(err, "failed to ensure OpenStack CCM %T", obj)
		}
	}

	return nil
}

func osServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      openstackSAName,
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func osSecret(cloudConfig string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      openstackConfigSecretName,
			Namespace: metav1.NamespaceSystem,
		},
		StringData: map[string]string{
			"cloud.conf": cloudConfig,
		},
	}
}

func osCCMClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:cloud-controller-manager",
			Annotations: map[string]string{
				"rbac.authorization.kubernetes.io/autoupdate": "true",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"get", "create", "update"},
			},
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
				Resources: []string{"serviceaccounts"},
				Verbs:     []string{"create", "get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
				Verbs:     []string{"create", "get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"list", "get", "watch"},
			},
		},
	}
}

func osDaemonSet(image string) *appsv1.DaemonSet {
	var (
		osLabels                        = map[string]string{"k8s-app": openstackDeploymentName}
		runAsUser                 int64 = 1001
		hostPathDirectoryOrCreate       = corev1.HostPathDirectoryOrCreate

		caCertsPath = "/etc/ssl/certs"
	)

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      openstackDeploymentName,
			Namespace: metav1.NamespaceSystem,
			Labels:    osLabels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: osLabels,
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
					Labels: osLabels,
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
					},
					ServiceAccountName: openstackSAName,
					Containers: []corev1.Container{
						{
							Name:  "openstack-cloud-controller-manager",
							Image: image,
							Command: []string{
								"/bin/openstack-cloud-controller-manager",
								"--v=1",
								"--cloud-config=/etc/config/cloud.conf",
								"--cloud-provider=openstack",
								"--use-service-account-credentials=true",
								"--address=127.0.0.1",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "k8s-certs",
									MountPath: "/etc/kubernetes/pki",
									ReadOnly:  true,
								},
								{
									Name:      "ca-certs",
									MountPath: "/etc/ssl/certs",
									ReadOnly:  true,
								},
								{
									Name:      "cloud-config-volume",
									MountPath: "/etc/config",
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
							Name: "k8s-certs",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/etc/kubernetes/pki",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "ca-certs",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: caCertsPath,
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "cloud-config-volume",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: openstackConfigSecretName,
								},
							},
						},
					},
				},
			},
		},
	}
}
