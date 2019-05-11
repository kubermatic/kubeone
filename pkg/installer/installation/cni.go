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
	"github.com/kubermatic/kubeone/pkg/templates/canal"
	"github.com/kubermatic/kubeone/pkg/templates/weave"
	"github.com/kubermatic/kubeone/pkg/util"
)

func ensureCNI(ctx *util.Context) error {
	switch ctx.Cluster.ClusterNetwork.CNI.Provider {
	case kubeone.CNIProviderCanal:
		return ensureCNICanal(ctx)
	case kubeone.CNIProviderWeaveNet:
		return ensureCNIWeaveNet(ctx)
	}

	return errors.Errorf("unknown CNI provider: %s", ctx.Cluster.ClusterNetwork.CNI.Provider)
}

func ensureCNIWeaveNet(ctx *util.Context) error {
	ctx.Logger.Infoln("Applying weave-net CNI plugin…")
	return weave.Deploy(ctx)
}

func ensureCNICanal(ctx *util.Context) error {
	ctx.Logger.Infoln("Applying canal CNI plugin…")
	return canal.Deploy(ctx)
}
