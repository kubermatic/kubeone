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
func EnsureNamespace(namespacesClient corev1types.NamespaceInterface, required *corev1.Namespace) error {
	existing, err := namespacesClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = namespacesClient.Create(required)
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

	_, err = namespacesClient.Update(existing)
	return err
}

// EnsureServiceAccount checks does ServiceAccount already exists and creates it if it doesn't. If it already exists,
// the function compares labels and annotations, and if they're not as expected updates the ServiceAccount.
func EnsureServiceAccount(serviceAccountsClient corev1types.ServiceAccountInterface, required *corev1.ServiceAccount) error {
	existing, err := serviceAccountsClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = serviceAccountsClient.Create(required)
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

	_, err = serviceAccountsClient.Update(existing)
	return err
}

// EnsureClusterRole checks does RBAC ClusterRole already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and rules, and if they're not as expected updates the ClusterRole.
func EnsureClusterRole(clusterRolesClient rbacv1types.ClusterRoleInterface, required *rbacv1.ClusterRole) error {
	existing, err := clusterRolesClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clusterRolesClient.Create(required)
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

	_, err = clusterRolesClient.Update(existing)
	return err
}

// EnsureClusterRoleBinding checks does RBAC ClusterRoleBinding already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, role references, and subjects, and if they're not as expected updates the ClusterRoleBinding.
func EnsureClusterRoleBinding(clusterRoleBindingsClient rbacv1types.ClusterRoleBindingInterface, required *rbacv1.ClusterRoleBinding) error {
	existing, err := clusterRoleBindingsClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clusterRoleBindingsClient.Create(required)
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

	_, err = clusterRoleBindingsClient.Update(existing)
	return err
}

// EnsureRole checks does RBAC Role already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and rules, and if they're not as expected updates the Role.
func EnsureRole(rolesClient rbacv1types.RoleInterface, required *rbacv1.Role) error {
	existing, err := rolesClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = rolesClient.Create(required)
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

	_, err = rolesClient.Update(existing)
	return err
}

// EnsureRoleBinding checks does RBAC RoleBinding already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, role references, and subjects, and if they're not as expected updates the RoleBinding.
func EnsureRoleBinding(roleBindingsClient rbacv1types.RoleBindingInterface, required *rbacv1.RoleBinding) error {
	existing, err := roleBindingsClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = roleBindingsClient.Create(required)
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

	_, err = roleBindingsClient.Update(existing)
	return err
}

// EnsureConfigMap checks does ConfigMap already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and data, and if they're not as expected updates the ConfigMap.
func EnsureConfigMap(configMapsClient corev1types.ConfigMapInterface, required *corev1.ConfigMap) error {
	existing, err := configMapsClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = configMapsClient.Create(required)
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

	_, err = configMapsClient.Update(existing)
	return err
}

// EnsureSecret checks does Secret already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and data, and if they're not as expected updates the Secret.
func EnsureSecret(secretsClient corev1types.SecretInterface, required *corev1.Secret) error {
	existing, err := secretsClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = secretsClient.Create(required)
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

	_, err = secretsClient.Update(existing)
	return err
}

// EnsureDeployment checks does Deployment already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the Deployment.
func EnsureDeployment(deploymentsClient appsv1types.DeploymentInterface, required *appsv1.Deployment) error {
	existing, err := deploymentsClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = deploymentsClient.Create(required)
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

	_, err = deploymentsClient.Update(existing)
	return err
}

// EnsureDaemonSet checks does DaemonSet already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the DaemonSet.
func EnsureDaemonSet(daemonSetsClient appsv1types.DaemonSetInterface, required *appsv1.DaemonSet) error {
	existing, err := daemonSetsClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = daemonSetsClient.Create(required)
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

	_, err = daemonSetsClient.Update(existing)
	return err
}

// EnsureService checks does Service already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the Service.
func EnsureService(servicesClient corev1types.ServiceInterface, required *corev1.Service) error {
	existing, err := servicesClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = servicesClient.Create(required)
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

	_, err = servicesClient.Update(existing)
	return err
}

// EnsureCRD checks does CRD already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the CRD.
func EnsureCRD(customResourceDefinitionsClient apiextensionsv1beta1types.CustomResourceDefinitionInterface, required *apiextensionsv1beta1.CustomResourceDefinition) error {
	existing, err := customResourceDefinitionsClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = customResourceDefinitionsClient.Create(required)
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

	_, err = customResourceDefinitionsClient.Update(existing)
	return err
}

// EnsureMutatingWebhookConfiguration checks does MutatingWebhookConfiguration already exists and creates it if it doesn't.
// If it already exists, the function compares labels, annotations, and spec, and if they're not as expected updates
// the MutatingWebhookConfiguration.
func EnsureMutatingWebhookConfiguration(mutatingWebhookConfigurationsClient admissionregistrationv1beta1types.MutatingWebhookConfigurationInterface, required *admissionregistrationv1beta1.MutatingWebhookConfiguration) error {
	existing, err := mutatingWebhookConfigurationsClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = mutatingWebhookConfigurationsClient.Create(required)
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

	_, err = mutatingWebhookConfigurationsClient.Update(existing)
	return err
}
