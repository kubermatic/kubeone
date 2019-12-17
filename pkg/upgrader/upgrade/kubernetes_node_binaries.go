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

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/scripts"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

func upgradeKubernetesNodeBinaries(s *state.State) error {
	return s.RunTaskOnAllNodes(upgradeKubernetesNodeBinariesExecutor, false)
}

func upgradeKubernetesNodeBinariesExecutor(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	logger := s.Logger.WithField("node", node.PublicAddress)

	logger.Infoln("Upgrading Kubernetes node binaries on control planes…")
	if err := upgradeKubernetesNodeBinariesScript(s, *node); err != nil {
		return errors.Wrap(err, "failed to upgrade kubernetes binaries on leader control plane")
	}

	return nil
}

func upgradeKubernetesNodeBinariesScript(s *state.State, node kubeoneapi.HostConfig) error {
	var err error

	switch node.OperatingSystem {
	case "ubuntu", "debian":
		err = upgradeKubernetesNodeBinariesDebian(s)
	case "coreos":
		err = upgradeKubernetesNodeBinariesCoreOS(s)
	case "centos":
		err = upgradeKubernetesNodeBinariesCentOS(s)
	default:
		err = errors.Errorf("'%s' is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func upgradeKubernetesNodeBinariesDebian(s *state.State) error {
	cmd, err := scripts.UpgradeKubeletAndKubectlDebian(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func upgradeKubernetesNodeBinariesCentOS(s *state.State) error {
	cmd, err := scripts.UpgradeKubeletAndKubectlCentOS(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func upgradeKubernetesNodeBinariesCoreOS(s *state.State) error {
	cmd, err := scripts.UpgradeKubeletAndKubectlCoreOS(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}
