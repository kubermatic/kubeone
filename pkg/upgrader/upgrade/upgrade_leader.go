package upgrade

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func upgradeLeader(ctx *util.Context) error {
	return ctx.RunTaskOnLeader(upgradeLeaderExecutor)
}

func upgradeLeaderExecutor(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	logger := ctx.Logger.WithField("node", node.PublicAddress)

	logger.Infoln("Labeling leader control plane…")
	err := labelNode(ctx.Clientset.CoreV1().Nodes(), node)
	if err != nil {
		return errors.Wrap(err, "failed to label leader control plane node")
	}

	logger.Infoln("Upgrading kubeadm on leader control plane…")
	err = upgradeKubeadm(ctx, node)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade kubeadm on leader control plane")
	}

	logger.Infoln("Running 'kubeadm upgrade' on leader control plane node…")
	err = upgradeLeaderControlPlane(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to run 'kubeadm upgrade' on leader control plane")
	}

	logger.Infoln("Upgrading kubelet on leader control plane…")
	err = upgradeKubelet(ctx, node)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade kubelet on leader control plane")
	}

	logger.Infoln("Unlabeling leader control plane…")
	err = unlabelNode(ctx.Clientset.CoreV1().Nodes(), node)
	if err != nil {
		return errors.Wrap(err, "failed to unlabel leader control plane node")
	}

	return nil
}
