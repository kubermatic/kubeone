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

func upgradeLeader(s *state.State) error {
	return s.RunTaskOnLeader(upgradeLeaderExecutor)
}

func upgradeLeaderExecutor(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	logger := s.Logger.WithField("node", node.PublicAddress)

	logger.Infoln("Labeling leader control plane…")
	if err := labelNode(s.DynamicClient, node); err != nil {
		return errors.Wrap(err, "failed to label leader control plane node")
	}

	logger.Infoln("Draining leader control plane…")
	if err := drainNode(s, *node); err != nil {
		return errors.Wrap(err, "failed to drain leader control plane node")
	}

	logger.Infoln("Upgrading Kubernetes binaries on leader control plane…")
	if err := upgradeKubernetesBinaries(s, *node); err != nil {
		return errors.Wrap(err, "failed to upgrade kubernetes binaries on leader control plane")
	}

	logger.Infoln("Generating kubeadm config …")
	if err := generateKubeadmConfig(s, *node); err != nil {
		return errors.Wrap(err, "failed to generate kubeadm config")
	}

	logger.Infoln("Uploading kubeadm config to leader control plane node…")
	if err := uploadKubeadmConfig(s, conn); err != nil {
		return errors.Wrap(err, "failed to upload kubeadm config")
	}

	logger.Infoln("Running 'kubeadm upgrade' on leader control plane node…")
	if err := upgradeLeaderControlPlane(s); err != nil {
		return errors.Wrap(err, "failed to run 'kubeadm upgrade' on leader control plane")
	}

	logger.Infoln("Uncordoning leader control plane…")
	if err := uncordonNode(s, *node); err != nil {
		return errors.Wrap(err, "failed to uncordon leader control plane node")
	}

	logger.Infof("Waiting %v seconds to ensure all components are up…", timeoutNodeUpgrade.String())
	time.Sleep(timeoutNodeUpgrade)

	logger.Infoln("Unlabeling leader control plane…")
	if err := unlabelNode(s.DynamicClient, node); err != nil {
		return errors.Wrap(err, "failed to unlabel leader control plane node")
	}

	return nil
}
