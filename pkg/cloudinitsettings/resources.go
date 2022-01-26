/*
Copyright 2022 The KubeOne Authors.

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

package cloudinitsettings

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// cloudInitSettingsNamespace is used in order to reach, authenticate and be authorized by the api server, to fetch
	// the machine provisioning secrets
	cloudInitSettingsNamespace = "cloud-init-settings"

	serviceAccountName = "cloud-init-getter"
	roleName           = "cloud-init-getter"
	roleBindingName    = "cloud-init-getter"
)

// Ensure creates/updates the credentials secret
func Ensure(s *state.State) error {
	if err := clientutil.CreateOrUpdate(s.Context, s.DynamicClient, namespace()); err != nil {
		return errors.Wrap(err, "unable to create cloud-init-settings namespace")
	}

	s.Logger.Infoln("Creating resources for cloud-init-settings...")

	ctx := context.Background()
	k8sobjects := []client.Object{
		namespace(),
		serviceAccount(),
		role(),
		roleBinding(),
	}

	for _, obj := range k8sobjects {
		if err := clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj); err != nil {
			return errors.Wrap(err, "failed to ensure CloudInitSettings namespace")
		}
	}
	return nil
}

func namespace() *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: cloudInitSettingsNamespace,
		},
	}
}
func serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: cloudInitSettingsNamespace,
		},
	}
}

func role() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: cloudInitSettingsNamespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list"},
			},
		},
	}
}

func roleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: cloudInitSettingsNamespace,
		},
		RoleRef: rbacv1.RoleRef{
			Name:     roleName,
			Kind:     "Role",
			APIGroup: rbacv1.GroupName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: rbacv1.ServiceAccountKind,
				Name: serviceAccountName,
			},
		},
	}
}
