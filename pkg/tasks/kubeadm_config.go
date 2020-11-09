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
	"fmt"

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/kubeadm"
)

func determinePauseImage(s *state.State) error {
	if s.Cluster.RegistryConfiguration == nil || s.Cluster.RegistryConfiguration.OverwriteRegistry == "" {
		return nil
	}

	s.Logger.Infoln("Determining Kubernetes pause image...")

	return s.RunTaskOnLeader(determinePauseImageExecutor)
}

func determinePauseImageExecutor(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	cmd, err := scripts.KubeadmPauseImageVersion(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	out, _, err := s.Runner.RunRaw(cmd)
	if err != nil {
		return err
	}

	s.PauseImage = s.Cluster.RegistryConfiguration.ImageRegistry("k8s.gcr.io") + "/pause:" + out

	return err
}

func generateKubeadm(s *state.State) error {
	s.Logger.Infoln("Generating kubeadm config file...")

	if err := determinePauseImage(s); err != nil {
		return errors.Wrap(err, "failed to determine pause image")
	}

	kubeadmProvider, err := kubeadm.New(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "failed to init kubeadm")
	}

	for idx := range s.Cluster.ControlPlane.Hosts {
		node := s.Cluster.ControlPlane.Hosts[idx]
		kubeadmConf, err := kubeadmProvider.Config(s, node)
		if err != nil {
			return errors.Wrap(err, "failed to create kubeadm configuration")
		}

		s.Configuration.AddFile(fmt.Sprintf("cfg/master_%d.yaml", node.ID), kubeadmConf)
	}

	for idx := range s.Cluster.StaticWorkers.Hosts {
		node := s.Cluster.StaticWorkers.Hosts[idx]
		kubeadmConf, err := kubeadmProvider.ConfigWorker(s, node)
		if err != nil {
			return errors.Wrap(err, "failed to create kubeadm configuration")
		}

		s.Configuration.AddFile(fmt.Sprintf("cfg/worker_%d.yaml", node.ID), kubeadmConf)
	}

	return s.RunTaskOnAllNodes(uploadKubeadmToNode, state.RunParallel)
}

func uploadKubeadmToNode(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
	return errors.Wrap(s.Configuration.UploadTo(conn, s.WorkDir), "failed to upload")
}
