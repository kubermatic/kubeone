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

package dnscache

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	imagetag         = "k8s.gcr.io/k8s-dns-node-cache:1.15.2"
	dnscacheCorefile = `
cluster.local:53 {
	errors
	cache {
		success 9984 30
		denial 9984 5
	}
	reload
	loop
	bind 169.254.20.10
	forward . 10.96.0.10 {
		force_tcp
	}
	prometheus :9253
	health 169.254.20.10:8080
}

in-addr.arpa:53 {
	errors
	cache 30
	reload
	loop
	bind 169.254.20.10
	forward . 10.96.0.10 {
		force_tcp
	}
	prometheus :9253
}

ip6.arpa:53 {
	errors
	cache 30
	reload
	loop
	bind 169.254.20.10
	forward . 10.96.0.10 {
		force_tcp
	}
	prometheus :9253
}

.:53 {
	errors
	cache 30
	reload
	loop
	bind 169.254.20.10
	forward . /etc/resolv.conf {
		force_tcp
	}
	prometheus :9253
}
`
)

// Deploy generate and POST all objects to apiserver
func Deploy(ctx *util.Context) error {
	if ctx.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	objs := []runtime.Object{
		dnscacheServiceAccount(),
		dnscacheConfigMap(),
		dnscacheDaemonSet(),
	}

	bgCtx := context.Background()
	for _, o := range objs {
		if err := simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, o); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func dnscacheServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "node-local-dns",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"kubernetes.io/cluster-service": "true",
			},
		},
	}
}

func dnscacheConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "node-local-dns",
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			"Corefile": dnscacheCorefile,
		},
	}
}

func dnscacheDaemonSet() *appsv1.DaemonSet {
	maxUnavailable := intstr.FromString("30%")
	k8sAppLabels := map[string]string{"k8s-app": "node-local-dns"}
	trueBool := true
	hostPathFileOrCreate := corev1.HostPathFileOrCreate
	terminationGracePeriodSeconds := int64(0)

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "node-local-dns",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"k8s-app":                       "node-local-dns",
				"kubernetes.io/cluster-service": "true",
			},
		},
		Spec: appsv1.DaemonSetSpec{
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: &maxUnavailable,
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: k8sAppLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: k8sAppLabels,
				},
				Spec: corev1.PodSpec{
					PriorityClassName:  "system-node-critical",
					ServiceAccountName: "node-local-dns",
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSDefault,
					Tolerations: []corev1.Toleration{
						{
							Operator: corev1.TolerationOpExists,
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "node-cache",
							Image: imagetag,
							Args: []string{
								"-localip",
								"169.254.20.10",
								"-conf",
								"/etc/coredns/Corefile",
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("25m"),
									corev1.ResourceMemory: resource.MustParse("5Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("25m"),
									corev1.ResourceMemory: resource.MustParse("30Mi"),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &trueBool,
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 53,
									Name:          "dns",
									Protocol:      corev1.ProtocolUDP,
								},
								{
									ContainerPort: 53,
									Name:          "dns-tcp",
									Protocol:      corev1.ProtocolTCP,
								},
								{
									ContainerPort: 9253,
									Name:          "metrics",
									Protocol:      corev1.ProtocolTCP,
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Host: "169.254.20.10",
										Path: "/health",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 60,
								TimeoutSeconds:      5,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "xtables-lock",
									MountPath: "/run/xtables.lock",
									ReadOnly:  false,
								},
								{
									Name:      "config-volume",
									MountPath: "/etc/coredns",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "xtables-lock",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/run/xtables.lock",
									Type: &hostPathFileOrCreate,
								},
							},
						},
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "node-local-dns",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "Corefile",
											Path: "Corefile",
										},
									},
								},
							},
						},
					},
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
				},
			},
		},
	}
}
