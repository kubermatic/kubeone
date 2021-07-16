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
	"k8c.io/kubeone/pkg/nodeutils"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
)

func upgradeStaticWorkers(s *state.State) error {
	// we upgrade seqentially to minimize cluster disruption
	return s.RunTaskOnStaticWorkers(upgradeStaticWorkersExecutor, state.RunSequentially)
}

func upgradeStaticWorkersExecutor(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	logger := s.Logger.WithField("node", node.PublicAddress)

	logger.Infoln("Labeling static worker node...")

	if err := labelNode(s.DynamicClient, node); err != nil {
		return errors.Wrap(err, "failed to label static worker node")
	}

	drainer := nodeutils.NewDrainer(s.RESTConfig, logger)

	logger.Infoln("Cordoning static worker node...")
	if err := drainer.Cordon(s.Context, node.Hostname, true); err != nil {
		return errors.Wrap(err, "failed to cordon follower control plane node")
	}

	logger.Infoln("Draining static worker node...")
	if err := drainer.Drain(s.Context, node.Hostname); err != nil {
		return errors.Wrap(err, "failed to drain follower control plane node")
	}

	logger.Infoln("Upgrading Kubernetes binaries on static worker node...")
	if err := upgradeKubeadmAndCNIBinaries(s, *node); err != nil {
		return errors.Wrap(err, "failed to upgrade kubernetes binaries on static worker node")
	}

	logger.Infoln("Running 'kubeadm upgrade' on the static worker node...")
	if err := upgradeStaticWorker(s); err != nil {
		return errors.Wrap(err, "failed to upgrade static worker node")
	}

	logger.Infoln("Upgrading kubernetes system binaries on the static worker node...")
	if err := upgradeKubeletAndKubectlBinaries(s, *node); err != nil {
		return errors.Wrap(err, "failed to upgrade kubernetes system binaries on the static worker node")
	}

	logger.Infoln("Uncordoning static worker node...")
	if err := drainer.Cordon(s.Context, node.Hostname, false); err != nil {
		return errors.Wrap(err, "failed to uncordon follower control plane node")
	}

	logger.Infof("Waiting %v to ensure all components are up...", timeoutNodeUpgrade)
	time.Sleep(timeoutNodeUpgrade)

	logger.Infoln("Unlabeling static worker node...")
	if err := unlabelNode(s.DynamicClient, node); err != nil {
		return errors.Wrap(err, "failed to unlabel static worker node node")
	}

	return nil
}
