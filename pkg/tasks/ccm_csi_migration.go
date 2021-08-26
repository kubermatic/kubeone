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
	"time"

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
)

func validateExternalCloudProviderConfig(s *state.State) error {
	if s.LiveCluster.CCMStatus != nil && s.LiveCluster.CCMStatus.ExternalCCMDeployed &&
		!s.LiveCluster.CCMStatus.InTreeCloudProviderEnabled {
		return errors.New("the cluster is already running external ccm")
	}
	if s.Cluster.CloudProvider.Openstack == nil {
		return errors.New("ccm/csi migration is currently supported only for openstack")
	}
	if !s.Cluster.CloudProvider.External {
		return errors.New(".cloudProvider.external must be enabled to start the migration")
	}

	return nil
}

func regenerateStaticPodManifests(s *state.State) error {
	return s.RunTaskOnControlPlane(regenerateStaticPodManifestsInternal, state.RunSequentially)
}

func regenerateStaticPodManifestsInternal(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	logger := s.Logger.WithField("node", node.PublicAddress)

	logger.Info("Regenerating Kubernetes API server and controller-manager manifests...")
	cmd, err := scripts.CCMMigrationRegenerateStaticPodManifests(s.WorkDir, node.ID, s.KubeadmVerboseFlag())
	if err != nil {
		return err
	}
	_, _, err = s.Runner.RunRaw(cmd)
	if err != nil {
		return err
	}

	sleepTime := 30 * time.Second
	logger.Infof("Waiting %s to ensure main control plane components are up...", sleepTime)
	time.Sleep(sleepTime)

	return nil
}
