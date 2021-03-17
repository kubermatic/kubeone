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

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/kubeadm/kubeadmargs"

	corev1 "k8s.io/api/core/v1"
	policybeta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	pspAdmissionPlugin = "PodSecurityPolicy"
	pspRoleNamespace   = metav1.NamespaceSystem
)

func activateKubeadmPSP(feature *kubeoneapi.PodSecurityPolicy, args *kubeadmargs.Args) {
	if feature == nil || !feature.Enable {
		return
	}

	args.APIServer.AppendMapStringStringExtraArg(apiServerAdmissionPluginsFlag, pspAdmissionPlugin)
}

func installKubeSystemPSP(psp *kubeoneapi.PodSecurityPolicy, s *state.State) error {
	if psp == nil || !psp.Enable {
		return nil
	}

	ctx := context.Background()
	k8sobjects := []client.Object{
		privilegedPSP(),
		privilegedPSPClusterRole(),
		privilegedPSPRoleBinding(),
	}

	for _, obj := range k8sobjects {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj); err != nil {
			return errors.Wrap(err, "failed to ensure PodSecurityPolicy role binding")
		}
	}

	return nil
}

func privilegedPSP() *policybeta1.PodSecurityPolicy {
	t := true

	return &policybeta1.PodSecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "privileged",
		},
		Spec: policybeta1.PodSecurityPolicySpec{
			Privileged:               true,
			HostNetwork:              true,
			HostIPC:                  true,
			HostPID:                  true,
			AllowPrivilegeEscalation: &t,
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
