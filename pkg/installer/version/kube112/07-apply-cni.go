package kube112

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates"
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
	ctx.Logger.Infoln("Applying Flannel CNI plugin…")

	node := ctx.Cluster.Hosts[0]
	logger := ctx.Logger.WithFields(logrus.Fields{
		"node": node.PublicAddress,
	})

	flannel, err := templates.FlannelConfiguration(ctx.Cluster)
	if err != nil {
		return fmt.Errorf("failed to create flannel configuration: %v", err)
	}
	ctx.Configuration.AddFile("kube-flannel.yaml", flannel)

	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	logger.Infoln("Uploading Flannel YAML manifest…")

	if err := ctx.Configuration.UploadTo(conn, "kubermatic-installer/cfg/"); err != nil {
		return fmt.Errorf("unable to upload flannel yaml manifest to node: %v", err)
	}

	_, _, _, err = util.RunShellCommand(conn, ctx.Verbose, `
set -xeu pipefail

export KUBECONFIG=/etc/kubernetes/admin.conf

sudo kubectl create -f ./{{ .WORK_DIR }}/cfg/kube-flannel.yaml
`, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
	})
	if err != nil {
		return err
	}

	return nil
}
