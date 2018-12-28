package kube112

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func deployCA(ctx *util.Context) error {
	ctx.Logger.Infoln("Deploying PKI…")
	return ctx.RunTaskOnAllNodes(deployCAOnNode, true)
}

func deployCAOnNode(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Uploading files…")
	return ctx.Configuration.UploadTo(conn, ctx.WorkDir)
}
