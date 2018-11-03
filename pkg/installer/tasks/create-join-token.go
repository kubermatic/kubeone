package tasks

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type CreateJoinTokenTask struct{}

func (t *CreateJoinTokenTask) Execute(ctx *Context) error {
	ctx.Logger.Infoln("Creating join token…")

	node := ctx.Manifest.Hosts[0]
	logger := ctx.Logger.WithFields(logrus.Fields{
		"node":   node.Address,
		"master": true,
	})

	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.Address, err)
	}

	logger.Infoln("Running kubeadm…")
	stdout, stderr, _, err := conn.Exec(`
set -xeu pipefail

export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"

sudo kubeadm token create --print-join-command`)
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr)
	}

	ctx.JoinCommand = stdout

	return err
}
