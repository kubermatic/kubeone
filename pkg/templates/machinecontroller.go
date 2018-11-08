package templates

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac_v1beta1 "k8s.io/api/rbac/v1beta1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kubermatic/kubeone/pkg/manifest"
)

const (
	MachineControllerAppLabelKey   = "app"
	MachineControllerAppLabelValue = "machine-controller"
	MachineControllerTag           = "v0.9.9"
)

func MachineControllerConfiguration(manifest *manifest.Manifest, instance int) (string, error) {
	items := []interface{}{
		machineControllerClusterRole(),
		machineControllerClusterRoleBinding(),
		nodeBootstrapperClusterRoleBinding(),
		machineControllerKubeSystemRole(),
		machineControllerKubePublicRole(),
		machineControllerDefaultRole(),
		machineControllerClusterInfoRole(),
		machineControllerKubeSystemRoleBinding(),
		machineControllerKubePublicRoleBinding(),
		machineControllerDefaultRoleBinding(),
		machineControllerClusterInfoRoleBinding(),
		machineControllerMachineCRD(),
		machineControllerClusterCRD(),
		machineControllerMachineSetCRD(),
		machineControllerMachineDeploymentCRD(),
		machineControllerDeployment(),
	}

	return kubernetesToYAML(items)
}

func machineControllerClusterRole() rbac_v1beta1.ClusterRole {
	return rbac_v1beta1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:kubermatic-machine-controller",
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		Rules: []rbac_v1beta1.PolicyRule{
			{
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups:     []string{"apiextensions.k8s.io"},
				Resources:     []string{"customresourcedefinitions"},
				ResourceNames: []string{"machines.machine.k8s.io"},
				Verbs:         []string{"delete"},
			},
			{
				APIGroups:     []string{"apiextensions.k8s.io"},
				Resources:     []string{"customresourcedefinitions"},
				ResourceNames: []string{"machines.machine.k8s.io"},
				Verbs:         []string{"*"},
			},
			{
				APIGroups: []string{"machine.k8s.io"},
				Resources: []string{"machines"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"list", "get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/eviction"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
			{
				APIGroups: []string{"cluster.k8s.io"},
				Resources: []string{"machines", "machinesets", "machinesets/status", "machinedeployments", "machinedeployments/status", "clusters", "clusters/status"},
				Verbs:     []string{"*"},
			},
		},
	}
}

func machineControllerClusterRoleBinding() rbac_v1beta1.ClusterRoleBinding {
	return rbac_v1beta1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:kubermatic-machine-controller:controller",
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbac_v1beta1.RoleRef{
			Name:     "system:kubermatic-machine-controller",
			Kind:     "ClusterRole",
			APIGroup: rbac_v1beta1.GroupName,
		},
		Subjects: []rbac_v1beta1.Subject{
			{
				Kind:     "User",
				Name:     "machine-controller",
				APIGroup: rbac_v1beta1.GroupName,
			},
		},
	}
}

func nodeBootstrapperClusterRoleBinding() rbac_v1beta1.ClusterRoleBinding {
	return rbac_v1beta1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:kubermatic-machine-controller:kubelet-bootstrap",
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbac_v1beta1.RoleRef{
			Name:     "system:node-bootstrapper",
			Kind:     "ClusterRole",
			APIGroup: rbac_v1beta1.GroupName,
		},
		Subjects: []rbac_v1beta1.Subject{
			{
				Kind:     "Group",
				Name:     "system:bootstrappers:machine-controller:default-node-token",
				APIGroup: rbac_v1beta1.GroupName,
			},
		},
	}
}

func nodeSignerClusterRoleBinding() rbac_v1beta1.ClusterRoleBinding {
	return rbac_v1beta1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:kubermatic-machine-controller:node-signer",
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbac_v1beta1.RoleRef{
			Name:     "system:certificates.k8s.io:certificatesigningrequests:nodeclient",
			Kind:     "ClusterRole",
			APIGroup: rbac_v1beta1.GroupName,
		},
		Subjects: []rbac_v1beta1.Subject{
			{
				Kind:     "Group",
				Name:     "system:bootstrappers:machine-controller:default-node-token",
				APIGroup: rbac_v1beta1.GroupName,
			},
		},
	}
}

func machineControllerKubeSystemRole() rbac_v1beta1.Role {
	return rbac_v1beta1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		Rules: []rbac_v1beta1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs: []string{
					"create",
					"update",
					"list",
					"watch",
				},
			},
			{
				APIGroups:     []string{""},
				Resources:     []string{"endpoints"},
				ResourceNames: []string{"machine-controller"},
				Verbs:         []string{"*"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
				Verbs:     []string{"create"},
			},
		},
	}
}

func machineControllerKubePublicRole() rbac_v1beta1.Role {
	return rbac_v1beta1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: metav1.NamespacePublic,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		Rules: []rbac_v1beta1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
		},
	}
}

func machineControllerDefaultRole() rbac_v1beta1.Role {
	return rbac_v1beta1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		Rules: []rbac_v1beta1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
		},
	}
}

func machineControllerClusterInfoRole() rbac_v1beta1.Role {
	return rbac_v1beta1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-info",
			Namespace: metav1.NamespacePublic,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		Rules: []rbac_v1beta1.PolicyRule{
			{
				APIGroups:     []string{""},
				ResourceNames: []string{"cluster-info"},
				Resources:     []string{"configmaps"},
				Verbs:         []string{"get"},
			},
		},
	}
}

