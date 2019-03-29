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
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/pkg/registry/core/service/ipallocator"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// MachineController related constants
const (
	MachineControllerNamespace             = metav1.NamespaceSystem
	MachineControllerAppLabelKey           = "app"
	MachineControllerAppLabelValue         = "machine-controller"
	MachineControllerTag                   = "v1.1.2"
	MachineControllerCredentialsSecretName = "machine-controller-credentials"
)

// Deploy deploys MachineController deployment with RBAC on the cluster
func Deploy(ctx *util.Context) error {
	if ctx.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	bgCtx := context.Background()

	// ServiceAccounts
	if err := simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, machineControllerServiceAccount()); err != nil {
		return errors.Wrap(err, "failed to ensure machine-controller service account")
	}

	// ClusterRoles
	if err := simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, machineControllerClusterRole()); err != nil {
		return errors.Wrap(err, "failed to ensure machine-controller cluster role")
	}

	// ClusterRoleBindings
	crbGenerators := []func() *rbacv1.ClusterRoleBinding{
		nodeSignerClusterRoleBinding,
		machineControllerClusterRoleBinding,
		nodeBootstrapperClusterRoleBinding,
	}

	for _, crbGen := range crbGenerators {
		if err := simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, crbGen()); err != nil {
			return errors.Wrap(err, "failed to ensure machine-controller cluster-role binding")
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
		if err := simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, roleGen()); err != nil {
			return errors.Wrap(err, "failed to ensure machine-controller role")
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
		if err := simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, roleBindingGen()); err != nil {
			return errors.Wrap(err, "failed to ensure machine-controller role binding")
		}
	}

	// Secrets
	secret := machineControllerCredentialsSecret(ctx.Cluster)
	if err := simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, secret); err != nil {
		return errors.Wrap(err, "failed to ensure machine-controller credentials secret")
	}

	// Deployments
	deployment, err := machineControllerDeployment(ctx.Cluster)
	if err != nil {
		return errors.Wrap(err, "failed to generate machine-controller deployment")
	}

	if err = simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, deployment); err != nil {
		return errors.Wrap(err, "failed to ensure machine-controller deployment")
	}

	// CRDs
	crdGenerators := []func() *apiextensions.CustomResourceDefinition{
		machineControllerMachineCRD,
		machineControllerClusterCRD,
		machineControllerMachineSetCRD,
		machineControllerMachineDeploymentCRD,
	}

	for _, crdGen := range crdGenerators {
		if err = simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, crdGen()); err != nil {
			return errors.Wrap(err, "failed to ensure machine-controller CRDs")
		}
	}

	// HACK: re-init dynamic client in order to re-init RestMapper, to drop caches
	err = util.HackIssue321InitDynamicClient(ctx)
	return errors.Wrap(err, "failed to re-init dynamic client")
}

// WaitForMachineController waits for machine-controller-webhook to become running
// func WaitForMachineController(corev1Client corev1types.CoreV1Interface) error {
func WaitForMachineController(client dynclient.Client) error {
	listOpts := dynclient.ListOptions{Namespace: WebhookNamespace}
	err := listOpts.SetLabelSelector(fmt.Sprintf("%s=%s", MachineControllerAppLabelKey, MachineControllerAppLabelValue))
	if err != nil {
		return errors.Wrap(err, "failed to parse machine-controller labels")
	}

	return wait.Poll(5*time.Second, 3*time.Minute, func() (bool, error) {
		machineControllerPods := corev1.PodList{}
		err = client.List(context.Background(), &listOpts, &machineControllerPods)
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
				Plural:   "machines",
				Singular: "machine",
				Kind:     "Machine",
				ListKind: "MachineList",
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
					Name:     "Address",
					Type:     "string",
					JSONPath: ".status.addresses[0].address",
				},
				{
					Name:     "Age",
					Type:     "date",
					JSONPath: ".metadata.creationTimestamp",
				},
			},
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
				Plural:   "clusters",
				Singular: "cluster",
				Kind:     "Cluster",
				ListKind: "ClusterList",
			},
			Subresources: &apiextensions.CustomResourceSubresources{
				Status: &apiextensions.CustomResourceSubresourceStatus{},
			},
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
				Plural:   "machinesets",
				Singular: "machineset",
				Kind:     "MachineSet",
				ListKind: "MachineSetList",
			},
			Subresources: &apiextensions.CustomResourceSubresources{
				Status: &apiextensions.CustomResourceSubresourceStatus{},
			},
			AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
				{
					Name:     "Replicas",
					Type:     "integer",
					JSONPath: ".spec.replicas",
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
					Name:     "Age",
					Type:     "date",
					JSONPath: ".metadata.creationTimestamp",
				},
			},
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
				Plural:   "machinedeployments",
				Singular: "machinedeployment",
				Kind:     "MachineDeployment",
				ListKind: "MachineDeploymentList",
			},
			Subresources: &apiextensions.CustomResourceSubresources{
				Status: &apiextensions.CustomResourceSubresourceStatus{},
			},
			AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
				{
					Name:     "Replicas",
					Type:     "integer",
					JSONPath: ".spec.replicas",
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
					Name:     "Age",
					Type:     "date",
					JSONPath: ".metadata.creationTimestamp",
				},
			},
		},
	}
}

func machineControllerDeployment(cluster *config.Cluster) (*appsv1.Deployment, error) {
	var replicas int32 = 1

	clusterDNS, err := clusterDNSIP(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get clusterDNS IP")
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
		StringData: cluster.MachineController.Credentials,
	}
}

func getEnvVarCredentials(cluster *config.Cluster) []corev1.EnvVar {
	env := make([]corev1.EnvVar, 0)

	for k := range cluster.MachineController.Credentials {
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
		return nil, errors.Wrap(err, "failed to parse network.service_subnet")
	}

	// Select the 10th IP in Services CIDR range as ClusterDNSIP
	clusterDNS, err := ipallocator.GetIndexedIP(svcSubnetCIDR, 10)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get IP from service subnet")
	}

	return &clusterDNS, nil
}
