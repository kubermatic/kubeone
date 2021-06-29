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

	"k8c.io/kubeone/pkg/addons"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"
	"k8c.io/kubeone/pkg/templates/weave"
)

func ensureCNI(s *state.State) error {
	switch {
	case s.Cluster.ClusterNetwork.CNI.Canal != nil:
		if err := addons.EnsureAddonByName(s, resources.AddonCNICanal); err != nil {
			return err
		}
	case s.Cluster.ClusterNetwork.CNI.WeaveNet != nil:
		if s.Cluster.ClusterNetwork.CNI.WeaveNet.Encrypted {
			if err := weave.EnsureSecret(s); err != nil {
				return err
			}
		}
		if err := addons.EnsureAddonByName(s, resources.AddonCNIWeavenet); err != nil {
			return err
		}
	case s.Cluster.ClusterNetwork.CNI.External != nil:
		s.Logger.Infoln("External CNI plugin will be used")
	default:
		return errors.Errorf("unknown CNI provider")
	}

	return kubeconfig.HackIssue321InitDynamicClient(s)
}
