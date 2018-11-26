package kube110

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// Reset undos all changes made by KubeOne to the configured machines.
func Reset(ctx *util.Context) error {
	ctx.Logger.Infoln("Resetting kubeadm…")

	return util.RunTaskOnAllNodes(ctx, resetNode)
}

func resetNode(ctx *util.Context, node config.HostConfig, _ int, conn ssh.Connection) error {
	ctx.Logger.Infoln("Resetting node…")

	_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, resetScript, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
	})

	return err
}

const resetScript = `
sudo kubeadm reset
rm -rf "{{ .WORK_DIR }}"
`
