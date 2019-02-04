package templates

import (
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
)

// CreateServiceAccounts takes a slice of ServiceAccounts and ensures they exist and are in the desired state
func CreateServiceAccounts(clientset *kubernetes.Clientset, sa []*corev1.ServiceAccount) error {
	var errs []error
	for _, res := range sa {
		if err := ensureServiceAccount(clientset, res); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.NewAggregate(errs)
}

// ensureServiceAccount checks does ServiceAccount already exists and creates it if it doesn't. If it already exists,
// the function compares labels and annotations, and if they're not as expected updates the ServiceAccount.
func ensureServiceAccount(clientset *kubernetes.Clientset, required *corev1.ServiceAccount) error {
	existing, err := clientset.CoreV1().ServiceAccounts(required.Namespace).Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clientset.CoreV1().ServiceAccounts(required.Namespace).Create(required)
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

	_, err = clientset.CoreV1().ServiceAccounts(existing.Namespace).Update(existing)
	return err
}

// CreateClusterRoles takes a slice of RBAC ClusterRoles and ensures they exist and are in the desired state
func CreateClusterRoles(clientset *kubernetes.Clientset, cr []*rbacv1.ClusterRole) error {
	var errs []error
	for _, res := range cr {
		if err := ensureClusterRole(clientset, res); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.NewAggregate(errs)
}

// ensureClusterRole checks does RBAC ClusterRole already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and rules, and if they're not as expected updates the ClusterRole.
func ensureClusterRole(clientset *kubernetes.Clientset, required *rbacv1.ClusterRole) error {
	existing, err := clientset.RbacV1().ClusterRoles().Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clientset.RbacV1().ClusterRoles().Create(required)
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

	_, err = clientset.RbacV1().ClusterRoles().Update(existing)
	return err
}

// CreateClusterRoleBindings takes a slice of RBAC ClusterRoleBindings and ensures they exist and are in the desired state
func CreateClusterRoleBindings(clientset *kubernetes.Clientset, crb []*rbacv1.ClusterRoleBinding) error {
	var errs []error
	for _, res := range crb {
		if err := ensureClusterRoleBinding(clientset, res); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.NewAggregate(errs)
}

// ensureClusterRoleBinding checks does RBAC ClusterRoleBinding already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, role references, and subjects, and if they're not as expected updates the ClusterRoleBinding.
func ensureClusterRoleBinding(clientset *kubernetes.Clientset, required *rbacv1.ClusterRoleBinding) error {
	existing, err := clientset.RbacV1().ClusterRoleBindings().Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clientset.RbacV1().ClusterRoleBindings().Create(required)
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

	_, err = clientset.RbacV1().ClusterRoleBindings().Update(existing)
	return err
}

// CreateRoles takes a slice of RBAC Roles and ensures they exist and are in the desired state
func CreateRoles(clientset *kubernetes.Clientset, roles []*rbacv1.Role) error {
	var errs []error
	for _, res := range roles {
		if err := ensureRole(clientset, res); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.NewAggregate(errs)
}

// ensureRole checks does RBAC Role already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and rules, and if they're not as expected updates the Role.
func ensureRole(clientset *kubernetes.Clientset, required *rbacv1.Role) error {
	existing, err := clientset.RbacV1().Roles(required.Namespace).Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clientset.RbacV1().Roles(required.Namespace).Create(required)
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

	_, err = clientset.RbacV1().Roles(existing.Namespace).Update(existing)
	return err
}

// CreateRoleBindings takes a slice of RBAC RoleBindings and ensures they exist and are in the desired state
func CreateRoleBindings(clientset kubernetes.Interface, rb []*rbacv1.RoleBinding) error {
	var errs []error
	for _, res := range rb {
		if err := ensureRoleBinding(clientset, res); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.NewAggregate(errs)
}

// ensureRoleBinding checks does RBAC RoleBinding already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, role references, and subjects, and if they're not as expected updates the RoleBinding.
func ensureRoleBinding(clientset kubernetes.Interface, required *rbacv1.RoleBinding) error {
	existing, err := clientset.RbacV1().RoleBindings(required.Namespace).Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clientset.RbacV1().RoleBindings(required.Namespace).Create(required)
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

	_, err = clientset.RbacV1().RoleBindings(existing.Namespace).Update(existing)
	return err
}

// CreateSecrets takes a slice of Secrets and ensures they exist and are in the desired state
func CreateSecrets(clientset kubernetes.Interface, s []*corev1.Secret) error {
	var errs []error
	for _, res := range s {
		if err := ensureSecret(clientset, res); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.NewAggregate(errs)
}

// ensureSecret checks does Secret already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and data, and if they're not as expected updates the Secret.
func ensureSecret(clientset kubernetes.Interface, required *corev1.Secret) error {
	existing, err := clientset.CoreV1().Secrets(required.Namespace).Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clientset.CoreV1().Secrets(required.Namespace).Create(required)
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

	_, err = clientset.CoreV1().Secrets(existing.Namespace).Update(existing)
	return err
}

// CreateDeployments takes a slice of Deployments and ensures they exist and are in the desired state
func CreateDeployments(clientset kubernetes.Interface, deploy []*appsv1.Deployment) error {
	var errs []error
	for _, res := range deploy {
		if err := ensureDeployment(clientset, res); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.NewAggregate(errs)
}

// ensureDeployment checks does Deployment already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the Deployment.
func ensureDeployment(clientset kubernetes.Interface, required *appsv1.Deployment) error {
	existing, err := clientset.AppsV1().Deployments(required.Namespace).Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clientset.AppsV1().Deployments(required.Namespace).Create(required)
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

	_, err = clientset.AppsV1().Deployments(existing.Namespace).Update(existing)
	return err
}

// CreateServices takes a slice of Services and ensures they exist and are in the desired state
func CreateServices(clientset kubernetes.Interface, svc []*corev1.Service) error {
	var errs []error
	for _, res := range svc {
		if err := ensureService(clientset, res); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.NewAggregate(errs)
}

// ensureService checks does Service already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the Service.
func ensureService(clientset kubernetes.Interface, required *corev1.Service) error {
	existing, err := clientset.CoreV1().Services(required.Namespace).Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clientset.CoreV1().Services(required.Namespace).Create(required)
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

	_, err = clientset.CoreV1().Services(existing.Namespace).Update(existing)
	return err
}

// CreateCRDs takes a slice of CRDs and ensures they exist and are in the desired state
func CreateCRDs(clientset apiextensionsclientset.Interface, crds []*apiextensionsv1beta1.CustomResourceDefinition) error {
	var errs []error
	for _, res := range crds {
		if err := ensureCRD(clientset, res); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.NewAggregate(errs)
}

// ensureCRD checks does CRD already exists and creates it if it doesn't. If it already exists,
// the function compares labels, annotations, and spec, and if they're not as expected updates the CRD.
func ensureCRD(clientset apiextensionsclientset.Interface, required *apiextensionsv1beta1.CustomResourceDefinition) error {
	existing, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(required)
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

	_, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Update(existing)
	return err
}

// CreateMutatingWebhookConfigurations takes a slice of MutatingWebhookConfigurations and ensures they exist and are in the desired state
func CreateMutatingWebhookConfigurations(clientset *kubernetes.Clientset, crds []*admissionregistrationv1beta1.MutatingWebhookConfiguration) error {
	var errs []error
	for _, res := range crds {
		if err := ensureMutatingWebhookConfiguration(clientset, res); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.NewAggregate(errs)
}

// ensureMutatingWebhookConfiguration checks does MutatingWebhookConfiguration already exists and creates it if it doesn't.
// If it already exists, the function compares labels, annotations, and spec, and if they're not as expected updates
// the MutatingWebhookConfiguration.
func ensureMutatingWebhookConfiguration(clientset *kubernetes.Clientset, required *admissionregistrationv1beta1.MutatingWebhookConfiguration) error {
	existing, err := clientset.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = clientset.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(required)
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

	_, err = clientset.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Update(existing)
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
