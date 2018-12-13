package kube112

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func createJoinToken(ctx *util.Context) error {
	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, _ *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Creating join token…")

		// The command shows a
		// `I1213 20:32:32.578846   18179 version.go:236] remote version is much newer: v1.13.1; falling back to: stable-1.1`
		// message even thought all of Kubeadm, Kubelet, Kubectl and the control plane are
		// on 1.12, so we just grep out the part we care about
		stdout, _, _, err := util.RunCommand(conn, `
sudo kubeadm token create --print-join-command 2>&1|grep 'kubeadm join'|tr -d '\n'`, ctx.Verbose)
		if err != nil {
			return err
		}

		ctx.JoinCommand = stdout

		return nil
	})
}
