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

package nodelocaldns

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/state"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VirtualIP that will be used as DNS
const VirtualIP = "169.254.20.10"

const (
	imageRegistry            = "k8s.gcr.io"
	image                    = "/k8s-dns-node-cache:"
	tag                      = "1.15.13"
	componentLabel           = "nodelocaldns"
	dnscacheCorefileTemplate = `
__PILLAR__DNS__DOMAIN__:53 {
	errors
	cache {
		success 9984 30
		denial 9984 5
	}
	reload
	loop
	bind __PILLAR__LOCAL__DNS__
	forward . __PILLAR__CLUSTER__DNS__ {
		force_tcp
	}
	prometheus :9253
	health __PILLAR__LOCAL__DNS__:8080
}
in-addr.arpa:53 {
	errors
	cache 30
	reload
	loop
	bind __PILLAR__LOCAL__DNS__
	forward . __PILLAR__CLUSTER__DNS__ {
		force_tcp
	}
	prometheus :9253
}
ip6.arpa:53 {
	errors
	cache 30
	reload
	loop
	bind __PILLAR__LOCAL__DNS__
	forward . __PILLAR__CLUSTER__DNS__ {
		force_tcp
	}
	prometheus :9253
}
.:53 {
	errors
	cache 30
	reload
	loop
	bind __PILLAR__LOCAL__DNS__
	forward . __PILLAR__UPSTREAM__SERVERS__ {
		force_tcp
	}
	prometheus :9253
}
`
)

// Deploy generate and POST all objects to apiserver
func Deploy(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	s.Logger.Infoln("Ensure node local DNS cache...")

	image := s.Cluster.RegistryConfiguration.ImageRegistry(imageRegistry) + image + tag

	objs := []client.Object{
		dnscacheServiceAccount(),
		dnscacheService(),
		dnscachePrometheusService(),
		dnscacheConfigMap(s.Cluster.ClusterNetwork.ServiceDomainName),
		dnscacheDaemonSet(image),
	}

	ctx := context.Background()
	withLabel := clientutil.WithComponentLabel(componentLabel)
	for _, o := range objs {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, o, withLabel); err != nil {
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
				"kubernetes.io/cluster-service":   "true",
				"addonmanager.kubernetes.io/mode": "Reconcile",
			},
		},
	}
}

func dnscachePrometheusService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "node-local-dns",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"k8s-app": "node-local-dns",
			},
			Annotations: map[string]string{
				"prometheus.io/port":   "9253",
				"prometheus.io/scrape": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Selector: map[string]string{
				"k8s-app": "node-local-dns",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Port:       9253,
					TargetPort: intstr.FromInt(9253),
				},
			},
		},
	}
}

func dnscacheService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-dns-upstream",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"k8s-app":                         "kube-dns",
				"kubernetes.io/cluster-service":   "true",
				"addonmanager.kubernetes.io/mode": "Reconcile",
				"kubernetes.io/name":              "KubeDNSUpstream",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"k8s-app": "kube-dns",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "dns",
					Port:     53,
					Protocol: corev1.ProtocolUDP,
				},
				{
					Name:     "dns-tcp",
					Port:     53,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
}

func dnscacheConfigMap(pillarDNSDomain string) *corev1.ConfigMap {
	corefile := strings.ReplaceAll(dnscacheCorefileTemplate, "__PILLAR__DNS__DOMAIN__", pillarDNSDomain)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "node-local-dns",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"addonmanager.kubernetes.io/mode": "Reconcile",
			},
		},
		Data: map[string]string{
			"Corefile": corefile,
		},
	}
}

func dnscacheDaemonSet(image string) *appsv1.DaemonSet {
	maxUnavailable := intstr.FromString("10%")
	k8sAppLabels := map[string]string{"k8s-app": "node-local-dns"}
	trueBool := true
	hostPathFileOrCreate := corev1.HostPathFileOrCreate
	terminationGracePeriodSeconds := int64(0)

	// "sleep 10" is needed to avoid race-condition for iptables creation between node-local-cache and kube-proxy, to
	// make sure that kube-proxy will insert its iptables rules first.
	// see more at: https://github.com/kubermatic/kubeone/pull/1058
	execScript := fmt.Sprintf(`
sleep 10;
exec /node-cache -localip %s -conf /etc/Corefile -upstreamsvc kube-dns-upstream`,
		VirtualIP)

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "node-local-dns",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				"k8s-app":                         "node-local-dns",
				"kubernetes.io/cluster-service":   "true",
				"addonmanager.kubernetes.io/mode": "Reconcile",
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
							Image: image,
							Command: []string{
								"/bin/sh",
								"-c",
								execScript,
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("25m"),
									corev1.ResourceMemory: resource.MustParse("5Mi"),
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
										Host: VirtualIP,
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
								{
									Name:      "kube-dns-config",
									MountPath: "/etc/kube-dns",
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
											Path: "Corefile.base",
										},
									},
								},
							},
						},
						{
							Name: "kube-dns-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "kube-dns",
									},
									Optional: &trueBool,
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
