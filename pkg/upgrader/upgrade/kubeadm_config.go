package upgrade

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm"
	"github.com/kubermatic/kubeone/pkg/util"
)

func generateKubeadmConfig(ctx *util.Context, node *config.HostConfig) error {
	kubeadmConf, err := kubeadm.Config(ctx, node)
	if err != nil {
		return errors.Wrap(err, "failed to create kubeadm configuration")
	}

	ctx.Configuration.AddFile("cfg/master_0.yaml", kubeadmConf)

	return nil
}

func uploadKubeadmConfig(ctx *util.Context, sshConn ssh.Connection) error {
	return errors.Wrap(ctx.Configuration.UploadTo(sshConn, ctx.WorkDir), "failed to upload")
}
