package ark

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// namespace deploys default Ark namespace (heptio-ark)
func namespace() *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: arkNamespace,
		},
	}
}

// serviceAccount creates ServiceAccount used by Ark pods
func serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ark",
			Namespace: arkNamespace,
			Labels: map[string]string{
				"component": "ark",
			},
		},
	}
}

// clusterRoleBinding is a ClusterAdmin RoleBinding allowing Ark to do backups and restores
func clusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ark",
			Labels: map[string]string{
				"component": "ark",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "cluster-admin",
			Kind:     "ClusterRole",
			APIGroup: rbacv1.GroupName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      arkServiceAccount,
				Namespace: arkNamespace,
			},
		},
	}
}
