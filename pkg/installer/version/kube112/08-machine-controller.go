package kube112

import (
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
)

func installMachineController(ctx *util.Context) error {
	if !*ctx.Cluster.MachineController.Deploy {
		ctx.Logger.Info("Skipping machine-controller deployment because it was disabled in configuration.")
		return nil
	}

	ctx.Logger.Infoln("Installing machine-controller…")
	if err := machinecontroller.Deploy(ctx); err != nil {
		return err
	}

	ctx.Logger.Infoln("Installing machine-controller webhooks…")
	if err := machinecontroller.DeployWebhookConfiguration(ctx); err != nil {
		return err
	}

	return nil
}
