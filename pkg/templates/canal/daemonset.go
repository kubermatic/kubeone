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

package canal

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// daemonSet installs the calico/node container, as well as the Calico CNI plugins and network config on each
// master and worker node in a Kubernetes cluster
func daemonSet(ifacePatch bool, clusterCIDR, installCNIImage, calicoImage, flannelImage string) *appsv1.DaemonSet {
	maxUnavailable := intstr.FromInt(1)
	terminationGracePeriodSeconds := int64(0)
	privileged := true
	optional := true
	fileOrCreate := corev1.HostPathFileOrCreate
	directoryOrCreate := corev1.HostPathDirectoryOrCreate
	bidirectionalMountPropagation := corev1.MountPropagationBidirectional

	commonLabels := map[string]string{
		"k8s-app": "canal",
	}

	flannelEnv := []corev1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name: "FLANNELD_IP_MASQ",
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					Key: "masquerade",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "canal-config",
					},
				},
			},
		},
	}

	if ifacePatch {
		flannelEnv = append(flannelEnv, corev1.EnvVar{
			Name: "FLANNELD_IFACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		})
	}

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "canal",
			Namespace: metav1.NamespaceSystem,
			Labels:    commonLabels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: commonLabels,
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: &maxUnavailable,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: commonLabels,
					Annotations: map[string]string{
						// This, along with the CriticalAddonsOnly toleration below,
						// marks the pod as a critical add-on, ensuring it gets
						// priority scheduling and that its resources are reserved
						// if it ever gets evicted
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"beta.kubernetes.io/os": "linux",
					},
					HostNetwork: true,
					Tolerations: []corev1.Toleration{
						{
							// Make sure canal gets scheduled on all nodes
							Effect:   corev1.TaintEffectNoSchedule,
							Operator: corev1.TolerationOpExists,
						},
						{
							// Mark the pod as a critical add-on for rescheduling
							Key:      "CriticalAddonsOnly",
							Operator: corev1.TolerationOpExists,
						},
						{
							Effect:   corev1.TaintEffectNoExecute,
							Operator: corev1.TolerationOpExists,
						},
					},
					ServiceAccountName: "canal",
					// Minimize downtime during a rolling upgrade or deletion; tell Kubernetes to do a "force
					// deletion": https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					PriorityClassName:             "system-node-critical",
					InitContainers: []corev1.Container{
						{
							// This container installs the Calico CNI binaries
							// and CNI network config file on each node
							Name:  "install-cni",
							Image: installCNIImage,
							Command: []string{
								"/opt/cni/bin/install",
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									// Allow KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT to be overridden for eBPF mode.
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "kubernetes-services-endpoint",
										},
										Optional: &optional,
									},
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "CNI_CONF_NAME",
									Value: "10-canal.conflist",
								},
								{
									Name: "CNI_NETWORK_CONFIG",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "canal-config",
											},
											Key: "cni_network_config",
										},
									},
								},
								{
									Name: "KUBERNETES_NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
								{
									// CNI MTU Config variable
									Name: "CNI_MTU",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "canal-config",
											},
											Key: "veth_mtu",
										},
									},
								},
								{
									// Prevents the container from sleeping forever
									Name:  "SLEEP",
									Value: "false",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cni-bin-dir",
									MountPath: "/host/opt/cni/bin",
								},
								{
									Name:      "cni-net-dir",
									MountPath: "/host/etc/cni/net.d",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "calico-node",
							Image: calicoImage,
							EnvFrom: []corev1.EnvFromSource{
								{
									// Allow KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT to be overridden for eBPF mode.
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "kubernetes-services-endpoint",
										},
										Optional: &optional,
									},
								},
							},
							Env: []corev1.EnvVar{
								{
									// Use Kubernetes API as the backing datastore
									Name:  "DATASTORE_TYPE",
									Value: "kubernetes",
								},
								{
									// Configure route aggregation based on pod CIDR
									Name:  "USE_POD_CIDR",
									Value: "true",
								},
								{
									// Wait for the datastore
									Name:  "WAIT_FOR_DATASTORE",
									Value: "true",
								},
								{
									// Wait for the datastore
									Name: "NODENAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
								{
									// Don't enable BGP
									Name:  "CALICO_NETWORKING_BACKEND",
									Value: "none",
								},
								{
									// Cluster type to identify the deployment type
									Name:  "CLUSTER_TYPE",
									Value: "k8s,canal",
								},
								{
									// Period, in seconds, at which felix re-applies all iptables state
									Name:  "FELIX_IPTABLESREFRESHINTERVAL",
									Value: "60",
								},
								{
									// No IP address needed.
									Name:  "IP",
									Value: "",
								},
								{
									// The default IPv4 pool to create on startup if none exists. Pod IPs will be
									// chosen from this range. Changing this value after installation will have
									// no effect. This should fall within --cluster-cidr
									Name:  "CALICO_IPV4POOL_CIDR",
									Value: clusterCIDR,
								},
								{
									// Disable file logging so kubectl logs works.
									Name:  "CALICO_DISABLE_FILE_LOGGING",
									Value: "true",
								},
								{
									// Set Felix endpoint to host default action to ACCEPT.
									Name:  "FELIX_DEFAULTENDPOINTTOHOSTACTION",
									Value: "ACCEPT",
								},
								{
									// Disable IPv6 on Kubernetes.
									Name:  "FELIX_IPV6SUPPORT",
									Value: "false",
								},
								{
									// Set Felix logging to "info".
									Name:  "FELIX_LOGSEVERITYSCREEN",
									Value: "info",
								},
								{
									Name:  "FELIX_HEALTHENABLED",
									Value: "true",
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Resources: corev1.ResourceRequirements{
								Requests: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU: resource.MustParse("250m"),
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/calico-node",
											"-felix-live",
										},
									},
								},
								PeriodSeconds:       int32(10),
								InitialDelaySeconds: int32(10),
								FailureThreshold:    int32(6),
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/readiness",
										Port: intstr.FromInt(9099),
										Host: "localhost",
									},
								},
								PeriodSeconds: int32(10),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/lib/modules",
									Name:      "lib-modules",
									ReadOnly:  true,
								},
								{
									MountPath: "/run/xtables.lock",
									Name:      "xtables-lock",
									ReadOnly:  false,
								},
								{
									MountPath: "/var/run/calico",
									Name:      "var-run-calico",
									ReadOnly:  false,
								},
								{
									MountPath: "/var/lib/calico",
									Name:      "var-lib-calico",
									ReadOnly:  false,
								},
								{
									MountPath: "/var/run/nodeagent",
									Name:      "policysync",
								},
								{
									MountPath:        "/sys/fs/",
									Name:             "sysfs",
									MountPropagation: &bidirectionalMountPropagation,
								},
							},
						},
						{
							// This container runs flannel using the kube-subnet-mgr backend
							// for allocating subnets.
							Name:  "kube-flannel",
							Image: flannelImage,
							Command: []string{
								"/opt/bin/flanneld",
								"--ip-masq",
								"--kube-subnet-mgr",
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Env: flannelEnv,
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/run/xtables.lock",
									Name:      "xtables-lock",
									ReadOnly:  false,
								},
								{
									Name:      "flannel-cfg",
									MountPath: "/etc/kube-flannel/",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							// Used by calico/node.
							Name: "lib-modules",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/lib/modules",
								},
							},
						},
						{
							Name: "var-run-calico",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/run/calico",
								},
							},
						},
						{
							Name: "var-lib-calico",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/calico",
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
						{
							Name: "sysfs",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/sys/fs/",
									Type: &directoryOrCreate,
								},
							},
						},
						{
							Name: "flannel-cfg",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "canal-config",
									},
								},
							},
						},
						{
							Name: "cni-bin-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/opt/cni/bin",
								},
							},
						},
						{
							Name: "cni-net-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/etc/cni/net.d",
								},
							},
						},
						{
							Name: "policysync",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/run/nodeagent",
									Type: &directoryOrCreate,
								},
							},
						},
					},
				},
			},
		},
	}
}
