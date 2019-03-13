package features

import (
	"strings"

	"github.com/pkg/errors"

	kubeadmv1beta1 "github.com/kubermatic/kubeone/pkg/apis/kubeadm/v1beta1"
	"github.com/kubermatic/kubeone/pkg/templates"
	"github.com/kubermatic/kubeone/pkg/util"

	corev1 "k8s.io/api/core/v1"
	policybeta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	pspAdmissionPlugin            = "PodSecurityPolicy"
	apiServerAdmissionPluginsFlag = "enable-admission-plugins"
	pspRoleNamespace              = "kube-system"
)

var (
	defaultAdmissionPlugins = []string{
		"NamespaceLifecycle",
		"LimitRanger",
		"ServiceAccount",
		"PersistentVolumeClaimResize",
		"DefaultStorageClass",
		"DefaultTolerationSeconds",
		"MutatingAdmissionWebhook",
		"ValidatingAdmissionWebhook",
		"ResourceQuota",
		"Priority",
	}
)

func activateKubeadmPSP(activate bool, clusterConfig *kubeadmv1beta1.ClusterConfiguration) {
	if !activate {
		return
	}

	if clusterConfig.APIServer.ExtraArgs == nil {
		clusterConfig.APIServer.ExtraArgs = make(map[string]string)
	}

	if _, ok := clusterConfig.APIServer.ExtraArgs[apiServerAdmissionPluginsFlag]; ok {
		clusterConfig.APIServer.ExtraArgs[apiServerAdmissionPluginsFlag] += "," + pspAdmissionPlugin
	} else {
		clusterConfig.APIServer.ExtraArgs[apiServerAdmissionPluginsFlag] = strings.Join(append(defaultAdmissionPlugins, pspAdmissionPlugin), ",")
	}
}

func installKubeSystemPSP(activate bool, ctx *util.Context) error {
	if !activate {
		return nil
	}

	rbacClient := ctx.Clientset.RbacV1()

	err := templates.EnsurePodSecurityPolicy(ctx.Clientset.PolicyV1beta1().PodSecurityPolicies(), privilegedPSP())
	if err != nil {
		return errors.Wrap(err, "failed to ensure PodSecurityPolicy")
	}

	err = templates.EnsureClusterRole(rbacClient.ClusterRoles(), privilegedPSPClusterRole())
	if err != nil {
		return errors.Wrap(err, "failed to ensure PodSecurityPolicy cluster role")
	}

	err = templates.EnsureRoleBinding(rbacClient.RoleBindings(pspRoleNamespace), privilegedPSPRoleBinding())
	if err != nil {
		return errors.Wrap(err, "failed to ensure PodSecurityPolicy role binding")
	}

	return nil
}

func privilegedPSP() *policybeta1.PodSecurityPolicy {
	return &policybeta1.PodSecurityPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy/v1beta1",
			Kind:       "PodSecurityPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "privileged",
		},
		Spec: policybeta1.PodSecurityPolicySpec{
			Privileged:               true,
			HostNetwork:              true,
			HostIPC:                  true,
			HostPID:                  true,
			AllowPrivilegeEscalation: boolPtr(true),
			AllowedCapabilities:      []corev1.Capability{"*"},
			Volumes:                  []policybeta1.FSType{policybeta1.All},
			HostPorts: []policybeta1.HostPortRange{
				{Min: 0, Max: 65535},
			},
			RunAsUser: policybeta1.RunAsUserStrategyOptions{
				Rule: policybeta1.RunAsUserStrategyRunAsAny,
			},
			SELinux: policybeta1.SELinuxStrategyOptions{
				Rule: policybeta1.SELinuxStrategyRunAsAny,
			},
			SupplementalGroups: policybeta1.SupplementalGroupsStrategyOptions{
				Rule: policybeta1.SupplementalGroupsStrategyRunAsAny,
			},
			FSGroup: policybeta1.FSGroupStrategyOptions{
				Rule: policybeta1.FSGroupStrategyRunAsAny,
			},
		},
	}
}

func privilegedPSPClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "privileged-psp",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"policy"},
				Resources:     []string{"podsecuritypolicies"},
				Verbs:         []string{"use"},
				ResourceNames: []string{"privileged"},
			},
		},
	}
}

func privilegedPSPRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "privileged-psp",
			Namespace: pspRoleNamespace,
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "privileged-psp",
			Kind:     "ClusterRole",
			APIGroup: rbacv1.GroupName,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: rbacv1.GroupName,
				Kind:     "Group",
				Name:     "system:nodes",
			},
			{
				APIGroup: rbacv1.GroupName,
				Kind:     "Group",
				Name:     "system:serviceaccounts:" + pspRoleNamespace,
			},
		},
	}
}

func boolPtr(b bool) *bool {
	return &b
}
