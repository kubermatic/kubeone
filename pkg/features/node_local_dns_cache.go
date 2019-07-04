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

package features

import (
	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/templates/dnscache"
	"github.com/kubermatic/kubeone/pkg/templates/kubelet"
)

func installNodeLocalDNSCache(nodelocalcache *kubeoneapi.NodeLocalDNSCache, s *state.State) error {
	if nodelocalcache == nil {
		return nil
	}

	if !nodelocalcache.Enable {
		return nil
	}

	s.Logger.Info("deploying node local DNS cache…")

	if err := dnscache.Deploy(s); err != nil {
		return errors.Wrap(err, "failed to deploy node local DNS cache")
	}

	konfig, err := kubelet.GetConfig(s)
	if err != nil {
		return errors.Wrap(err, "failed to get kubelet config")
	}
	// override what was there, we don't care anymore
	konfig.ClusterDNS = []string{dnscache.VirtualIP}

	if err = kubelet.SaveConfig(s, konfig); err != nil {
		return errors.Wrap(err, "failed to ensure kubelet flags")
	}

	return errors.Wrap(s.RunTaskOnAllNodes(kubelet.DeployConfig, false), "failed to ")
}
