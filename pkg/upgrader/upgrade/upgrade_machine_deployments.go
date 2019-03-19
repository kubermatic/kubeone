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

	"github.com/kubermatic/kubeone/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func upgradeMachineDeployments(ctx *util.Context) error {
	if !ctx.UpgradeMachineDeployments {
		ctx.Logger.Info("Upgrade MachineDeployments skip per lack of flag…")
		return nil
	}

	ctx.Logger.Info("Upgrade MachineDeployments…")

	machineDeployments := clusterv1alpha1.MachineDeploymentList{}
	err := ctx.DynamicClient.List(
		context.Background(),
		&dynclient.ListOptions{Namespace: metav1.NamespaceSystem},
		&machineDeployments,
	)
	if err != nil {
		return errors.Wrap(err, "failed to list MachineDeployments")
	}

	for _, md := range machineDeployments.Items {
		md := md
		md.Spec.Template.Spec.Versions.Kubelet = ctx.Cluster.Versions.Kubernetes
		err := ctx.DynamicClient.Update(context.Background(), &md)
		if err != nil {
			return errors.Wrap(err, "failed to upgrade MachineDeployment")
		}
	}

	return nil
}
