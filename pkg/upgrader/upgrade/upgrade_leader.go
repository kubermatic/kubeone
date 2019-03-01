package upgrade

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/util"
)

func upgradeLeader(ctx *util.Context) error {
	return ctx.RunTaskOnLeader(upgradeLeaderExecutor)
}

func upgradeLeaderExecutor(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	logger := ctx.Logger.WithField("node", node.PublicAddress)

	logger.Infoln("Labeling leader control plane…")
	if err := labelNode(ctx.Clientset.CoreV1().Nodes(), node); err != nil {
		return errors.Wrap(err, "failed to label leader control plane node")
	}

	logger.Infoln("Upgrading kubeadm on leader control plane…")
	if err := upgradeKubeadm(ctx, node); err != nil {
		return errors.Wrap(err, "failed to upgrade kubeadm on leader control plane")
	}

	logger.Infoln("Generating kubeadm config …")
	if err := generateKubeadmConfig(ctx, node); err != nil {
		return errors.Wrap(err, "failed to generate kubeadm config")
	}

	logger.Infoln("Uploading kubeadm config to leader control plane node…")
	if err := uploadKubeadmConfig(ctx, conn); err != nil {
		return errors.Wrap(err, "failed to upload kubeadm config")
	}

	logger.Infoln("Running 'kubeadm upgrade' on leader control plane node…")
	if err := upgradeLeaderControlPlane(ctx); err != nil {
		return errors.Wrap(err, "failed to run 'kubeadm upgrade' on leader control plane")
	}

	logger.Infoln("Upgrading kubelet on leader control plane…")
	if err := upgradeKubelet(ctx, node); err != nil {
		return errors.Wrap(err, "failed to upgrade kubelet on leader control plane")
	}

	logger.Infoln("Unlabeling leader control plane…")
	if err := unlabelNode(ctx.Clientset.CoreV1().Nodes(), node); err != nil {
		return errors.Wrap(err, "failed to unlabel leader control plane node")
	}

	return nil
}
