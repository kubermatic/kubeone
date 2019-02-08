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
func EnsureNamespace(namespaceInterface corev1types.NamespaceInterface, required *corev1.Namespace) error {
	existing, err := namespaceInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = namespaceInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if !modified {
		return nil
	}

	_, err = namespaceInterface.Update(existing)
	return err
}

// EnsureServiceAccount checks does ServiceAccount already exists and creates it if it doesn't. If it already exists,
// the function compares labels and annotations, and if they're not as expected updates the ServiceAccount.
func EnsureServiceAccount(serviceAccountInterface corev1types.ServiceAccountInterface, required *corev1.ServiceAccount) error {
	existing, err := serviceAccountInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = serviceAccountInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if !modified {
		return nil
	}

	_, err = serviceAccountInterface.Update(existing)
	return err
}

// EnsureClusterRole checks does RBAC ClusterRole already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and rules, and if they're not as expected updates the ClusterRole.
func EnsureClusterRole(clusterRoleInterface rbacv1types.ClusterRoleInterface, required *rbacv1.ClusterRole) error {
	existing, err := clusterRoleInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clusterRoleInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Rules, existing.Rules) && !modified {
		return nil
	}

	_, err = clusterRoleInterface.Update(existing)
	return err
}

// EnsureClusterRoleBinding checks does RBAC ClusterRoleBinding already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, role references, and subjects, and if they're not as expected updates the ClusterRoleBinding.
func EnsureClusterRoleBinding(clusterRoleBindingInterface rbacv1types.ClusterRoleBindingInterface, required *rbacv1.ClusterRoleBinding) error {
	existing, err := clusterRoleBindingInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clusterRoleBindingInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.RoleRef, existing.RoleRef) && equality.Semantic.DeepEqual(required.Subjects, existing.Subjects) && !modified {
		return nil
	}

	_, err = clusterRoleBindingInterface.Update(existing)
	return err
}

// EnsureRole checks does RBAC Role already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and rules, and if they're not as expected updates the Role.
func EnsureRole(roleInterface rbacv1types.RoleInterface, required *rbacv1.Role) error {
	existing, err := roleInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = roleInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Rules, existing.Rules) && !modified {
		return nil
	}

	_, err = roleInterface.Update(existing)
	return err
}

// EnsureRoleBinding checks does RBAC RoleBinding already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, role references, and subjects, and if they're not as expected updates the RoleBinding.
func EnsureRoleBinding(roleBindingInterface rbacv1types.RoleBindingInterface, required *rbacv1.RoleBinding) error {
	existing, err := roleBindingInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = roleBindingInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.RoleRef, existing.RoleRef) && equality.Semantic.DeepEqual(required.Subjects, existing.Subjects) && !modified {
		return nil
	}

	_, err = roleBindingInterface.Update(existing)
	return err
}

// EnsureSecret checks does Secret already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and data, and if they're not as expected updates the Secret.
func EnsureSecret(secretInterface corev1types.SecretInterface, required *corev1.Secret) error {
	existing, err := secretInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = secretInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Data, existing.Data) && !modified {
		return nil
	}

	_, err = secretInterface.Update(existing)
	return err
}

// EnsureDeployment checks does Deployment already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the Deployment.
func EnsureDeployment(deploymentInterface appsv1types.DeploymentInterface, required *appsv1.Deployment) error {
	existing, err := deploymentInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = deploymentInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Spec, existing.Spec) && !modified {
		return nil
	}

	_, err = deploymentInterface.Update(existing)
	return err
}

// EnsureDaemonSet checks does DaemonSet already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the DaemonSet.
func EnsureDaemonSet(daemonSetInterface appsv1types.DaemonSetInterface, required *appsv1.DaemonSet) error {
	existing, err := daemonSetInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = daemonSetInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Spec, existing.Spec) && !modified {
		return nil
	}

	_, err = daemonSetInterface.Update(existing)
	return err
}

// EnsureService checks does Service already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the Service.
func EnsureService(serviceInterface corev1types.ServiceInterface, required *corev1.Service) error {
	existing, err := serviceInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = serviceInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Spec, existing.Spec) && !modified {
		return nil
	}

	_, err = serviceInterface.Update(existing)
	return err
}

// EnsureCRD checks does CRD already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the CRD.
func EnsureCRD(customResourceDefinitionInterface apiextensionsv1beta1types.CustomResourceDefinitionInterface, required *apiextensionsv1beta1.CustomResourceDefinition) error {
	existing, err := customResourceDefinitionInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = customResourceDefinitionInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Spec, existing.Spec) && !modified {
		return nil
	}

	_, err = customResourceDefinitionInterface.Update(existing)
	return err
}

// EnsureMutatingWebhookConfiguration checks does MutatingWebhookConfiguration already exists and creates it if it doesn't.
// If it already exists, the function compares labels, annotations, and spec, and if they're not as expected updates
// the MutatingWebhookConfiguration.
func EnsureMutatingWebhookConfiguration(mutatingWebhookConfigurationInterface admissionregistrationv1beta1types.MutatingWebhookConfigurationInterface, required *admissionregistrationv1beta1.MutatingWebhookConfiguration) error {
	existing, err := mutatingWebhookConfigurationInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = mutatingWebhookConfigurationInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	mergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	mergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Webhooks, existing.Webhooks) && !modified {
		return nil
	}

	_, err = mutatingWebhookConfigurationInterface.Update(existing)
	return err
}

func mergeStringMap(modified *bool, destination *map[string]string, required map[string]string) {
	if *destination == nil {
		*destination = map[string]string{}
	}

	for k, v := range required {
		if destinationV, ok := (*destination)[k]; !ok || destinationV != v {
			(*destination)[k] = v
			*modified = true
		}
	}
}