func machineControllerKubeSystemRoleBinding() rbac_v1beta1.RoleBinding {
	return rbac_v1beta1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbac_v1beta1.RoleRef{
			Name:     "machine-controller",
			Kind:     "Role",
			APIGroup: rbac_v1beta1.GroupName,
		},
		Subjects: []rbac_v1beta1.Subject{
			{
				Kind:      "User",
				Name:      "machine-controller",
				Namespace: metav1.NamespaceSystem,
				APIGroup:  rbac_v1beta1.GroupName,
			},
		},
	}
}

func machineControllerKubePublicRoleBinding() rbac_v1beta1.RoleBinding {
	return rbac_v1beta1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: metav1.NamespacePublic,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbac_v1beta1.RoleRef{
			Name:     "machine-controller",
			Kind:     "Role",
			APIGroup: rbac_v1beta1.GroupName,
		},
		Subjects: []rbac_v1beta1.Subject{
			{
				Kind:      "User",
				Name:      "machine-controller",
				Namespace: metav1.NamespacePublic,
				APIGroup:  rbac_v1beta1.GroupName,
			},
		},
	}
}

func machineControllerDefaultRoleBinding() rbac_v1beta1.RoleBinding {
	return rbac_v1beta1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbac_v1beta1.RoleRef{
			Name:     "machine-controller",
			Kind:     "Role",
			APIGroup: rbac_v1beta1.GroupName,
		},
		Subjects: []rbac_v1beta1.Subject{
			{
				Kind:      "User",
				Name:      "machine-controller",
				Namespace: metav1.NamespaceDefault,
				APIGroup:  rbac_v1beta1.GroupName,
			},
		},
	}
}

func machineControllerClusterInfoRoleBinding() rbac_v1beta1.RoleBinding {
	return rbac_v1beta1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-info",
			Namespace: metav1.NamespacePublic,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbac_v1beta1.RoleRef{
			Name:     "cluster-info",
			Kind:     "Role",
			APIGroup: rbac_v1beta1.GroupName,
		},
		Subjects: []rbac_v1beta1.Subject{
			{
				Kind:      "User",
				Name:      "cluster-info",
				Namespace: metav1.NamespacePublic,
				APIGroup:  rbac_v1beta1.GroupName,
			},
		},
	}
}

func machineControllerMachineCRD() apiextensionsv1beta1.CustomResourceDefinition {
	return apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "machines.cluster.k8s.io",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "cluster.k8s.io",
			Version: "v1alpha1",
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:     "Machine",
				ListKind: "MachineList",
				Plural:   "machines",
				Singular: "machine",
			},
		},
	}
}

func machineControllerClusterCRD() apiextensionsv1beta1.CustomResourceDefinition {
	return apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "clusters.cluster.k8s.io",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "cluster.k8s.io",
			Version: "v1alpha1",
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:     "Cluster",
				ListKind: "ClusterList",
				Plural:   "clusters",
				Singular: "cluster",
			},
		},
	}
}

func machineControllerMachineSetCRD() apiextensionsv1beta1.CustomResourceDefinition {
	return apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "machinesets.cluster.k8s.io",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "cluster.k8s.io",
			Version: "v1alpha1",
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:     "MachineSet",
				ListKind: "MachineSetList",
				Plural:   "machinesets",
				Singular: "machineset",
			},
		},
	}
}

func machineControllerMachineDeploymentCRD() apiextensionsv1beta1.CustomResourceDefinition {
	return apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "machinedeployments.cluster.k8s.io",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "cluster.k8s.io",
			Version: "v1alpha1",
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:     "MachineDeployment",
				ListKind: "MachineDeploymentList",
				Plural:   "machinedeployments",
				Singular: "machinedeployment",
			},
		},
	}
}

func machineControllerDeployment() appsv1.Deployment {
	var replicas int32 = 1

	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "machine-controller",
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					MachineControllerAppLabelKey: MachineControllerAppLabelValue,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 0,
					},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/path":   "/metrics",
						"prometheus.io/port":   "8085",
					},
				},
				Spec: corev1.PodSpec{
					Tolerations: []corev1.Toleration{
						{
							Key:      "node-role.kubernetes.io/master",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "machine-controller",
							Image:           "docker.io/kubermatic/machine-controller:" + MachineControllerTag,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/usr/local/bin/machine-controller"},
							Args: []string{
								"-logtostderr",
								"-v", "4",
								"-internal-listen-address", "0.0.0.0:8085",
							},
							// TODO(xmudrii): check what do to with vars.
							//Env:                      getEnvVars(data),
							TerminationMessagePath:   corev1.TerminationMessagePathDefault,
							TerminationMessagePolicy: corev1.TerminationMessageReadFile,
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.FromInt(8085),
									},
								},
								FailureThreshold: 3,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								TimeoutSeconds:   15,
							},
							LivenessProbe: &corev1.Probe{
								FailureThreshold: 8,
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/live",
										Port: intstr.FromInt(8085),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								TimeoutSeconds:      15,
							},
						},
					},
				},
			},
		},
	}
}
