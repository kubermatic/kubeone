package installation

import (
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/util"
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
