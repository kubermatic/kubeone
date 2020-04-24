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
	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/scripts"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

func joinStaticWorkerNodes(s *state.State) error {
	return s.RunTaskOnStaticWorkers(joinStaticWorkerInternal, state.RunParallel)
}

func joinStaticWorkerInternal(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	logger := s.Logger.WithField("node", node.PublicAddress)

	logger.Info("Joining worker node")
	cmd, err := scripts.KubeadmJoinWorker(s.WorkDir, node.ID, s.KubeadmVerboseFlag())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return err
}
