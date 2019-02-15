package templates

import (
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsv1beta1types "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	admissionregistrationv1beta1types "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"
	appsv1types "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1types "k8s.io/client-go/kubernetes/typed/core/v1"
	rbacv1types "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

// EnsureNamespace checks does Namespace already exists and creates it if it doesn't. If it already exists,
// the function compares labels and annotations, and if they're not as expected updates the Namespace.
func EnsureNamespace(namespaceClient corev1types.NamespaceInterface, required *corev1.Namespace) error {
	existing, err := namespaceClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = namespaceClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if !modified {
		return nil
	}

	_, err = namespaceClient.Update(existing)
	return err
}

// EnsureServiceAccount checks does ServiceAccount already exists and creates it if it doesn't. If it already exists,
// the function compares labels and annotations, and if they're not as expected updates the ServiceAccount.
func EnsureServiceAccount(serviceAccoutnClient corev1types.ServiceAccountInterface, required *corev1.ServiceAccount) error {
	existing, err := serviceAccoutnClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = serviceAccoutnClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if !modified {
		return nil
	}

	_, err = serviceAccoutnClient.Update(existing)
	return err
}

// EnsureClusterRole checks does RBAC ClusterRole already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and rules, and if they're not as expected updates the ClusterRole.
func EnsureClusterRole(clusterRoleClient rbacv1types.ClusterRoleInterface, required *rbacv1.ClusterRole) error {
	existing, err := clusterRoleClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clusterRoleClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Rules, existing.Rules) && !modified {
		return nil
	}

	_, err = clusterRoleClient.Update(existing)
	return err
}

// EnsureClusterRoleBinding checks does RBAC ClusterRoleBinding already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, role references, and subjects, and if they're not as expected updates the ClusterRoleBinding.
func EnsureClusterRoleBinding(clusterRoleBindingClient rbacv1types.ClusterRoleBindingInterface, required *rbacv1.ClusterRoleBinding) error {
	existing, err := clusterRoleBindingClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clusterRoleBindingClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.RoleRef, existing.RoleRef) && equality.Semantic.DeepEqual(required.Subjects, existing.Subjects) && !modified {
		return nil
	}

	_, err = clusterRoleBindingClient.Update(existing)
	return err
}

// EnsureRole checks does RBAC Role already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and rules, and if they're not as expected updates the Role.
func EnsureRole(roleClient rbacv1types.RoleInterface, required *rbacv1.Role) error {
	existing, err := roleClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = roleClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Rules, existing.Rules) && !modified {
		return nil
	}

	_, err = roleClient.Update(existing)
	return err
}

// EnsureRoleBinding checks does RBAC RoleBinding already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, role references, and subjects, and if they're not as expected updates the RoleBinding.
func EnsureRoleBinding(roleBindingClient rbacv1types.RoleBindingInterface, required *rbacv1.RoleBinding) error {
	existing, err := roleBindingClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = roleBindingClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.RoleRef, existing.RoleRef) && equality.Semantic.DeepEqual(required.Subjects, existing.Subjects) && !modified {
		return nil
	}

	_, err = roleBindingClient.Update(existing)
	return err
}

// EnsureConfigMap checks does ConfigMap already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and data, and if they're not as expected updates the ConfigMap.
func EnsureConfigMap(configMapClient corev1types.ConfigMapInterface, required *corev1.ConfigMap) error {
	existing, err := configMapClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = configMapClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Data, existing.Data) && !modified {
		return nil
	}

	_, err = configMapClient.Update(existing)
	return err
}

// EnsureSecret checks does Secret already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and data, and if they're not as expected updates the Secret.
func EnsureSecret(secretClient corev1types.SecretInterface, required *corev1.Secret) error {
	existing, err := secretClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = secretClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Data, existing.Data) && !modified {
		return nil
	}

	_, err = secretClient.Update(existing)
	return err
}

// EnsureDeployment checks does Deployment already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the Deployment.
func EnsureDeployment(deploymentClient appsv1types.DeploymentInterface, required *appsv1.Deployment) error {
	existing, err := deploymentClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = deploymentClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Spec, existing.Spec) && !modified {
		return nil
	}

	_, err = deploymentClient.Update(existing)
	return err
}

// EnsureDaemonSet checks does DaemonSet already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the DaemonSet.
func EnsureDaemonSet(daemonSetClient appsv1types.DaemonSetInterface, required *appsv1.DaemonSet) error {
	existing, err := daemonSetClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = daemonSetClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Spec, existing.Spec) && !modified {
		return nil
	}

	_, err = daemonSetClient.Update(existing)
	return err
}

// EnsureService checks does Service already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the Service.
func EnsureService(serviceClient corev1types.ServiceInterface, required *corev1.Service) error {
	existing, err := serviceClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = serviceClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Spec, existing.Spec) && !modified {
		return nil
	}

	_, err = serviceClient.Update(existing)
	return err
}

// EnsureCRD checks does CRD already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the CRD.
func EnsureCRD(customResourceDefinitionClient apiextensionsv1beta1types.CustomResourceDefinitionInterface, required *apiextensionsv1beta1.CustomResourceDefinition) error {
	existing, err := customResourceDefinitionClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = customResourceDefinitionClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Spec, existing.Spec) && !modified {
		return nil
	}

	_, err = customResourceDefinitionClient.Update(existing)
	return err
}

// EnsureMutatingWebhookConfiguration checks does MutatingWebhookConfiguration already exists and creates it if it doesn't.
// If it already exists, the function compares labels, annotations, and spec, and if they're not as expected updates
// the MutatingWebhookConfiguration.
func EnsureMutatingWebhookConfiguration(mutatingWebhookConfigurationClient admissionregistrationv1beta1types.MutatingWebhookConfigurationInterface, required *admissionregistrationv1beta1.MutatingWebhookConfiguration) error {
	existing, err := mutatingWebhookConfigurationClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = mutatingWebhookConfigurationClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Webhooks, existing.Webhooks) && !modified {
		return nil
	}

	_, err = mutatingWebhookConfigurationClient.Update(existing)
	return err
}
