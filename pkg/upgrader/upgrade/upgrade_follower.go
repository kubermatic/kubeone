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
	"time"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/util"
)

func upgradeFollower(ctx *util.Context) error {
	return ctx.RunTaskOnFollowers(upgradeFollowerExecutor, false)
}

func upgradeFollowerExecutor(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	logger := ctx.Logger.WithField("node", node.PublicAddress)

	logger.Infoln("Labeling follower control plane…")
	err := labelNode(ctx.DynamicClient, node)
	if err != nil {
		return errors.Wrap(err, "failed to label leader control plane node")
	}

	logger.Infoln("Upgrading Kubernetes binaries on follower control plane…")
	err = upgradeKubernetesBinaries(ctx, node)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade kubernetes binaries on follower control plane")
	}

	logger.Infoln("Running 'kubeadm upgrade' on the follower control plane node…")
	err = upgradeFollowerControlPlane(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade follower control plane")
	}

	logger.Infoln("Unlabeling follower control plane…")
	err = unlabelNode(ctx.DynamicClient, node)
	if err != nil {
		return errors.Wrap(err, "failed to unlabel follower control plane node")
	}

	logger.Infoln("Waiting 10 seconds to ensure all components are up…")
	time.Sleep(10 * time.Second)

	return nil
}
