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

package installation

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/templates/canal"
	"github.com/kubermatic/kubeone/pkg/templates/weave"
)

func ensureCNI(s *state.State) error {
	switch s.Cluster.ClusterNetwork.CNI.Provider {
	case kubeone.CNIProviderCanal:
		return ensureCNICanal(s)
	case kubeone.CNIProviderWeaveNet:
		return ensureCNIWeaveNet(s)
	}

	return errors.Errorf("unknown CNI provider: %s", s.Cluster.ClusterNetwork.CNI.Provider)
}

func ensureCNIWeaveNet(s *state.State) error {
	s.Logger.Infoln("Applying weave-net CNI plugin…")
	return weave.Deploy(s)
}

func ensureCNICanal(s *state.State) error {
	s.Logger.Infoln("Applying canal CNI plugin…")
	return canal.Deploy(s)
}
