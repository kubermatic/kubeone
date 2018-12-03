package kube112

import (
	"errors"
	"fmt"
	"time"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
)

func createWorkerMachines(ctx *util.Context) error {
	if len(ctx.Cluster.Workers) == 0 {
		return nil
	}

	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, _ config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Waiting for machine-controller to come up…")

		cmd := fmt.Sprintf(
			`kubectl -n "%s" get pods -l '%s=%s' -o jsonpath='{.items[0].status.phase}'`,
			machinecontroller.WebhookNamespace,
			machinecontroller.WebhookAppLabelKey,
			machinecontroller.WebhookAppLabelValue,
		)
		if !util.WaitForCondition(conn, ctx.Verbose, cmd, 1*time.Minute, util.IsRunning) {
			return errors.New("machine-controller-webhook did not come up")
		}

		cmd = fmt.Sprintf(
			`kubectl -n "%s" get pods -l '%s=%s' -o jsonpath='{.items[0].status.phase}'`,
			machinecontroller.MachineControllerNamespace,
			machinecontroller.MachineControllerAppLabelKey,
			machinecontroller.MachineControllerAppLabelValue,
		)
		if !util.WaitForCondition(conn, ctx.Verbose, cmd, 1*time.Minute, util.IsRunning) {
			return errors.New("machine-controller did not come up")
		}

		// it can still take a bit before the MC is actually ready
		time.Sleep(5 * time.Second)

		ctx.Logger.Infoln("Creating worker machines…")
		_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, `kubectl apply -f ./{{ .WORK_DIR }}/workers.yaml`, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
		})

		return err
	})
}
