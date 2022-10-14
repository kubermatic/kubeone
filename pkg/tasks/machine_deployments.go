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

package tasks

import (
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/machinecontroller"

	clusterv1alpha1 "github.com/kubermatic/machine-controller/pkg/apis/cluster/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	machineDeploymentsDocsLink = `https://docs.kubermatic.com/kubeone/main/guides/machine-controller/`
)

func createMachineDeployments(s *state.State) error {
	if len(s.Cluster.DynamicWorkers) == 0 {
		return nil
	}

	if !s.CreateMachineDeployments {
		s.Logger.Info("Skipped creating MachineDeployments.")

		return nil
	}

	s.Logger.Infoln("Creating worker machines...")
	s.Logger.Warnln("KubeOne will not manage MachineDeployments objects besides initially creating them and optionally upgrading them...")
	s.Logger.Warnf("For more info about MachineDeployments see: %s", machineDeploymentsDocsLink)

	return machinecontroller.CreateMachineDeployments(s)
}

func upgradeMachineDeployments(s *state.State) error {
	if !s.UpgradeMachineDeployments {
		s.Logger.Info("Upgrade MachineDeployments skip per lack of flag...")

		return nil
	}

	s.Logger.Info("Upgrade MachineDeployments...")

	machineDeployments := clusterv1alpha1.MachineDeploymentList{}
	err := s.DynamicClient.List(
		s.Context,
		&machineDeployments,
		dynclient.InNamespace(metav1.NamespaceSystem),
	)
	if err != nil {
		return fail.KubeClient(err, "getting %T", machineDeployments)
	}

	for _, md := range machineDeployments.Items {
		machineKey := dynclient.ObjectKey{Name: md.Name, Namespace: md.Namespace}

		retErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			machine := clusterv1alpha1.MachineDeployment{}
			if err = s.DynamicClient.Get(s.Context, machineKey, &machine); err != nil {
				return err
			}

			machine.Spec.Template.Spec.Versions.Kubelet = s.Cluster.Versions.Kubernetes

			return s.DynamicClient.Update(s.Context, &machine)
		})

		if retErr != nil {
			return fail.KubeClient(err, "updating %T %s", md, machineKey)
		}
	}

	return nil
}
