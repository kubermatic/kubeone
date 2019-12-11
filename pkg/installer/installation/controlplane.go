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

package installation

import (
	"time"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/scripts"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

func joinControlplaneNode(s *state.State) error {
	s.Logger.Infoln("Joining controlplane node…")
	return s.RunTaskOnFollowers(joinControlPlaneNodeInternal, false)
}

func joinControlPlaneNodeInternal(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	logger := s.Logger.WithField("node", node.PublicAddress)

	sleepTime := 30 * time.Second
	logger.Infof("Waiting %s to ensure main control plane components are up…", sleepTime)
	time.Sleep(sleepTime)

	logger.Info("Joining control plane node")
	cmd, err := scripts.KubeadmJoin(s.WorkDir, node.ID, s.KubeadmVerboseFlag())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return err
}
