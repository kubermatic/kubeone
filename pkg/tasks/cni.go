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

	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/canal"
	"k8c.io/kubeone/pkg/templates/weave"
)

func ensureCNI(s *state.State) error {
	switch {
	case s.Cluster.ClusterNetwork.CNI.Canal != nil:
		return ensureCNICanal(s)
	case s.Cluster.ClusterNetwork.CNI.WeaveNet != nil:
		return ensureCNIWeaveNet(s)
	case s.Cluster.ClusterNetwork.CNI.External != nil:
		return ensureCNIExternal(s)
	}

	return errors.Errorf("unknown CNI provider")
}

func ensureCNIWeaveNet(s *state.State) error {
	s.Logger.Infoln("Applying weave-net CNI plugin…")
	return weave.Deploy(s)
}

func ensureCNICanal(s *state.State) error {
	s.Logger.Infoln("Applying canal CNI plugin…")
	return canal.Deploy(s)
}

func ensureCNIExternal(s *state.State) error {
	s.Logger.Infoln("External CNI plugin will be used")
	return nil
}

func patchCNI(s *state.State) error {
	if !s.PatchCNI {
		return nil
	}

	s.Logger.Info("Patching CNI…")
	return ensureCNI(s)
}
