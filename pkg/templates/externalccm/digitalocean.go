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
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/state"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	digitaloceanImageRegistry  = "docker.io"
	digitaloceanImage          = "/digitalocean/digitalocean-cloud-controller-manager:v0.1.23"
	digitaloceanSAName         = "cloud-controller-manager"
	digitaloceanDeploymentName = "digitalocean-cloud-controller-manager"
)

func ensureDigitalOcean(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	ctx := context.Background()
	sa := doServiceAccount()
	crole := doClusterRole()
	image := s.Cluster.RegistryConfiguration.ImageRegistry(digitaloceanImageRegistry) + digitaloceanImage
	k8sobject := []client.Object{
		sa,
		crole,
		genClusterRoleBinding("system:cloud-controller-manager", crole, sa),
		doDeployment(image),
	}

	withLabel := clientutil.WithComponentLabel(ccmComponentLabel)
	for _, obj := range k8sobject {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj, withLabel); err != nil {
			return errors.Wrapf(err, "failed to ensure digitalocean CCM %T", obj)
		}
	}

	return nil
}

func doServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      digitaloceanSAName,
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func doClusterRole() *rbacv1.ClusterRole {
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

func doDeployment(image string) *appsv1.Deployment {
	var (
		replicas  int32 = 1
		revisions int32 = 2
	)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      digitaloceanDeploymentName,
			Namespace: metav1.NamespaceSystem,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             &replicas,
			RevisionHistoryLimit: &revisions,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "digitalocean-cloud-controller-manager",
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
						"app": "digitalocean-cloud-controller-manager",
					},
				},
				Spec: corev1.PodSpec{
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: digitaloceanSAName,
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
							Name:  "digitalocean-cloud-controller-manager",
							Image: image,
							Command: []string{
								"/bin/digitalocean-cloud-controller-manager",
								"--leader-elect=false",
							},
							Env: []corev1.EnvVar{
								{
									Name: "DO_ACCESS_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: credentials.SecretName,
											},
											Key: credentials.DigitalOceanTokenKeyMC,
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
