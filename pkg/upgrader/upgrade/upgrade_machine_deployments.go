package upgrade

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/installer/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterclientset "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

func upgradeMachineDeployments(ctx *util.Context) error {
	if !ctx.UpgradeMachineDeployments {
		ctx.Logger.Info("Upgrade MachineDeployments skip per lack of flag…")
		return nil
	}

	ctx.Logger.Info("Upgrade MachineDeployments…")

	clusterapiClientset, err := clusterclientset.NewForConfig(ctx.RESTConfig)
	if err != nil {
		return errors.Wrap(err, "failed to get clusterapiClientset")
	}

	clusterapiClient := clusterapiClientset.ClusterV1alpha1()
	machineDeploymentsClient := clusterapiClient.MachineDeployments("kube-system")
	machineDeployments, err := machineDeploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list MachineDeployments")
	}

	for _, md := range machineDeployments.Items {
		md := md
		md.Spec.Template.Spec.Versions.Kubelet = ctx.Cluster.Versions.Kubernetes
		_, err := machineDeploymentsClient.Update(&md)
		if err != nil {
			return errors.Wrap(err, "failed to upgrade MachineDeployments")
		}
	}

	return nil
}
