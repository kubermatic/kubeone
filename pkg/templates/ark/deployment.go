package ark

import (
	"github.com/kubermatic/kubeone/pkg/config"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// deployment deploys Ark version 0.10.0 using default settings
func deployment(cluster *config.Cluster) *appsv1.Deployment {
	replicas := int32(1)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ark",
			Namespace: arkNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"component": "ark",
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "8085",
						"prometheus.io/path":   "/metrics",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyAlways,
					ServiceAccountName: "ark",
					Containers: []corev1.Container{
						{
							Name:  "ark",
							Image: arkContainerImage,
							Command: []string{
								"/ark",
							},
							Args: []string{
								"server",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cloud-credentials",
									MountPath: "/credentials",
								},
								{
									Name:      "plugins",
									MountPath: "/plugins",
								},
								{
									Name:      "scratch",
									MountPath: "/scratch",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "AWS_SHARED_CREDENTIALS_FILE",
									Value: "/credentials/cloud",
								},
								{
									Name:  "ARK_SCRATCH_DIR",
									Value: "/scratch",
								},
								{
									Name:  "AWS_CLUSTER_NAME",
									Value: cluster.Name,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "cloud-credentials",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "cloud-credentials",
								},
							},
						},
						{
							Name: "plugins",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "scratch",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

func resticDaemonset() *appsv1.DaemonSet {
	user := int64(0)
	hostToContainer := corev1.MountPropagationHostToContainer
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "restic",
			Namespace: arkNamespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "restic",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "restic",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: arkServiceAccount,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &user,
					},
					Volumes: []corev1.Volume{
						{
							Name: "cloud-credentials",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "cloud-credentials",
								},
							},
						},
						{
							Name: "host-pods",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/pods",
								},
							},
						},
						{
							Name: "scratch",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "ark",
							Image: arkContainerImage,
							Command: []string{
								"/ark",
							},
							Args: []string{
								"restic",
								"server",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cloud-credentials",
									MountPath: "/credentials",
								},
								{
									Name:             "host-pods",
									MountPath:        "/host_pods",
									MountPropagation: &hostToContainer,
								},
								{
									Name:      "scratch",
									MountPath: "/scratch",
								},
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
									Name: "HEPTIO_ARK_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name:  "AWS_SHARED_CREDENTIALS_FILE",
									Value: "/credentials/cloud",
								},
								{
									Name:  "ARK_SCRATCH_DIR",
									Value: "/scratch",
								},
							},
						},
					},
				},
			},
		},
	}
}
