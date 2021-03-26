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
	"k8s.io/apimachinery/pkg/util/sets"
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

	return s.RunTaskOnControlPlane(func(s *state.State, node *kubeoneapi.HostConfig, _ ssh.Connection) error {
		return restartKubeAPIServerOnOS(s, *node)
	}, state.RunSequentially)
}

func restartKubeAPIServerOnOS(s *state.State, node kubeoneapi.HostConfig) error {
	return runOnOS(s, node.OperatingSystem, map[kubeoneapi.OperatingSystemName]runOnOSFn{
		kubeoneapi.OperatingSystemNameAmazon:  restartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameCentOS:  restartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameDebian:  restartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameFlatcar: restartKubeAPIServerDocker,
		kubeoneapi.OperatingSystemNameRHEL:    restartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameUbuntu:  restartKubeAPIServerCrictl,
	})
}

func restartKubeAPIServerCrictl(s *state.State) error {
	_, _, err := s.Runner.RunRaw(scripts.RestartKubeAPIServerCrictl())

	return errors.WithStack(err)
}

func restartKubeAPIServerDocker(s *state.State) error {
	_, _, err := s.Runner.RunRaw(scripts.RestartKubeAPIServerDocker())

	return errors.WithStack(err)
}

func labelNodeOSes(s *state.State) error {
	candidateNodes := sets.NewString()
	nodeList := corev1.NodeList{}

	if err := s.DynamicClient.List(s.Context, &nodeList); err != nil {
		return err
	}

	for _, node := range nodeList.Items {
		candidateNodes.Insert(node.Name)
		for _, addr := range node.Status.Addresses {
			candidateNodes.Insert(addr.Address)
		}
	}

	hostsSet := map[string]kubeoneapi.HostConfig{}

	for _, host := range append(s.Cluster.ControlPlane.Hosts, s.Cluster.StaticWorkers.Hosts...) {
		if candidateNodes.Has(host.Hostname) || candidateNodes.Has(host.PrivateAddress) || candidateNodes.Has(host.PublicAddress) {
			hostsSet[host.Hostname] = host
		}
	}

	for nodeName, host := range hostsSet {
		nodeName := nodeName
		host := host
		updateErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			var node corev1.Node

			if err := s.DynamicClient.Get(s.Context, types.NamespacedName{Name: nodeName}, &node); err != nil {
				return err
			}

			node.Labels["v1.kubeone.io/operating-system"] = string(host.OperatingSystem)
			return s.DynamicClient.Update(s.Context, &node)
		})

		if updateErr != nil {
			return updateErr
		}
	}

	return nil
}
