package kube112

import (
	"strings"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func createJoinToken(ctx *util.Context) error {
	originalContext := ctx
	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, _ *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Creating join token…")

		stdout, _, err := ctx.Runner.Run(`sudo kubeadm token create --print-join-command`, nil)
		if err != nil {
			return err
		}

		stdout = strings.Replace(stdout, "\n", "", -1)
		originalContext.JoinCommand = stdout

		return nil
	})
}
