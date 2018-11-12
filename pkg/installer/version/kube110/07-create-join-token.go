package kube110

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/sirupsen/logrus"
)

func createJoinToken(ctx *util.Context) error {
	node := ctx.Manifest.Hosts[0]
	logger := ctx.Logger.WithFields(logrus.Fields{
		"node": node.PublicAddress,
	})

	logger.Infoln("Creating join token…")

	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	stdout, _, _, err := util.RunCommand(conn, `
set -xeu pipefail

export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"

sudo kubeadm token create --print-join-command`, ctx.Verbose)
	if err != nil {
		return err
	}

	ctx.JoinCommand = stdout

	return nil
}
