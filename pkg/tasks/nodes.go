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
	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
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

func uncordonNode(s *state.State, host kubeoneapi.HostConfig) error {
	updateErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var node corev1.Node

		if err := s.DynamicClient.Get(s.Context, types.NamespacedName{Name: host.Hostname}, &node); err != nil {
			return err
		}

		node.Spec.Unschedulable = false
		return s.DynamicClient.Update(s.Context, &node)
	})

	return errors.WithStack(updateErr)
}

func restartKubeAPIServer(s *state.State) error {
	s.Logger.Infoln("Restarting unhealthy API servers if needed...")
	return s.RunTaskOnControlPlane(func(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
		_, _, err := s.Runner.Run(scripts.RestartKubeAPIServer(), nil)
		if err != nil {
			return err
		}

		return nil
	}, state.RunSequentially)
}
