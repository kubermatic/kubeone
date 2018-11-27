package kube111

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func createJoinToken(ctx *util.Context) error {
	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, _ config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Creating join token…")

		stdout, _, _, err := util.RunCommand(conn, `sudo kubeadm token create --print-join-command`, ctx.Verbose)
		if err != nil {
			return err
		}

		ctx.JoinCommand = stdout

		return nil
	})
}
