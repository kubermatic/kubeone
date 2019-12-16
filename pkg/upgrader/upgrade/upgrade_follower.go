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

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

func upgradeFollower(s *state.State) error {
	return s.RunTaskOnFollowers(upgradeFollowerExecutor, false)
}

func upgradeFollowerExecutor(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	logger := s.Logger.WithField("node", node.PublicAddress)

	logger.Infoln("Labeling follower control plane…")
	err := labelNode(s.DynamicClient, node)
	if err != nil {
		return errors.Wrap(err, "failed to label leader control plane node")
	}

	logger.Infoln("Draining follower control plane…")
	if err := drainNode(s, *node); err != nil {
		return errors.Wrap(err, "failed to drain leader control plane node")
	}

	logger.Infoln("Upgrading Kubernetes binaries on follower control plane…")
	err = upgradeKubernetesBinaries(s, *node)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade kubernetes binaries on follower control plane")
	}

	logger.Infof("Waiting %v seconds to ensure kubelet is up…", timeoutKubeletUpgrade.String())
	time.Sleep(timeoutKubeletUpgrade)

	logger.Infoln("Running 'kubeadm upgrade' on the follower control plane node…")
	err = upgradeFollowerControlPlane(s)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade follower control plane")
	}

	logger.Infof("Waiting %v seconds to ensure all components are up…", timeoutNodeUpgrade.String())
	time.Sleep(timeoutNodeUpgrade)

	logger.Infoln("Cordoning follower control plane…")
	if err := cordonNode(s, *node); err != nil {
		return errors.Wrap(err, "failed to cordon follower control plane node")
	}

	logger.Infoln("Unlabeling follower control plane…")
	err = unlabelNode(s.DynamicClient, node)
	if err != nil {
		return errors.Wrap(err, "failed to unlabel follower control plane node")
	}

	return nil
}
