/*
Copyright 2026 The KubeOne Authors.

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

package infra

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/fail"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// LabelClusterName is set on every infra-side object to the KubeOne cluster name.
	LabelClusterName = "kubeone_cluster_name"
	// LabelComponent is set on every infra-side object to the name of the component it belongs to.
	LabelComponent = "kubeone_component"
	// LabelInfraNamespace is set on cluster-scoped infra-side objects to the InfraNamespace they belong to.
	LabelInfraNamespace = "kubeone_infra_namespace"

	// KubeconfigSecretKey is the key in kubeconfig Secrets holding the kubeconfig.
	KubeconfigSecretKey = "kubeconfig"

	tokenWaitInterval = 2 * time.Second
	tokenWaitTimeout  = 2 * time.Minute

	nameHashLength = 8
)

// Component describes a KubeOne component that runs in the infra cluster.
type Component struct {
	// Name of the component, e.g. "machine-controller".
	Name string

	// Rules granted to the component in the InfraNamespace.
	Rules []rbacv1.PolicyRule

	// ClusterRules are granted to the component cluster-wide, for reading
	// cluster-scoped resources such as StorageClasses. Keep them read-only.
	ClusterRules []rbacv1.PolicyRule

	// NeedsKubeconfig is true for components that cannot use in-cluster
	// configuration to access the infra cluster. For them, a ServiceAccount
	// token Secret and a kubeconfig Secret are created.
	NeedsKubeconfig bool
}

// ObjectName returns the name of an infra-side object of the component.
func ObjectName(clusterName, component string) string {
	return clusterName + "-" + component
}

// ServiceName returns the name of an infra-side Service of the component.
// Service names must be DNS-1035 labels, while cluster names are DNS-1123
// subdomains, so names that are too long or otherwise invalid are shortened
// and suffixed with a hash of the full name.
func ServiceName(clusterName, component string) string {
	name := ObjectName(clusterName, component)
	if len(validation.IsDNS1035Label(name)) == 0 {
		return name
	}

	label := strings.ReplaceAll(name, ".", "-")
	if label[0] < 'a' || label[0] > 'z' {
		label = "k" + label
	}

	return shortenWithHash(label, name, validation.DNS1035LabelMaxLength)
}

// ClusterLabelValue returns the value of LabelClusterName for the cluster.
// Label values are limited to 63 characters, so longer cluster names are
// shortened and suffixed with a hash of the full name.
func ClusterLabelValue(clusterName string) string {
	if len(clusterName) <= validation.LabelValueMaxLength {
		return clusterName
	}

	return shortenWithHash(clusterName, clusterName, validation.LabelValueMaxLength)
}

// shortenWithHash truncates s so that s, a dash and a hash of key fit in
// maxLen characters.
func shortenWithHash(s, key string, maxLen int) string {
	sum := sha256.Sum256([]byte(key))
	suffix := "-" + hex.EncodeToString(sum[:])[:nameHashLength]

	s = s[:min(len(s), maxLen-len(suffix))]

	return strings.TrimRight(s, "-.") + suffix
}

// Labels returns the labels set on every infra-side object of the component.
func Labels(clusterName, component string) map[string]string {
	return map[string]string{
		LabelClusterName: ClusterLabelValue(clusterName),
		LabelComponent:   component,
	}
}

func (c Component) objectMeta(clusterName, namespace, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    Labels(clusterName, c.Name),
	}
}

// ServiceAccountName returns the name of the component's infra ServiceAccount.
func (c Component) ServiceAccountName(clusterName string) string {
	return ObjectName(clusterName, c.Name)
}

// TokenSecretName returns the name of the component's infra ServiceAccount token Secret.
func (c Component) TokenSecretName(clusterName string) string {
	return ObjectName(clusterName, c.Name) + "-token"
}

// KubeconfigSecretName returns the name of the Secret holding the component's
// kubeconfig for the infra cluster.
func (c Component) KubeconfigSecretName(clusterName string) string {
	return ObjectName(clusterName, c.Name) + "-kubeconfig"
}

// TenantKubeconfigSecretName returns the name of the Secret holding the
// component's kubeconfig for the tenant cluster.
func (c Component) TenantKubeconfigSecretName(clusterName string) string {
	return ObjectName(clusterName, c.Name) + "-tenant-kubeconfig"
}

func (c Component) ServiceAccount(clusterName, namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: c.objectMeta(clusterName, namespace, c.ServiceAccountName(clusterName)),
	}
}

func (c Component) Role(clusterName, namespace string) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: c.objectMeta(clusterName, namespace, ObjectName(clusterName, c.Name)),
		Rules:      c.Rules,
	}
}

func (c Component) RoleBinding(clusterName, namespace string) *rbacv1.RoleBinding {
	name := ObjectName(clusterName, c.Name)

	return &rbacv1.RoleBinding{
		ObjectMeta: c.objectMeta(clusterName, namespace, name),
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      c.ServiceAccountName(clusterName),
				Namespace: namespace,
			},
		},
	}
}

// ClusterRoleName returns the name of the component's ClusterRole and
// ClusterRoleBinding. It includes the namespace, as cluster names are only
// unique within an InfraNamespace.
func (c Component) ClusterRoleName(clusterName, namespace string) string {
	return "kubeone:" + namespace + ":" + ObjectName(clusterName, c.Name)
}

func (c Component) clusterObjectMeta(clusterName, namespace string) metav1.ObjectMeta {
	labels := Labels(clusterName, c.Name)
	labels[LabelInfraNamespace] = namespace

	return metav1.ObjectMeta{
		Name:   c.ClusterRoleName(clusterName, namespace),
		Labels: labels,
	}
}

func (c Component) ClusterRole(clusterName, namespace string) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: c.clusterObjectMeta(clusterName, namespace),
		Rules:      c.ClusterRules,
	}
}

func (c Component) ClusterRoleBinding(clusterName, namespace string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: c.clusterObjectMeta(clusterName, namespace),
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     c.ClusterRoleName(clusterName, namespace),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      c.ServiceAccountName(clusterName),
				Namespace: namespace,
			},
		},
	}
}

func (c Component) TokenSecret(clusterName, namespace string) *corev1.Secret {
	meta := c.objectMeta(clusterName, namespace, c.TokenSecretName(clusterName))
	meta.Annotations = map[string]string{
		corev1.ServiceAccountNameKey: c.ServiceAccountName(clusterName),
	}

	return &corev1.Secret{
		ObjectMeta: meta,
		Type:       corev1.SecretTypeServiceAccountToken,
	}
}

func (c Component) kubeconfigSecret(clusterName, namespace, name string, kubeconfig []byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: c.objectMeta(clusterName, namespace, name),
		Type:       corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			KubeconfigSecretKey: kubeconfig,
		},
	}
}

// EnsureIdentity reconciles the component's infra ServiceAccount, Role and
// RoleBinding, and the ClusterRole and ClusterRoleBinding if the component has
// ClusterRules. For components that need a kubeconfig, it also reconciles the
// ServiceAccount token Secret and the infra kubeconfig Secret.
func (c Component) EnsureIdentity(ctx context.Context, infraClient ctrlrt.Client, clusterName, namespace string) error {
	objs := []ctrlrt.Object{
		c.ServiceAccount(clusterName, namespace),
		c.Role(clusterName, namespace),
		c.RoleBinding(clusterName, namespace),
	}

	if len(c.ClusterRules) > 0 {
		objs = append(objs, c.ClusterRole(clusterName, namespace), c.ClusterRoleBinding(clusterName, namespace))
	}

	for _, obj := range objs {
		if err := clientutil.CreateOrUpdate(ctx, infraClient, obj); err != nil {
			return err
		}
	}

	if !c.NeedsKubeconfig {
		return nil
	}

	if err := clientutil.CreateOrUpdate(ctx, infraClient, c.TokenSecret(clusterName, namespace)); err != nil {
		return err
	}

	token, err := WaitServiceAccountToken(ctx, infraClient, ctrlrt.ObjectKey{Namespace: namespace, Name: c.TokenSecretName(clusterName)})
	if err != nil {
		return err
	}

	kubeconfig, err := InClusterKubeconfig(clusterName, namespace, token)
	if err != nil {
		return err
	}

	return clientutil.CreateOrReplace(ctx, infraClient, c.kubeconfigSecret(clusterName, namespace, c.KubeconfigSecretName(clusterName), kubeconfig))
}

// EnsureTenantKubeconfig builds a kubeconfig for the tenant cluster from the
// tenant ServiceAccount token Secret tenantTokenKey, and stores it in the
// InfraNamespace.
func (c Component) EnsureTenantKubeconfig(
	ctx context.Context,
	tenantClient, infraClient ctrlrt.Client,
	clusterName, namespace, server string,
	tenantTokenKey ctrlrt.ObjectKey,
) error {
	token, err := WaitServiceAccountToken(ctx, tenantClient, tenantTokenKey)
	if err != nil {
		return err
	}

	kubeconfig, err := TenantKubeconfig(clusterName, server, token)
	if err != nil {
		return err
	}

	return clientutil.CreateOrReplace(ctx, infraClient, c.kubeconfigSecret(clusterName, namespace, c.TenantKubeconfigSecretName(clusterName), kubeconfig))
}

// WaitServiceAccountToken waits until the token controller populates the
// ServiceAccount token Secret and returns it.
func WaitServiceAccountToken(ctx context.Context, c ctrlrt.Client, key ctrlrt.ObjectKey) (*corev1.Secret, error) {
	secret := &corev1.Secret{}

	err := wait.PollUntilContextTimeout(ctx, tokenWaitInterval, tokenWaitTimeout, true, func(ctx context.Context) (bool, error) {
		if err := c.Get(ctx, key, secret); err != nil {
			return false, ctrlrt.IgnoreNotFound(err)
		}

		return tokenPopulated(secret), nil
	})

	return secret, fail.KubeClient(err, "waiting for ServiceAccount token %s", key)
}

func tokenPopulated(secret *corev1.Secret) bool {
	return len(secret.Data[corev1.ServiceAccountTokenKey]) > 0 && len(secret.Data[corev1.ServiceAccountRootCAKey]) > 0
}
