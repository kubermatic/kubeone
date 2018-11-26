package kube112

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/sirupsen/logrus"
)

func applyCNI(ctx *util.Context, cni string) error {
	switch cni {
	case "flannel":
		return applyFlannelCNI(ctx)
	default:
		return fmt.Errorf("unknown cni plugin selected")
	}
}

func applyFlannelCNI(ctx *util.Context) error {
	node := ctx.Cluster.Hosts[0]
	logger := ctx.Logger.WithFields(logrus.Fields{
		"node": node.PublicAddress,
	})

	logger.Infoln("Applying Flannel CNI plugin…")

	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	_, _, _, err = util.RunShellCommand(conn, ctx.Verbose, `
set -xeu pipefail

export KUBECONFIG=/etc/kubernetes/admin.conf

sudo kubectl create -f ./{{ .WORK_DIR }}/kube-flannel.yaml
`, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
	})
	if err != nil {
		return err
	}

	return nil
}
