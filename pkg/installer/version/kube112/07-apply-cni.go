package kube112

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func applyCNI(ctx *util.Context, cni string) error {
	switch cni {
	case "flannel":
		return applyFlannelCNI(ctx)
	default:
		return fmt.Errorf("unknown CNI plugin selected")
	}
}

func applyFlannelCNI(ctx *util.Context) error {
	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, _ *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Applying Flannel CNI plugin…")

		_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, `sudo kubectl create -f ./{{ .WORK_DIR }}/kube-flannel.yaml`, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
		})

		return err
	})
}
