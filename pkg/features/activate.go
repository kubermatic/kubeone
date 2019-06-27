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
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/kubeadmargs"
	"github.com/kubermatic/kubeone/pkg/util"
)

// Activate configured features.
// Installing CRDs, creating policies and so on
func Activate(ctx *util.Context) error {
	if err := installKubeSystemPSP(ctx.Cluster.Features.PodSecurityPolicy, ctx); err != nil {
		return errors.Wrap(err, "failed to install PodSecurityPolicy")
	}

	if err := installMetricsServer(ctx.Cluster.Features.MetricsServer, ctx); err != nil {
		return errors.Wrap(err, "failed to install metrics-server")
	}

	if err := installNodeLocalDNSCache(ctx.Cluster.Features.NodeLocalDNSCache, ctx); err != nil {
		return errors.Wrap(err, "failed to install nodeLocal DNSCache")
	}

	return nil
}

// UpdateKubeadmClusterConfiguration update additional config options in the kubeadm's
// v1beta1.ClusterConfiguration according to enabled features
func UpdateKubeadmClusterConfiguration(featuresCfg kubeoneapi.Features, args *kubeadmargs.Args) {
	activateKubeadmPSP(featuresCfg.PodSecurityPolicy, args)
	activateKubeadmDynamicAuditLogs(featuresCfg.DynamicAuditLog, args)
	activateKubeadmOIDC(featuresCfg.OpenIDConnect, args)
}
