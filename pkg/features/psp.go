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

package features

import (
	"context"

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/kubeadmargs"

	corev1 "k8s.io/api/core/v1"
	policybeta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	pspAdmissionPlugin            = "PodSecurityPolicy"
	apiServerAdmissionPluginsFlag = "enable-admission-plugins"
	pspRoleNamespace              = metav1.NamespaceSystem
)

func updatePSPKubeadmConfig(feature *kubeoneapi.PodSecurityPolicy, args *kubeadmargs.Args) {
	if feature == nil {
		return
	}

	if !feature.Enable {
		return
	}

	args.APIServer.AppendMapStringStringExtraArg(apiServerAdmissionPluginsFlag, pspAdmissionPlugin)
}

func installKubeSystemPSP(psp *kubeoneapi.PodSecurityPolicy, s *state.State) error {
	if psp == nil {
		return nil
	}

	if !psp.Enable {
		return nil
	}

	ctx := context.Background()
	okFunc := func(runtime.Object) error { return nil }

	_, err := controllerutil.CreateOrUpdate(ctx, s.DynamicClient, privilegedPSP(), okFunc)
	if err != nil {
		return errors.Wrap(err, "failed to ensure PodSecurityPolicy")
	}

	_, err = controllerutil.CreateOrUpdate(ctx, s.DynamicClient, privilegedPSPClusterRole(), okFunc)
	if err != nil {
		return errors.Wrap(err, "failed to ensure PodSecurityPolicy cluster role")
	}

	_, err = controllerutil.CreateOrUpdate(ctx, s.DynamicClient, privilegedPSPRoleBinding(), okFunc)
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
