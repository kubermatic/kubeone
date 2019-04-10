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

package metricsserver

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

// Deploy generate and POST all objects to apiserver
func Deploy(ctx *util.Context) error {
	if ctx.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	objs := []runtime.Object{
		aggregatedMetricsReaderClusterRole(),
		authDelegatorClusterRoleBinding(),
		metricsServerKubeSystemRoleBinding(),
		metricsServerAPIService(),
		metricsServerServiceAccount(),
		metricsServerDeployment(),
		metricsServerService(),
		metricServerClusterRole(),
		metricServerClusterRoleBinding(),
	}

	bgCtx := context.Background()
	for _, o := range objs {
		if err := simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, o); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func aggregatedMetricsReaderClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:aggregated-metrics-reader",
			Labels: map[string]string{
				"rbac.authorization.k8s.io/aggregate-to-view":  "true",
				"rbac.authorization.k8s.io/aggregate-to-edit":  "true",
				"rbac.authorization.k8s.io/aggregate-to-admin": "true",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"metrics.k8s.io"},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
}

func authDelegatorClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "metrics-server:system:auth-delegator",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "system:auth-delegator",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "metrics-server",
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func metricsServerKubeSystemRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metrics-server-auth-reader",
			Namespace: metav1.NamespaceSystem,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "extension-apiserver-authentication-reader",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "metrics-server",
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func metricsServerAPIService() *apiregv1.APIService {
	return &apiregv1.APIService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiregistration.k8s.io/v1",
			Kind:       "APIService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "v1beta1.metrics.k8s.io",
		},
		Spec: apiregv1.APIServiceSpec{
			Service: &apiregv1.ServiceReference{
				Name:      "metrics-server",
				Namespace: metav1.NamespaceSystem,
			},
			Group:                 "metrics.k8s.io",
			Version:               "v1beta1",
			InsecureSkipTLSVerify: true,
			GroupPriorityMinimum:  100,
			VersionPriority:       100,
		},
	}
}

func metricsServerServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metrics-server",
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func metricsServerDeployment() *appsv1.Deployment {
	k8sAppLabels := map[string]string{"k8s-app": "metrics-server"}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metrics-server",
			Namespace: metav1.NamespaceSystem,
			Labels:    k8sAppLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: k8sAppLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "metrics-server",
					Labels: k8sAppLabels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "metrics-server",
					Volumes: []corev1.Volume{
						{
							Name: "tmp-dir",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "metrics-server",
							Image:           "k8s.gcr.io/metrics-server-amd64:v0.3.1",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string{
								"--kubelet-insecure-tls",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "tmp-dir",
									MountPath: "/tmp",
								},
							},
						},
					},
				},
			},
		},
	}
}

func metricsServerService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metrics-server",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"kubernetes.io/name": "Metrics-server",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"k8s-app": "metrics-server",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       443,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(443),
				},
			},
		},
	}
}

func metricServerClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:metrics-server",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "nodes", "nodes/stats"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
}

func metricServerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:metrics-server",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "system:metrics-server",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "metrics-server",
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}
