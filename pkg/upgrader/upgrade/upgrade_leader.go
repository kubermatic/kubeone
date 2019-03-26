/*
Copyright 2019 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	if err := labelNode(ctx.DynamicClient, node); err != nil {
		return errors.Wrap(err, "failed to label leader control plane node")
	}

	logger.Infoln("Upgrading Kubernetes binaries on leader control plane…")
	if err := upgradeKubernetesBinaries(ctx, node); err != nil {
		return errors.Wrap(err, "failed to upgrade kubernetes binaries on leader control plane")
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

	logger.Infoln("Unlabeling leader control plane…")
	if err := unlabelNode(ctx.DynamicClient, node); err != nil {
		return errors.Wrap(err, "failed to unlabel leader control plane node")
	}

	return nil
}
