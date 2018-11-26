package kube111

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func initKubernetesLeader(ctx *util.Context) error {
	ctx.Logger.Infoln("Initializing Kubernetes on leader…")

	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, _ config.HostConfig, _ int, conn ssh.Connection) error {
		ctx.Logger.Infoln("Running kubeadm…")

		_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, `sudo kubeadm init --config=./{{ .WORK_DIR }}/cfg/master_0.yaml`, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
		})

		return err
	})
}
