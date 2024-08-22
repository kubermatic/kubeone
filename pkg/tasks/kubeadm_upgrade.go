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
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/kubeadm"
)

func upgradeLeaderControlPlane(s *state.State, nodeID int) error {
	kadm, err := kubeadm.New(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	cmd, err := scripts.KubeadmUpgrade(kadm.UpgradeLeaderCommand(), s.WorkDir, true, nodeID)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return fail.SSH(err, "running kubeadm upgrade on control plane leader")
}

func upgradeFollowerControlPlane(s *state.State, nodeID int) error {
	kadm, err := kubeadm.New(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	cmd, err := scripts.KubeadmUpgrade(kadm.UpgradeFollowerCommand(), s.WorkDir, false, nodeID)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return fail.SSH(err, "running kubeadm upgrade on control plane follower")
}

func upgradeStaticWorker(s *state.State) error {
	kadm, err := kubeadm.New(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.Run(`sudo `+kadm.UpgradeStaticWorkerCommand(), nil)

	return fail.SSH(err, "running kubeadm upgrade on static worker")
}
