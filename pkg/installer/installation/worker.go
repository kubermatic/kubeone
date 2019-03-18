package installation

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/util"
)

func createWorkerMachines(ctx *util.Context) error {
	if len(ctx.Cluster.Workers) == 0 {
		return nil
	}

	ctx.Logger.Infoln("Waiting for machine-controller to come up…")
	if err := machinecontroller.WaitForWebhook(ctx.DynamicClient); err != nil {
		return errors.Wrap(err, "machine-controller-webhook did not come up")
	}

	if err := machinecontroller.WaitForMachineController(ctx.DynamicClient); err != nil {
		return errors.Wrap(err, "machine-controller did not come up")
	}

	// it can still take a bit before the MC is actually ready
	time.Sleep(10 * time.Second)

	ctx.Logger.Infoln("Creating worker machines…")
	return machinecontroller.DeployMachineDeployments(ctx)
}
