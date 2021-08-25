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
	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/state"
)

func validateExternalCloudProviderConfig(s *state.State) error {
	if s.LiveCluster.CCMMigration != nil && s.LiveCluster.CCMMigration.ExternalCCMDeployed &&
		!s.LiveCluster.CCMMigration.InTreeCloudProviderEnabled {
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
