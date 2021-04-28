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
	"encoding/json"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/images"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type packetCloudSA struct {
	APIKey    string `json:"apiKey"`
	ProjectID string `json:"projectID"`
}

const (
	packetSAName            = "cloud-controller-manager"
	packetDeploymentName    = "packet-cloud-controller-manager"
	packetCloudSASecretName = "packet-cloud-config"
)

func ensurePacket(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	ctx := context.Background()
	sa := packetServiceAccount()
	crole := packetClusterRole()

	creds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to fetch credentials")
	}
	secret, err := packetCloudSASecret(creds)
	if err != nil {
		return errors.Wrap(err, "failed to generate packet cloud config secret")
	}

	ccmImage := s.Images.Get(images.PacketCCM)
	k8sobjects := []client.Object{
		sa,
		crole,
		genClusterRoleBinding("system:cloud-controller-manager", crole, sa),
		secret,
		packetDeployment(ccmImage),
	}

	withLabel := clientutil.WithComponentLabel(ccmComponentLabel)
	for _, obj := range k8sobjects {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj, withLabel); err != nil {
			return errors.Wrapf(err, "failed to ensure packet CCM %T", obj)
		}
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

func packetCloudSASecret(creds map[string]string) (*corev1.Secret, error) {
	cloudSA := &packetCloudSA{
		APIKey:    creds[credentials.PacketAPIKeyMC],
		ProjectID: creds[credentials.PacketProjectID],
	}
	b, err := json.Marshal(cloudSA)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal cloud-sa to json")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      packetCloudSASecretName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string][]byte{
			"cloud-sa.json": b,
		},
	}, nil
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

func packetDeployment(image string) *appsv1.Deployment {
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
							Image: image,
							Command: []string{
								"./packet-cloud-controller-manager",
								"--cloud-provider=packet",
								"--leader-elect=false",
								"--allow-untagged-cloud=true",
								"--authentication-skip-lookup=true",
								"--provider-config=/etc/cloud-sa/cloud-sa.json",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cloud-sa-volume",
									ReadOnly:  true,
									MountPath: "/etc/cloud-sa",
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
					Volumes: []corev1.Volume{
						{
							Name: "cloud-sa-volume",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: packetCloudSASecretName,
								},
							},
						},
					},
				},
			},
		},
	}
}
