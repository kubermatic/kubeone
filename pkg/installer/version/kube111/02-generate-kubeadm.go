package kube111

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates"
)

func generateKubeadm(ctx *util.Context) error {
	ctx.Logger.Infoln("Generating kubeadm config file…")

	for idx := range ctx.Cluster.Hosts {
		kubeadm, err := templates.KubeadmConfig(ctx.Cluster, idx)
		if err != nil {
			return fmt.Errorf("failed to create kubeadm configuration: %v", err)
		}

		ctx.Configuration.AddFile(fmt.Sprintf("cfg/master_%d.yaml", idx), kubeadm)
	}

	return util.RunTaskOnNodes(ctx, generateKubeadmOnNode)
}

func generateKubeadmOnNode(ctx *util.Context, _ config.HostConfig, _ int, conn ssh.Connection) error {
	err := ctx.Configuration.UploadTo(conn, ctx.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to upload: %v", err)
	}

	return nil
}
