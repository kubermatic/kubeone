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

package tasks

import (
	"time"

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
)

func upgradeFollower(s *state.State) error {
	return s.RunTaskOnFollowers(upgradeFollowerExecutor, state.RunSequentially)
}

func upgradeFollowerExecutor(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	logger := s.Logger.WithField("node", node.PublicAddress)

	logger.Infoln("Labeling follower control plane…")
	if err := labelNode(s.DynamicClient, node); err != nil {
		return errors.Wrap(err, "failed to label follower control plane node")
	}

	logger.Infoln("Draining follower control plane…")
	if err := drainNode(s, *node); err != nil {
		return errors.Wrap(err, "failed to drain follower control plane node")
	}

	logger.Infoln("Upgrading Kubernetes binaries on follower control plane…")
	if err := upgradeKubeadmAndCNIBinaries(s, *node); err != nil {
		return errors.Wrap(err, "failed to upgrade kubernetes binaries on follower control plane")
	}

	logger.Infoln("Running 'kubeadm upgrade' on the follower control plane node…")
	if err := upgradeFollowerControlPlane(s); err != nil {
		return errors.Wrap(err, "failed to upgrade follower control plane")
	}

	logger.Infoln("Upgrading kubernetes system binaries on the follower control plane…")
	if err := upgradeKubeletAndKubectlBinaries(s, *node); err != nil {
		return errors.Wrap(err, "failed to upgrade kubernetes system binaries on follower control plane")
	}

	logger.Infoln("Uncordoning follower control plane…")
	if err := uncordonNode(s, *node); err != nil {
		return errors.Wrap(err, "failed to uncordon follower control plane node")
	}

	logger.Infof("Waiting %v to ensure all components are up…", timeoutNodeUpgrade)
	time.Sleep(timeoutNodeUpgrade)

	logger.Infoln("Unlabeling follower control plane…")
	if err := unlabelNode(s.DynamicClient, node); err != nil {
		return errors.Wrap(err, "failed to unlabel follower control plane node")
	}

	return nil
}
