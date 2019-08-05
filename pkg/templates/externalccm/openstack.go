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
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/clientutil"
	"github.com/kubermatic/kubeone/pkg/state"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	openstackSAName = "cloud-controller-manager"
	openstackDeploymentName = "openstack-cloud-controller-manager"
	openstackConfigSecretName = "cloud-config"
)

// Starting with v1.15.x the OS cloud provider started pinning their CCM versions.
var ccmVersionMapping = map[string]string{
	"1.15.":  "v1.15.0",
}

func ensureOpenStack(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes client not initialised")
	}

	if s.Cluster.CloudProvider.CloudConfig == "" {
		return errors.New("cloudConfig not defined")
	}

	version := osGetVersion(s.Cluster.Versions.Kubernetes)

	// Quickly check if control plane Operating Systems match.
	var os string
	for _, node := range s.Cluster.Hosts {
		if os != "" && os != node.OperatingSystem {
			return errors.New("Controlplane node operating systems don't match")
		}
		os = node.OperatingSystem
	}

	ctx := context.Background()
	k8sobjects := []runtime.Object{
		osServiceAccount(),
		osSecret(s.Cluster.CloudProvider.CloudConfig),
		osClusterRole(),
		osClusterRoleBinding(),
	}

	for _, obj := range k8sobjects {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj); err != nil {
			return errors.Wrapf(err, "failed to ensure OpenStack CCM %T", obj)
		}
	}

	ds := osDaemonSet(version, os)
	want, err := semver.NewConstraint("<=" + version)
	if err != nil {
		return errors.Wrap(err, "failed to parse OpenStack CCM version constraint")
	}

	_, err = controllerutil.CreateOrUpdate(
		ctx,
		s.DynamicClient,
		ds,
		mutateDaemonsetWithVersionCheck(want),
	)

	if err != nil {
		s.Logger.Warnf("unable to ensure OpenStack CCM Deployment: %v, skipping", err)
	}
	return nil
}

func osServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: openstackSAName,
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func osSecret(cloudConfig string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: openstackConfigSecretName,
			Namespace: metav1.NamespaceSystem,
		},
		StringData: map[string]string{
			"cloud.conf": cloudConfig,
		},
	}
}

func osClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:cloud-controller-manager",
			Annotations: map[string]string{
				"rbac.authorization.kubernetes.io/autoupdate": "true",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes", "configmaps", "secrets", "serviceaccounts"},
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
				Verbs:     []string{"list", "patch", "update", "watch"},
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
		},
	}
}

func osClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:cloud-controller-manager",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Name:     "system:cloud-controller-manager",
			Kind:     "ClusterRole",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      openstackSAName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func osGetVersion(kubeVersion string) string {
	for k, v := range ccmVersionMapping {
		if strings.HasPrefix(kubeVersion, k) {
			return v
		}
	}
	// No version found, fall back to "latest"
	return "latest"
}

func osDaemonSet(version, os string) *appsv1.DaemonSet {
	var (
		revisions int32 = 3
	)
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate

	flexVolumePath := "/usr/libexec/kubernetes/kubelet-plugins/volume/exec"
	caCertsPath := "/etc/ssl/certs"

	if os == "coreos" {
		flexVolumePath = "/var/lib/kubelet/plugins/volume/exec"
		caCertsPath = "/usr/share/ca-certificates"
	}

	var user int64 = 1001

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: openstackDeploymentName,
			Namespace: metav1.NamespaceSystem, 
		},
		Spec: appsv1.DaemonSetSpec{
			RevisionHistoryLimit: &revisions,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "openstack-cloud-controller-manager",
				},
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
					Labels: map[string]string{
						"app": "openstack-cloud-controller-manager",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: openstackSAName,
					Tolerations: []corev1.Toleration{
						{
							Key:      "node-role.kubernetes.io/master",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						{
							Key:    "node.cloudprovider.kubernetes.io/uninitialized",
							Value:  "true",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
					Containers: []corev1.Container{
						{
							Name: "openstack-cloud-controller-manager",
							Image: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:" + version,
							Command: []string{
								"/bin/openstack-cloud-controller-manager",
								"--v=1",
								"--cloud-config=/etc/config/cloud.conf",
								"--cloud-provider=openstack",
								"--use-service-account-credentials=true",
								"--address=127.0.0.1",
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name: "k8s-certs",
									MountPath: "/etc/kubernetes/pki",
									ReadOnly: true,
								},
								corev1.VolumeMount{
									Name: "ca-certs",
									MountPath: "/etc/ssl/certs",
									ReadOnly: true,
								},
								corev1.VolumeMount{
									Name: "cloud-config-volume",
									MountPath: "/etc/config",
									ReadOnly: true,
								},
								corev1.VolumeMount{
									Name: "flexvolume-dir",
									MountPath: "/usr/libexec/kubernetes/kubelet-plugins/volume/exec",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("200m"),
								},
							},
						},
					},
					NodeSelector: map[string]string{
						"node-role.kubernetes.io/master": "",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &user,
					},
					HostNetwork: true,
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "flexvolume-dir",
							VolumeSource: corev1.VolumeSource{
								// TODO: Fix for CoreOS
								HostPath: &corev1.HostPathVolumeSource{
									Path: flexVolumePath,
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						corev1.Volume{
							Name: "k8s-certs",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/etc/kubernetes/pki",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						corev1.Volume{
							Name: "ca-certs",
							VolumeSource: corev1.VolumeSource{
								// TODO: Fix for CoreOS
								HostPath: &corev1.HostPathVolumeSource{
									Path: caCertsPath,
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						corev1.Volume{
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
