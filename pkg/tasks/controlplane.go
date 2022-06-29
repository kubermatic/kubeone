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
	"time"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/state"
)

const kubeadmPhaseKubeProxy = "addon/kube-proxy"

func joinControlplaneNode(s *state.State) error {
	s.Logger.Infoln("Joining controlplane node...")

	return s.RunTaskOnFollowers(joinControlPlaneNodeInternal, state.RunSequentially)
}

func joinControlPlaneNodeInternal(s *state.State, node *kubeoneapi.HostConfig, conn executor.Interface) error {
	logger := s.Logger.WithField("node", node.PublicAddress)

	sleepTime := 15 * time.Second
	logger.Infof("Waiting %s to ensure main control plane components are up...", sleepTime)
	time.Sleep(sleepTime)

	logger.Info("Joining control plane node")
	cmd, err := scripts.KubeadmJoin(s.WorkDir, node.ID, s.KubeadmVerboseFlag())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	if err != nil {
		return fail.SSH(err, "joining control plane node %q", node.PublicAddress)
	}

	return approvePendingCSR(s, node, conn)
}

func kubeadmCertsExecutor(s *state.State, node *kubeoneapi.HostConfig, conn executor.Interface) error {
	s.Logger.Infoln("Ensuring Certificates...")
	cmd, err := scripts.KubeadmCert(s.WorkDir, node.ID, s.KubeadmVerboseFlag())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return fail.SSH(err, "generating kubeadm certificates")
}

func initKubernetesLeader(s *state.State) error {
	s.Logger.Infoln("Initializing Kubernetes on leader...")

	return s.RunTaskOnLeader(func(s *state.State, node *kubeoneapi.HostConfig, conn executor.Interface) error {
		var skipPhase string
		if s.Cluster.ClusterNetwork.KubeProxy != nil && s.Cluster.ClusterNetwork.KubeProxy.SkipInstallation {
			skipPhase = kubeadmPhaseKubeProxy
		}

		s.Logger.Infoln("Running kubeadm...")

		cmd, err := scripts.KubeadmInit(s.WorkDir, node.ID, s.KubeadmVerboseFlag(), s.JoinToken, time.Hour.String(), skipPhase)
		if err != nil {
			return err
		}

		_, _, err = s.Runner.RunRaw(cmd)

		return fail.SSH(err, "initializing kubeadm cluster")
	})
}
