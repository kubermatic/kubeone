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

package machinecontroller

import (
	"context"
	"time"

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/clientutil"
	"github.com/kubermatic/kubeone/pkg/credentials"
	"github.com/kubermatic/kubeone/pkg/kubeconfig"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/templates/nodelocaldns"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// MachineController related constants
const (
	MachineControllerNamespace     = metav1.NamespaceSystem
	MachineControllerAppLabelKey   = "app"
	MachineControllerAppLabelValue = "machine-controller"
	MachineControllerTag           = "v1.5.8"
)

// Deploy deploys MachineController deployment with RBAC on the cluster
func Deploy(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	ctx := context.Background()

	deployment, err := machineControllerDeployment(s.Cluster, s.CredentialsFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to generate machine-controller deployment")
	}

	k8sobject := []runtime.Object{
		machineControllerServiceAccount(),
		machineControllerClusterRole(),
		nodeSignerClusterRoleBinding(),
		machineControllerClusterRoleBinding(),
		nodeBootstrapperClusterRoleBinding(),
		machineControllerKubeSystemRole(),
		machineControllerKubePublicRole(),
		machineControllerEndpointReaderRole(),
		machineControllerClusterInfoReaderRole(),
		machineControllerKubeSystemRoleBinding(),
		machineControllerKubePublicRoleBinding(),
		machineControllerDefaultRoleBinding(),
		machineControllerClusterInfoRoleBinding(),
		machineControllerMachineCRD(),
		machineControllerClusterCRD(),
		machineControllerMachineSetCRD(),
		machineControllerMachineDeploymentCRD(),
		deployment,
	}

	for _, obj := range k8sobject {
		if err = clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj); err != nil {
			return errors.Wrapf(err, "failed to ensure machine-controller %T", obj)
		}
	}

	// HACK: re-init dynamic client in order to re-init RestMapper, to drop caches
	err = kubeconfig.HackIssue321InitDynamicClient(s)
	return errors.Wrap(err, "failed to re-init dynamic client")
}

// WaitForMachineController waits for machine-controller-webhook to become running
// func WaitForMachineController(corev1Client corev1types.CoreV1Interface) error {
func WaitForMachineController(client dynclient.Client) error {
	listOpts := dynclient.ListOptions{
		Namespace: WebhookNamespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			MachineControllerAppLabelKey: MachineControllerAppLabelValue,
		}),
	}

	return wait.Poll(5*time.Second, 3*time.Minute, func() (bool, error) {
		machineControllerPods := corev1.PodList{}
		err := client.List(context.Background(), &machineControllerPods, &listOpts)
		if err != nil {
			return false, errors.Wrap(err, "failed to list machine-controller pod")
		}

		if len(machineControllerPods.Items) == 0 {
			return false, nil
		}

		mcpod := machineControllerPods.Items[0]

		if mcpod.Status.Phase == corev1.PodRunning {
			for _, podcond := range mcpod.Status.Conditions {
				if podcond.Type == corev1.PodReady && podcond.Status == corev1.ConditionTrue {
					return true, nil
				}
			}
		}

		return false, nil
	})
}

func machineControllerServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
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
				Verbs:     []string{"create", "get", "list", "watch"},
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
				Resources: []string{"persistentvolumes", "secrets", "configmaps"},
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
				Resources: []string{
					"clusters",
					"clusters/status",
					"machinedeployments",
					"machinedeployments/status",
					"machines",
					"machinesets",
					"machinesets/status",
				},
				Verbs: []string{"*"},
			},
		},
	}
}

func machineControllerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
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
					"list",
					"update",
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
		ObjectMeta: metav1.ObjectMeta{
			Name: "machines.cluster.k8s.io",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group: "cluster.k8s.io",
			Scope: apiextensions.NamespaceScoped,
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:     "machines",
				Singular:   "machine",
				Kind:       "Machine",
				ListKind:   "MachineList",
				ShortNames: []string{"ma"},
			},
			AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
				{
					Name:     "Provider",
					Type:     "string",
					JSONPath: ".spec.providerSpec.value.cloudProvider",
				},
				{
					Name:     "OS",
					Type:     "string",
					JSONPath: ".spec.providerSpec.value.operatingSystem",
				},
				{
					Name:     "MachineSet",
					Type:     "string",
					JSONPath: ".metadata.ownerReferences[0].name",
					Priority: 1,
				},
				{
					Name:     "Node",
					Type:     "string",
					JSONPath: ".status.nodeRef.name",
					Priority: 1,
				},

				{
					Name:     "Address",
					Type:     "string",
					JSONPath: ".status.addresses[0].address",
				},
				{
					Name:     "Kubelet",
					Type:     "string",
					JSONPath: ".spec.versions.kubelet",
				},
				{
					Name:     "Age",
					Type:     "date",
					JSONPath: ".metadata.creationTimestamp",
				},
				{
					Name:     "Deleted",
					Type:     "date",
					JSONPath: ".metadata.deletionTimestamp",
					Priority: 1,
				},
			},
		},
	}
}

func machineControllerClusterCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "clusters.cluster.k8s.io",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group: "cluster.k8s.io",
			Scope: apiextensions.NamespaceScoped,
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:     "clusters",
				Singular:   "cluster",
				Kind:       "Cluster",
				ListKind:   "ClusterList",
				ShortNames: []string{"cl"},
			},
			Subresources: &apiextensions.CustomResourceSubresources{
				Status: &apiextensions.CustomResourceSubresourceStatus{},
			},
		},
	}
}

func machineControllerMachineSetCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "machinesets.cluster.k8s.io",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group: "cluster.k8s.io",
			Scope: apiextensions.NamespaceScoped,
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:     "machinesets",
				Singular:   "machineset",
				Kind:       "MachineSet",
				ListKind:   "MachineSetList",
				ShortNames: []string{"ms"},
			},
			Subresources: &apiextensions.CustomResourceSubresources{
				Status: &apiextensions.CustomResourceSubresourceStatus{},
				Scale: &apiextensions.CustomResourceSubresourceScale{
					SpecReplicasPath:   ".spec.replicas",
					StatusReplicasPath: ".status.replicas",
				},
			},
			AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
				{
					Name:     "Replicas",
					Type:     "integer",
					JSONPath: ".spec.replicas",
				},
				{
					Name:     "Available-Replicas",
					Type:     "integer",
					JSONPath: ".status.availableReplicas",
				},
				{
					Name:     "Provider",
					Type:     "string",
					JSONPath: ".spec.template.spec.providerSpec.value.cloudProvider",
				},
				{
					Name:     "OS",
					Type:     "string",
					JSONPath: ".spec.template.spec.providerSpec.value.operatingSystem",
				},
				{
					Name:     "MachineDeployment",
					Type:     "string",
					JSONPath: ".metadata.ownerReferences[0].name",
					Priority: 1,
				},
				{
					Name:     "Kubelet",
					Type:     "string",
					JSONPath: ".spec.template.spec.versions.kubelet",
				},
				{
					Name:     "Age",
					Type:     "date",
					JSONPath: ".metadata.creationTimestamp",
				},
				{
					Name:     "Deleted",
					Type:     "date",
					JSONPath: ".metadata.deletionTimestamp",
					Priority: 1,
				},
			},
		},
	}
}

func machineControllerMachineDeploymentCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "machinedeployments.cluster.k8s.io",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group: "cluster.k8s.io",
			Scope: apiextensions.NamespaceScoped,
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:     "machinedeployments",
				Singular:   "machinedeployment",
				Kind:       "MachineDeployment",
				ListKind:   "MachineDeploymentList",
				ShortNames: []string{"md"},
			},
			Subresources: &apiextensions.CustomResourceSubresources{
				Status: &apiextensions.CustomResourceSubresourceStatus{},
				Scale: &apiextensions.CustomResourceSubresourceScale{
					SpecReplicasPath:   ".spec.replicas",
					StatusReplicasPath: ".status.replicas",
				},
			},
			AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
				{
					Name:     "Replicas",
					Type:     "integer",
					JSONPath: ".spec.replicas",
				},
				{
					Name:     "Available-Replicas",
					Type:     "integer",
					JSONPath: ".status.availableReplicas",
				},
				{
					Name:     "Provider",
					Type:     "string",
					JSONPath: ".spec.template.spec.providerSpec.value.cloudProvider",
				},
				{
					Name:     "OS",
					Type:     "string",
					JSONPath: ".spec.template.spec.providerSpec.value.operatingSystem",
				},
				{
					Name:     "Kubelet",
					Type:     "string",
					JSONPath: ".spec.template.spec.versions.kubelet",
				},
				{
					Name:     "Age",
					Type:     "date",
					JSONPath: ".metadata.creationTimestamp",
				},
				{
					Name:     "Deleted",
					Type:     "date",
					JSONPath: ".metadata.deletionTimestamp",
					Priority: 1,
				},
			},
		},
	}
}

func machineControllerDeployment(cluster *kubeoneapi.KubeOneCluster, credentialsFilePath string) (*appsv1.Deployment, error) {
	var replicas int32 = 1

	args := []string{
		"-logtostderr",
		"-v", "4",
		"-internal-listen-address", "0.0.0.0:8085",
		"-cluster-dns", nodelocaldns.VirtualIP,
	}

	if cluster.Proxy.HTTP != "" {
		args = append(args, "-node-http-proxy", cluster.Proxy.HTTP)
	}

	if cluster.Proxy.NoProxy != "" {
		args = append(args, "-node-no-proxy", cluster.Proxy.NoProxy)
	}

	if cluster.CloudProvider.External {
		args = append(args, "-external-cloud-provider")
	}

	envVar, err := credentials.EnvVarBindings(cluster.CloudProvider.Name, credentialsFilePath)
	envVar = append(envVar,
		corev1.EnvVar{
			Name:  "HTTPS_PROXY",
			Value: cluster.Proxy.HTTPS,
		},
		corev1.EnvVar{
			Name:  "NO_PROXY",
			Value: cluster.Proxy.NoProxy,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get env var bindings for a secret")
	}

	return &appsv1.Deployment{
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
							Name:                     "machine-controller",
							Image:                    "docker.io/kubermatic/machine-controller:" + MachineControllerTag,
							ImagePullPolicy:          corev1.PullIfNotPresent,
							Command:                  []string{"/usr/local/bin/machine-controller"},
							Args:                     args,
							Env:                      envVar,
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
