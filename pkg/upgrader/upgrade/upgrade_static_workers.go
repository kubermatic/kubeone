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

func upgradeWorkers(s *state.State) error {
	return s.RunTaskOnWorkerHosts(upgradeWorkerHostsExecutor, false)
}

func upgradeWorkerHostsExecutor(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	logger := s.Logger.WithField("node", node.PublicAddress)

	logger.Infoln("Labeling static worker node…")
	err := labelNode(s.DynamicClient, node)
	if err != nil {
		return errors.Wrap(err, "failed to label static worker node")
	}

	logger.Infoln("Draining static worker node…")
	err = drainWorkerNode(s, *node)
	if err != nil {
		return errors.Wrap(err, "failed to drain static worker node")
	}

	logger.Infoln("Upgrading Kubernetes binaries on static worker node…")
	err = upgradeKubernetesBinaries(s, *node)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade kubernetes binaries on static worker node")
	}

	logger.Infoln("Running 'kubeadm upgrade' on the static worker node…")
	err = upgradeFollowerControlPlane(s)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade static worker node")
	}

	logger.Infoln("Uncordoning static worker node…")
	err = uncordonWorkerNode(s, *node)
	if err != nil {
		return errors.Wrap(err, "failed to uncordon static worker node")
	}

	logger.Infof("Waiting %v to ensure all components are up…", timeoutNodeUpgrade)
	time.Sleep(timeoutNodeUpgrade)

	logger.Infoln("Unlabeling static worker node…")
	err = unlabelNode(s.DynamicClient, node)
	if err != nil {
		return errors.Wrap(err, "failed to unlabel static worker node node")
	}

	return nil
}
