package upgrade

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/util"

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
		&dynclient.ListOptions{Namespace: "kube-system"},
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
