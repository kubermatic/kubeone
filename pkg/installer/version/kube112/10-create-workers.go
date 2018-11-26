package kube112

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/sirupsen/logrus"
)

func createWorkerMachines(ctx *util.Context) error {
	node := ctx.Cluster.Hosts[0]
	logger := ctx.Logger.WithFields(logrus.Fields{
		"node": node.PublicAddress,
	})

	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	logger.Infoln("Creating worker machines…")

	_, _, _, err = util.RunShellCommand(conn, ctx.Verbose, `
set -xeu pipefail

kubectl apply -f ./{{ .WORK_DIR }}/workers.yaml
`, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
	})

	return err
}
