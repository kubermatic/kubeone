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
	hetznerCCMVersion     = "v1.3.0"
	hetznerSAName         = "cloud-controller-manager"
	hetznerDeploymentName = "hcloud-cloud-controller-manager"
)

func ensureHetzner(ctx *util.Context) error {
	if ctx.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	bgctx := context.Background()

	sa := hetznerServiceAccount()
	if err := simpleCreateOrUpdate(bgctx, ctx.DynamicClient, sa); err != nil {
		return errors.Wrap(err, "failed to ensure hetzner CCM ServiceAccount")
	}

	crb := hetznerClusterRoleBinding()
	if err := simpleCreateOrUpdate(bgctx, ctx.DynamicClient, crb); err != nil {
		return errors.Wrap(err, "failed to ensure hetzner CCM ClusterRoleBinding")
	}

	dep := hetznerDeployment()
	want, err := semver.NewConstraint("<= " + hetznerCCMVersion)
	if err != nil {
		return errors.Wrap(err, "failed to parse hetzner CCM version constraint")
	}

	_, err = controllerutil.CreateOrUpdate(bgctx,
		ctx.DynamicClient,
		dep,
		mutateDeploymentWithVersionCheck(want))
	if err != nil {
		ctx.Logger.Warnf("unable to ensure hetzner CCM Deployment: %v, skipping", err)
	}

	return nil
}

func hetznerServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hetznerSAName,
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func hetznerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:cloud-controller-manager",
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "cluster-admin",
			Kind:     "ClusterRole",
			APIGroup: rbacv1.GroupName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      hetznerSAName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func hetznerDeployment() *appsv1.Deployment {
	var (
		replicas  int32 = 1
		revisions int32 = 2
	)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hetznerDeploymentName,
			Namespace: metav1.NamespaceSystem,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             &replicas,
			RevisionHistoryLimit: &revisions,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "hcloud-cloud-controller-manager",
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
					Labels: map[string]string{
						"app": "hcloud-cloud-controller-manager",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: hetznerSAName,
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
							Name:  "hcloud-cloud-controller-manager",
							Image: "hetznercloud/hcloud-cloud-controller-manager:" + hetznerCCMVersion,
							Command: []string{
								"/bin/hcloud-cloud-controller-manager",
								"--cloud-provider=hcloud",
								"--leader-elect=false",
								"--allow-untagged-cloud",
							},
							Env: []corev1.EnvVar{
								{
									Name: "NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
								{
									Name: "HCLOUD_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: credentials.SecretName,
											},
											Key: credentials.HetznerTokenKey,
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
