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

package externalccm

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/addons"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	uninitializedTaint = "node.cloudprovider.kubernetes.io/uninitialized"
	ccmComponentLabel  = "ccm"
)

// Ensure external CCM deployen if Provider.External
func Ensure(s *state.State) error {
	if !s.Cluster.CloudProvider.External {
		return nil
	}

	s.Logger.Info("Ensure external CCM is up to date...")
	var err error

	switch {
	case s.Cluster.CloudProvider.Hetzner != nil:
		err = addons.EnsureAddonByName(s, resources.AddonCCMHetzner)
	case s.Cluster.CloudProvider.DigitalOcean != nil:
		err = addons.EnsureAddonByName(s, resources.AddonCCMDigitalOcean)
	case s.Cluster.CloudProvider.Packet != nil:
		err = addons.EnsureAddonByName(s, resources.AddonCCMPacket)
	case s.Cluster.CloudProvider.Openstack != nil:
		err = addons.EnsureAddonByName(s, resources.AddonCCMOpenStack)
	case s.Cluster.CloudProvider.Vsphere != nil:
		err = ensureVsphere(s)
	default:
		s.Logger.Infof("External CCM for %q not yet supported, skipping", s.Cluster.CloudProvider.CloudProviderName())
		return nil
	}

	if err != nil {
		return errors.Wrap(err, "failed to ensure CCM is installed")
	}

	err = waitForInitializedNodes(s)
	return errors.Wrap(err, "failed waiting for nodes to be initialized by CCM")
}

func waitForInitializedNodes(s *state.State) error {
	ctx := context.Background()

	s.Logger.Info("Waiting for nodes to initialize by CCM...")

	return wait.Poll(5*time.Second, 10*time.Minute, func() (bool, error) {
		nodes := corev1.NodeList{}

		if err := s.DynamicClient.List(ctx, &nodes); err != nil {
			return false, err
		}

		for _, node := range nodes.Items {
			for _, taint := range node.Spec.Taints {
				if taint.Key == uninitializedTaint && taint.Value == "true" {
					return false, nil
				}
			}
		}

		return true, nil
	})
}

func genClusterRoleBinding(name string, crole *rbacv1.ClusterRole, subj *corev1.ServiceAccount) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     crole.GetName(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      subj.GetName(),
				Namespace: subj.GetNamespace(),
			},
		},
	}
}
