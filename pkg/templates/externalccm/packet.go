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

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/clientutil"
	"github.com/kubermatic/kubeone/pkg/credentials"
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
	packetCCMVersion     = "v0.0.4"
	packetSAName         = "cloud-controller-manager"
	packetDeploymentName = "packet-cloud-controller-manager"
)

func ensurePacket(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	ctx := context.Background()
	k8sobjects := []runtime.Object{
		packetServiceAccount(),
		packetClusterRole(),
		packetClusterRoleBinding(),
	}

	for _, obj := range k8sobjects {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj); err != nil {
			return errors.Wrapf(err, "failed to ensure packet CCM %T", obj)
		}
	}

	dep := packetDeployment()
	want, err := semver.NewConstraint("<= " + packetCCMVersion)
	if err != nil {
		return errors.Wrap(err, "failed to parse packet CCM version constraint")
	}

	_, err = controllerutil.CreateOrUpdate(ctx,
		s.DynamicClient,
		dep,
		mutateDeploymentWithVersionCheck(want))
	if err != nil {
		s.Logger.Warnf("unable to ensure packet CCM Deployment: %v, skipping", err)
	}

	return nil
}

func packetServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      packetSAName,
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func packetClusterRole() *rbacv1.ClusterRole {
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
				Verbs:     []string{"list", "patch", "update", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"serviceaccounts"},
				Verbs:     []string{"create"},
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

func packetClusterRoleBinding() *rbacv1.ClusterRoleBinding {
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
				Name:      packetSAName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func packetDeployment() *appsv1.Deployment {
	var (
		replicas int32 = 1
	)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      packetDeploymentName,
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"app": "packet-cloud-controller-manager",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "packet-cloud-controller-manager",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
					Labels: map[string]string{
						"app": "packet-cloud-controller-manager",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: packetSAName,
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
						{
							Key:      "CriticalAddonsOnly",
							Operator: corev1.TolerationOpExists,
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "packet-cloud-controller-manager",
							Image: "packethost/packet-ccm:" + packetCCMVersion,
							Command: []string{
								"./packet-cloud-controller-manager",
								"--cloud-provider=packet",
								"--leader-elect=false",
								"--allow-untagged-cloud=true",
							},
							Env: []corev1.EnvVar{
								{
									Name: "PACKET_AUTH_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: credentials.SecretName,
											},
											Key: credentials.PacketAPIKeyMC,
										},
									},
								},
								{
									Name: "PACKET_PROJECT_ID",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: credentials.SecretName,
											},
											Key: credentials.PacketProjectID,
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("50Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}
