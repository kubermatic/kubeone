package kube112

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func installMachineController(ctx *util.Context) error {
	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, node config.HostConfig, _ int, conn ssh.Connection) error {
		ctx.Logger.Infoln("Installing machine-controller…")

		_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, `
kubectl apply -f ./{{ .WORK_DIR }}/machine-controller.yaml
kubectl apply -f ./{{ .WORK_DIR }}/machine-controller-webhook.yaml
`, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
		})

		return err
	})
}
