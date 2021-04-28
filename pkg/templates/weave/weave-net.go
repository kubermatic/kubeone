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

package weave

import (
	"crypto/rand"
	"encoding/base64"
	"strings"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/images"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	componentLabel = "weave-net"
)

// Deploy ensure weave-net resources exists in the cluster
func Deploy(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes dynamic client is not initialized")
	}

	ctx := s.Context

	if s.Cluster.ClusterNetwork.CNI.WeaveNet.Encrypted {
		pass, err := genPassword()
		if err != nil {
			return errors.Wrap(err, "failed to generate random password")
		}

		sec := secret(pass)
		key := client.ObjectKey{
			Name:      sec.GetName(),
			Namespace: sec.GetNamespace(),
		}

		secCopy := sec.DeepCopy()
		err = s.DynamicClient.Get(ctx, key, secCopy)
		switch {
		case k8serrors.IsNotFound(err):
			err = s.DynamicClient.Create(ctx, sec)
			if err != nil {
				return errors.Wrap(err, "failed to create weave-net Secret")
			}
		case err != nil:
			return errors.Wrap(err, "failed to get weave-net Secret")
		}
	}

	var peers []string
	for _, h := range s.Cluster.ControlPlane.Hosts {
		peers = append(peers, h.PrivateAddress)
	}

	kubeImage := s.Images.Get(images.WeaveNetCNIKube)
	npcImage := s.Images.Get(images.WeaveNetCNINPC)

	ds := daemonSet(s.Cluster.ClusterNetwork.CNI.WeaveNet.Encrypted, strings.Join(peers, " "), s.Cluster.ClusterNetwork.PodSubnet, kubeImage, npcImage)

	k8sobjects := []client.Object{
		serviceAccount(),
		clusterRole(),
		clusterRoleBinding(),
		role(),
		roleBinding(),
		ds,
	}

	withLabel := clientutil.WithComponentLabel(componentLabel)
	for _, obj := range k8sobjects {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj, withLabel); err != nil {
			return errors.Wrapf(err, "failed to ensure weave %s", obj.GetObjectKind().GroupVersionKind().Kind)
		}
	}

	return nil
}

func genPassword() (string, error) {
	pi := make([]byte, 32)
	_, err := rand.Reader.Read(pi)
	if err != nil {
		return "", errors.Wrap(err, "failed to read random bytes")
	}
	return base64.StdEncoding.EncodeToString(pi), nil
}

func serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-net",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"name": "weave-net",
			},
		},
	}
}

func clusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "weave-net",
			Labels: map[string]string{
				"name": "weave-net",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "nodes", "namespaces"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"networkpolicies"},
				Verbs:     []string{"watch", "list", "get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes/status"},
				Verbs:     []string{"patch", "update"},
			},
		},
	}
}

func clusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "weave-net",
			Labels: map[string]string{
				"name": "weave-net",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "weave-net",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "weave-net",
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func role() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-net",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"name": "weave-net",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{""},
				Resources:     []string{"configmaps"},
				ResourceNames: []string{"weave-net"},
				Verbs:         []string{"get", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"create"},
			},
		},
	}
}

func roleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-net",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"name": "weave-net",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     "weave-name",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "weave-net",
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func secret(pass string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-passwd",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"name": "weave-net",
			},
		},
		StringData: map[string]string{
			"weave-passwd": pass,
		},
	}
}

func dsEnv(passwordRef bool, peers string, podsubnet string) []corev1.EnvVar {
	env := []corev1.EnvVar{
		{
			Name: "HOSTNAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "spec.nodeName",
				},
			},
		},
		{
			Name:  "WEAVE_METRICS_ADDR",
			Value: "127.0.0.1:6782",
		},
		{
			Name:  "CHECKPOINT_DISABLE",
			Value: "1",
		},
		{
			Name:  "KUBE_PEERS",
			Value: peers,
		},
		{
			Name:  "IPALLOC_RANGE",
			Value: podsubnet,
		},
	}

	if passwordRef {
		env = append(env, corev1.EnvVar{
			Name: "WEAVE_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "weave-passwd",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "weave-passwd",
					},
				},
			},
		})
	}

	return env
}

func daemonSet(passwordRef bool, peers, podsubnet, kubeImage, npcImage string) *appsv1.DaemonSet {
	var (
		priviledged  = true
		fileOrCreate = corev1.HostPathFileOrCreate
	)

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-net",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"name": "weave-net",
			},
		},
		Spec: appsv1.DaemonSetSpec{
			MinReadySeconds: 5,
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "weave-net",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "weave-net",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "weave",
							Command: []string{"/home/weave/launch.sh"},
							Env:     dsEnv(passwordRef, peers, podsubnet),
							Image:   kubeImage,
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Host: "127.0.0.1",
										Path: "/status",
										Port: intstr.FromInt(6784),
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("10m"),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &priviledged,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "weavedb",
									MountPath: "/weavedb",
								},
								{
									Name:      "cni-bin",
									MountPath: "/host/opt",
								},
								{
									Name:      "cni-bin2",
									MountPath: "/host/home",
								},
								{
									Name:      "cni-conf",
									MountPath: "/host/etc",
								},
								{
									Name:      "dbus",
									MountPath: "/host/var/lib/dbus",
								},
								{
									Name:      "lib-modules",
									MountPath: "/lib/modules",
								},
								{
									Name:      "xtables-lock",
									MountPath: "/run/xtables.lock",
								},
							},
						},
						{
							Name: "weave-npc",
							Env: []corev1.EnvVar{
								{
									Name: "HOSTNAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
							},
							Image: npcImage,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("10m"),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &priviledged,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "xtables-lock",
									MountPath: "/run/xtables.lock",
								},
							},
						},
					},
					HostNetwork:        true,
					HostPID:            true,
					RestartPolicy:      corev1.RestartPolicyAlways,
					ServiceAccountName: "weave-net",
					Tolerations: []corev1.Toleration{
						{
							Effect:   corev1.TaintEffectNoSchedule,
							Operator: corev1.TolerationOpExists,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "weavedb",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/weave",
								},
							},
						},
						{
							Name: "cni-bin",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/opt",
								},
							},
						},
						{
							Name: "cni-bin2",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/home",
								},
							},
						},
						{
							Name: "cni-conf",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/etc",
								},
							},
						},
						{
							Name: "dbus",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/dbus",
								},
							},
						},
						{
							Name: "lib-modules",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/lib/modules",
								},
							},
						},
						{
							Name: "xtables-lock",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/run/xtables.lock",
									Type: &fileOrCreate,
								},
							},
						},
					},
				},
			},
		},
	}
}
