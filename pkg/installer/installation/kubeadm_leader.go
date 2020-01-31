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

func kubeadmCertsOnLeader(s *state.State) error {
	s.Logger.Infoln("Configuring certs and etcd on first controller…")
	return s.RunTaskOnLeader(kubeadmCertsExecutor)
}

func kubeadmCertsOnFollower(s *state.State) error {
	s.Logger.Infoln("Configuring certs and etcd on consecutive controller…")
	return s.RunTaskOnFollowers(kubeadmCertsExecutor, true)
}

func kubeadmCertsExecutor(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	s.Logger.Infoln("Ensuring Certificates…")
	cmd, err := scripts.KubeadmCert(s.WorkDir, node.ID, s.KubeadmVerboseFlag())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return err
}

func initKubernetesLeader(s *state.State) error {
	s.Logger.Infoln("Initializing Kubernetes on leader…")
	return s.RunTaskOnLeader(func(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
		s.Logger.Infoln("Running kubeadm…")

		cmd, err := scripts.KubeadmInit(s.WorkDir, node.ID, s.KubeadmVerboseFlag(), s.JoinToken, time.Hour.String())
		if err != nil {
			return err
		}

		_, _, err = s.Runner.RunRaw(cmd)

		return err
	})
}
