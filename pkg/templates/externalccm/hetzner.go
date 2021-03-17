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
	"fmt"

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
	hetznerImageRegistry  = "docker.io"
	hetznerImage          = "/hetznercloud/hcloud-cloud-controller-manager:v1.8.1"
	hetznerSAName         = "cloud-controller-manager"
	hetznerDeploymentName = "hcloud-cloud-controller-manager"
)

func ensureHetzner(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	ctx := context.Background()
	image := s.Cluster.RegistryConfiguration.ImageRegistry(hetznerImageRegistry) + hetznerImage
	k8sobject := []client.Object{
		hetznerServiceAccount(),
		hetznerClusterRoleBinding(),
		hetznerDeployment(s.Cluster.CloudProvider.Hetzner.NetworkID, s.Cluster.ClusterNetwork.PodSubnet, image),
	}

	withLabel := clientutil.WithComponentLabel(ccmComponentLabel)
	for _, obj := range k8sobject {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj, withLabel); err != nil {
			return errors.Wrapf(err, "failed to ensure hetzner CCM %T", obj)
		}
	}

	return nil
}

func hetznerServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hetznerSAName,
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func hetznerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
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

func hetznerDeployment(networkID, podSubnet, image string) *appsv1.Deployment {
	var (
		replicas  int32 = 1
		revisions int32 = 2
		cmd             = []string{
			"/bin/hcloud-cloud-controller-manager",
			"--cloud-provider=hcloud",
			"--leader-elect=false",
			"--allow-untagged-cloud",
		}
	)
	if len(networkID) > 0 {
		cmd = append(cmd, "--allocate-node-cidrs=true")
		cmd = append(cmd, fmt.Sprintf("--cluster-cidr=%s", podSubnet))
	}

	return &appsv1.Deployment{
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
							Name:    "hcloud-cloud-controller-manager",
							Image:   image,
							Command: cmd,
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
											Key: credentials.HetznerTokenKeyMC,
										},
									},
								},
								{
									Name:  "HCLOUD_NETWORK",
									Value: networkID,
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
