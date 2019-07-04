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

package upgrade

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/state"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func upgradeMachineDeployments(s *state.State) error {
	if !s.UpgradeMachineDeployments {
		s.Logger.Info("Upgrade MachineDeployments skip per lack of flag…")
		return nil
	}

	s.Logger.Info("Upgrade MachineDeployments…")

	bg := context.Background()

	machineDeployments := clusterv1alpha1.MachineDeploymentList{}
	err := s.DynamicClient.List(
		bg,
		&dynclient.ListOptions{Namespace: metav1.NamespaceSystem},
		&machineDeployments,
	)
	if err != nil {
		return errors.Wrap(err, "failed to list MachineDeployments")
	}

	for _, md := range machineDeployments.Items {
		machineKey := dynclient.ObjectKey{Name: md.Name, Namespace: md.Namespace}

		retErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			machine := clusterv1alpha1.MachineDeployment{}
			if err := s.DynamicClient.Get(bg, machineKey, &machine); err != nil {
				return err
			}

			machine.Spec.Template.Spec.Versions.Kubelet = s.Cluster.Versions.Kubernetes
			return s.DynamicClient.Update(bg, &machine)
		})

		if retErr != nil {
			return errors.Wrapf(retErr, "failed to update MachineDeployment %s", md.Name)
		}
	}

	return nil
}
