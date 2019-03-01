package upgrade

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/util"
)

func upgradeFollower(ctx *util.Context) error {
	return ctx.RunTaskOnFollowers(upgradeFollowerExecutor, false)
}

func upgradeFollowerExecutor(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Labeling follower control plane…")
	err := labelNode(ctx.Clientset.CoreV1().Nodes(), node)
	if err != nil {
		return errors.Wrap(err, "failed to label leader control plane node")
	}

	ctx.Logger.Infoln("Upgrading kubeadm on follower control plane…")
	err = upgradeKubeadm(ctx, node)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade kubeadm on follower control plane")
	}

	ctx.Logger.Infoln("Running 'kubeadm upgrade' on the follower control plane node…")
	err = upgradeFollowerControlPlane(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade follower control plane")
	}

	ctx.Logger.Infoln("Upgrading kubelet…")
	err = upgradeKubelet(ctx, node)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade kubelet")
	}

	ctx.Logger.Infoln("Unlabeling follower control plane…")
	err = unlabelNode(ctx.Clientset.CoreV1().Nodes(), node)
	if err != nil {
		return errors.Wrap(err, "failed to unlabel follower control plane node")
	}

	return nil
}
