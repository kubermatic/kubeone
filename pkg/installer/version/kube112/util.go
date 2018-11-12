package kube112

import (
	"fmt"
	"github.com/kubermatic/kubeone/pkg/config"
	"time"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func waitForApiserver(ctx *util.Context, node config.HostConfig, conn ssh.Connection) error {
	var err error

	ctx.Logger.Infoln("Waiting for apiserver…")

	command := fmt.Sprintf(
		`curl --max-time 3 --fail --cacert /etc/kubernetes/pki/ca.crt https://%s:6443/healthz`,
		node.PublicAddress)

	for remaining := 20; remaining >= 0; remaining-- {
		_, _, _, err = util.RunCommand(conn, command, ctx.Verbose)
		if err == nil {
			break
		}
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("Kubernetes apiserver did not come up, giving up")
	}

	return nil
}
