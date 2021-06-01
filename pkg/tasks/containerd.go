/*
Copyright 2021 The KubeOne Authors.

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
	"errors"

	"k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
)

func validateContainerdInConfig(s *state.State) error {
	if s.Cluster.ContainerRuntime.Containerd == nil {
		return errors.New("containerd must be enabled in config")
	}

	return nil
}

func migrateToContainerd(s *state.State) error {
	return s.RunTaskOnAllNodes(migrateToContainerdTask, state.RunSequentially)
}

func migrateToContainerdTask(s *state.State, node *kubeone.HostConfig, conn ssh.Connection) error {
	s.Logger.Info("Migrating container runtime to containerd")

	migrateScript, err := scripts.MigrateToContainerd(s.Cluster.RegistryConfiguration.InsecureRegistryAddress())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(migrateScript)

	return err
}
