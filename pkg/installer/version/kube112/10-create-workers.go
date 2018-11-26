package kube112

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func createWorkerMachines(ctx *util.Context) error {
	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, _ config.HostConfig, _ int, conn ssh.Connection) error {
		ctx.Logger.Infoln("Creating worker machines…")

		_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, `kubectl apply -f ./{{ .WORK_DIR }}/workers.yaml`, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
		})

		return err
	})
}
