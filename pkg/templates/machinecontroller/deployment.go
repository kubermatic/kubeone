package machinecontroller

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1types "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/kubernetes/pkg/registry/core/service/ipallocator"
)

// MachineController related constants
const (
	MachineControllerNamespace     = metav1.NamespaceSystem
	MachineControllerAppLabelKey   = "app"
	MachineControllerAppLabelValue = "machine-controller"
	MachineControllerTag           = "v1.0.4"

	MachineControllerCredentialsSecretName = "machine-controller-credentials"
)

// Deploy deploys MachineController deployment with RBAC on the cluster
func Deploy(ctx *util.Context) error {
	if ctx.Clientset == nil {
		return errors.New("kubernetes clientset not initialized")
	}
	if ctx.APIExtensionClientset == nil {
		return errors.New("kubernetes apiextension clientset not initialized")
	}

	coreClient := ctx.Clientset.CoreV1()
	rbacClient := ctx.Clientset.RbacV1()

	// ServiceAccounts
	sa := machineControllerServiceAccount()
	err := templates.EnsureServiceAccount(coreClient.ServiceAccounts(sa.Namespace), sa)
	if err != nil {
		return err
	}

	// ClusterRoles
	err = templates.EnsureClusterRole(rbacClient.ClusterRoles(), machineControllerClusterRole())
	if err != nil {
		return err
	}

	// ClusterRoleBindings
	crbGenerators := []func() *rbacv1.ClusterRoleBinding{
		nodeSignerClusterRoleBinding,
		machineControllerClusterRoleBinding,
		nodeBootstrapperClusterRoleBinding,
	}

	crbClient := rbacClient.ClusterRoleBindings()
	for _, crbGen := range crbGenerators {
		if err = templates.EnsureClusterRoleBinding(crbClient, crbGen()); err != nil {
			return err
		}
	}

	// Roles
	roleGenerators := []func() *rbacv1.Role{
		machineControllerKubeSystemRole,
		machineControllerKubePublicRole,
		machineControllerEndpointReaderRole,
		machineControllerClusterInfoReaderRole,
	}

	for _, roleGen := range roleGenerators {
		role := roleGen()
		if err = templates.EnsureRole(rbacClient.Roles(role.Namespace), role); err != nil {
			return err
		}
	}

	// RoleBindings
	roleBindingsGenerators := []func() *rbacv1.RoleBinding{
		machineControllerKubeSystemRoleBinding,
		machineControllerKubePublicRoleBinding,
		machineControllerDefaultRoleBinding,
		machineControllerClusterInfoRoleBinding,
	}

	for _, roleBindingGen := range roleBindingsGenerators {
		roleBinding := roleBindingGen()
		if err = templates.EnsureRoleBinding(rbacClient.RoleBindings(roleBinding.Namespace), roleBinding); err != nil {
			return err
		}
	}

	// Secrets
	secret := machineControllerCredentialsSecret(ctx.Cluster)
	err = templates.EnsureSecret(coreClient.Secrets(secret.Namespace), secret)
	if err != nil {
		return err
	}

	// Deployments
	deployment, err := machineControllerDeployment(ctx.Cluster)
	if err != nil {
		return err
	}

	deploymentClient := ctx.Clientset.AppsV1().Deployments(deployment.Namespace)
	err = templates.EnsureDeployment(deploymentClient, deployment)
	if err != nil {
		return err
	}

	// CRDs
	crdGenerators := []func() *apiextensions.CustomResourceDefinition{
		machineControllerMachineCRD,
		machineControllerClusterCRD,
		machineControllerMachineSetCRD,
		machineControllerMachineDeploymentCRD,
	}
	crdClient := ctx.APIExtensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions()

	for _, crdGen := range crdGenerators {
		err = templates.EnsureCRD(crdClient, crdGen())
		if err != nil {
			return err
		}
	}

	return nil
}

// WaitForMachineController waits for machine-controller-webhook to become running
func WaitForMachineController(coreClient corev1types.CoreV1Interface) error {
	return wait.Poll(500*time.Millisecond, time.Minute, func() (bool, error) {
		machineControllerPods, err := coreClient.Pods(WebhookNamespace).List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", MachineControllerAppLabelKey, MachineControllerAppLabelValue),
		})
		if err != nil {
			return false, err
		}
		if machineControllerPods.Items[0].Status.Phase == corev1.PodRunning {
			return true, nil
		}
		return false, nil
	})
}

func machineControllerServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: MachineControllerNamespace,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
	}
}

func machineControllerClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "machine-controller",
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
				Verbs:     []string{"get"},
			},
			// {
			// 	APIGroups:     []string{"apiextensions.k8s.io"},
			// 	Resources:     []string{"customresourcedefinitions"},
			// 	ResourceNames: []string{"machines.machine.k8s.io"},
			// 	Verbs:         []string{"delete"},
			// },
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
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"list", "get", "watch"},
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

func machineControllerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "machine-controller",
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Name:     "machine-controller",
			Kind:     "ClusterRole",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "machine-controller",
				Namespace: MachineControllerNamespace,
			},
		},
	}
}

func nodeBootstrapperClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "machine-controller:kubelet-bootstrap",
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Name:     "system:node-bootstrapper",
			Kind:     "ClusterRole",
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: rbacv1.GroupName,
				Kind:     "Group",
				Name:     "system:bootstrappers:machine-controller:default-node-token",
			},
		},
	}
}

func nodeSignerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "machine-controller:node-signer",
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "system:certificates.k8s.io:certificatesigningrequests:nodeclient",
			Kind:     "ClusterRole",
			APIGroup: rbacv1.GroupName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "Group",
				Name:     "system:bootstrappers:machine-controller:default-node-token",
				APIGroup: rbacv1.GroupName,
			},
		},
	}
}

func machineControllerKubeSystemRole() *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: MachineControllerNamespace,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		Rules: []rbacv1.PolicyRule{
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

func machineControllerKubePublicRole() *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: metav1.NamespacePublic,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		Rules: []rbacv1.PolicyRule{
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

func machineControllerEndpointReaderRole() *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		Rules: []rbacv1.PolicyRule{
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

func machineControllerClusterInfoReaderRole() *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-info",
			Namespace: metav1.NamespacePublic,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{""},
				ResourceNames: []string{"cluster-info"},
				Resources:     []string{"configmaps"},
				Verbs:         []string{"get"},
			},
		},
	}
}

func machineControllerKubeSystemRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: MachineControllerNamespace,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "machine-controller",
			Kind:     "Role",
			APIGroup: rbacv1.GroupName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "machine-controller",
				Namespace: MachineControllerNamespace,
			},
		},
	}
}

func machineControllerKubePublicRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: metav1.NamespacePublic,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "machine-controller",
			Kind:     "Role",
			APIGroup: rbacv1.GroupName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "machine-controller",
				Namespace: MachineControllerNamespace,
			},
		},
	}
}

func machineControllerDefaultRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "machine-controller",
			Kind:     "Role",
			APIGroup: rbacv1.GroupName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "machine-controller",
				Namespace: MachineControllerNamespace,
			},
		},
	}
}

func machineControllerClusterInfoRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-info",
			Namespace: metav1.NamespacePublic,
			Labels: map[string]string{
				MachineControllerAppLabelKey: MachineControllerAppLabelValue,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "cluster-info",
			Kind:     "Role",
			APIGroup: rbacv1.GroupName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "machine-controller",
				Namespace: MachineControllerNamespace,
			},
		},
	}
}

// NB: CRDs are defined as YAML literals because the Go structures
// from k8s.io would always create a "status" field, which breaks the
// validation and prevents them from being applied to the cluster.

func machineControllerMachineCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "machines.cluster.k8s.io",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "cluster.k8s.io",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:   "machines",
				Singular: "machine",
				Kind:     "Machine",
				ListKind: "MachineList",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

func machineControllerClusterCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "clusters.cluster.k8s.io",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "cluster.k8s.io",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:   "clusters",
				Singular: "cluster",
				Kind:     "Cluster",
				ListKind: "ClusterList",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

func machineControllerMachineSetCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "machinesets.cluster.k8s.io",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "cluster.k8s.io",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:   "machinesets",
				Singular: "machineset",
				Kind:     "MachineSet",
				ListKind: "MachineSetList",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

func machineControllerMachineDeploymentCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "machinedeployments.cluster.k8s.io",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "cluster.k8s.io",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:   "machinedeployments",
				Singular: "machinedeployment",
				Kind:     "MachineDeployment",
				ListKind: "MachineDeploymentList",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

func machineControllerDeployment(cluster *config.Cluster) (*appsv1.Deployment, error) {
	var replicas int32 = 1

	clusterDNS, err := clusterDNSIP(cluster)
	if err != nil {
		return nil, err
	}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-controller",
			Namespace: MachineControllerNamespace,
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
					Labels: map[string]string{
						MachineControllerAppLabelKey: MachineControllerAppLabelValue,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "machine-controller",
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
								"-cluster-dns", clusterDNS.String(),
							},
							Env:                      getEnvVarCredentials(cluster),
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
	}, nil
}

func machineControllerCredentialsSecret(cluster *config.Cluster) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      MachineControllerCredentialsSecretName,
			Namespace: MachineControllerNamespace,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: cluster.Provider.Credentials,
	}
}

func getEnvVarCredentials(cluster *config.Cluster) []corev1.EnvVar {
	env := make([]corev1.EnvVar, 0)

	for k := range cluster.Provider.Credentials {
		env = append(env, corev1.EnvVar{
			Name: k,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: MachineControllerCredentialsSecretName,
					},
					Key: k,
				},
			},
		})
	}

	return env
}

// clusterDNSIP returns the IP address of ClusterDNS Service,
// which is 10th IP of the Services CIDR.
func clusterDNSIP(cluster *config.Cluster) (*net.IP, error) {
	// Get the Services CIDR
	_, svcSubnetCIDR, err := net.ParseCIDR(cluster.Network.ServiceSubnet())
	if err != nil {
		return nil, err
	}

	// Select the 10th IP in Services CIDR range as ClusterDNSIP
	clusterDNS, err := ipallocator.GetIndexedIP(svcSubnetCIDR, 10)
	if err != nil {
		return nil, err
	}

	return &clusterDNS, nil
}
