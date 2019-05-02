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

	"github.com/kubermatic/kubeone/pkg/util"
	"github.com/kubermatic/kubeone/pkg/util/credentials"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	packetCCMVersion     = "v0.0.4"
	packetSAName         = "cloud-controller-manager"
	packetDeploymentName = "packet-cloud-controller-manager"
)

func ensurePacket(ctx *util.Context) error {
	if ctx.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	bgctx := context.Background()

	sa := packetServiceAccount()
	if err := simpleCreateOrUpdate(bgctx, ctx.DynamicClient, sa); err != nil {
		return errors.Wrap(err, "failed to ensure packet CCM ServiceAccount")
	}

	cr := packetClusterRole()
	if err := simpleCreateOrUpdate(bgctx, ctx.DynamicClient, cr); err != nil {
		return errors.Wrap(err, "failed to ensure packet CCM ClusterRole")
	}

	crb := packetClusterRoleBinding()
	if err := simpleCreateOrUpdate(bgctx, ctx.DynamicClient, crb); err != nil {
		return errors.Wrap(err, "failed to ensure packet CCM ClusterRoleBinding")
	}

	dep := packetDeployment()
	want, err := semver.NewConstraint("<= " + packetCCMVersion)
	if err != nil {
		return errors.Wrap(err, "failed to parse packet CCM version constraint")
	}

	_, err = controllerutil.CreateOrUpdate(bgctx,
		ctx.DynamicClient,
		dep,
		mutateDeploymentWithVersionCheck(want))
	if err != nil {
		ctx.Logger.Warnf("unable to ensure packet CCM Deployment: %v, skipping", err)
	}

	return nil
}

func packetServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      packetSAName,
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func packetClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
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
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
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
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
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
											Key: credentials.PacketAPIKey,
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
