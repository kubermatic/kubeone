package kube111

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/sirupsen/logrus"
)

func initKubernetesLeader(ctx *util.Context) error {
	ctx.Logger.Infoln("Initializing Kubernetes on leader…")

	node := ctx.Cluster.Hosts[0]
	logger := ctx.Logger.WithFields(logrus.Fields{
		"node": node.PublicAddress,
	})

	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	logger.Infoln("Running kubeadm…")

	_, _, _, err = util.RunShellCommand(conn, ctx.Verbose, `
set -xeu
export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"
sudo kubeadm init \
     --config=./{{ .WORK_DIR }}/cfg/master_0.yaml
`, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
	})
	if err != nil {
		return err
	}

	return nil
}
