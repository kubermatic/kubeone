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
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
)

func drainNode(s *state.State, node kubeoneapi.HostConfig) error {
	cmd, err := scripts.DrainNode(node.Hostname)
	if err != nil {
		return err
	}

	return s.RunTaskOnLeader(func(s *state.State, _ *kubeoneapi.HostConfig, _ ssh.Connection) error {
		_, _, err := s.Runner.RunRaw(cmd)

		return err
	})
}

func uncordonNode(s *state.State, node kubeoneapi.HostConfig) error {
	cmd, err := scripts.UncordonNode(node.Hostname)
	if err != nil {
		return err
	}

	return s.RunTaskOnLeader(func(s *state.State, _ *kubeoneapi.HostConfig, _ ssh.Connection) error {
		_, _, err := s.Runner.RunRaw(cmd)

		return err
	})
}
