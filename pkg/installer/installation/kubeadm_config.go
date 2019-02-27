package installation

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm"
	"github.com/kubermatic/kubeone/pkg/util"
)

func generateKubeadm(ctx *util.Context) error {
	ctx.Logger.Infoln("Generating kubeadm config file…")

	for idx := range ctx.Cluster.Hosts {
		kubeadm, err := kubeadm.Config(ctx, ctx.Cluster.Hosts[idx])
		if err != nil {
			return errors.Wrap(err, "failed to create kubeadm configuration")
		}

		ctx.Configuration.AddFile(fmt.Sprintf("cfg/master_%d.yaml", idx), kubeadm)
	}

	return ctx.RunTaskOnAllNodes(generateKubeadmOnNode, true)
}

func generateKubeadmOnNode(ctx *util.Context, _ *config.HostConfig, conn ssh.Connection) error {
	return errors.Wrap(ctx.Configuration.UploadTo(conn, ctx.WorkDir), "failed to upload")
}
