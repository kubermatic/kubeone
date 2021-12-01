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

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/kubeadm/kubeadmargs"
)

const (
	apiServerAdmissionPluginsFlag               = "enable-admission-plugins"
	apiServerAdmissionControlConfigFlag         = "admission-control-config-file"
	apiServerAdmissionControlConfigPath         = "/etc/kubernetes/admission/admission-config.yaml"
	apiServerEncryptionProviderFlag             = "encryption-provider-config"
	apiServerEncryptionProviderConfigPath       = "/etc/kubernetes/encryption-providers/encryption-providers.yaml"
	apiServerEncryptionProviderCustomConfigPath = "/etc/kubernetes/encryption-providers/custom-encryption-providers.yaml"
)

// Activate configured features.
// Installing CRDs, creating policies and so on
func Activate(s *state.State) error {
	s.Logger.Info("Activating additional features...")

	if err := installKubeSystemPSP(s.Cluster.Features.PodSecurityPolicy, s); err != nil {
		return errors.Wrap(err, "failed to install PodSecurityPolicy")
	}

	if err := installPodNodeSelector(s.Context, s.DynamicClient, s.Cluster.Features.PodNodeSelector); err != nil {
		return errors.Wrap(err, "failed to install podNodeSelector")
	}

	return nil
}

// UpdateKubeadmClusterConfiguration update additional config options in the kubeadm's
// v1beta1.ClusterConfiguration according to enabled features
func UpdateKubeadmClusterConfiguration(featuresCfg kubeoneapi.Features, args *kubeadmargs.Args) {
	activateKubeadmPSP(featuresCfg.PodSecurityPolicy, args)
	activateKubeadmStaticAuditLogs(featuresCfg.StaticAuditLog, args)
	activateKubeadmDynamicAuditLogs(featuresCfg.DynamicAuditLog, args)
	activateKubeadmOIDC(featuresCfg.OpenIDConnect, args)
	activateKubeadmPodNodeSelector(featuresCfg.PodNodeSelector, args)
	activateEncryptionProviders(featuresCfg.EncryptionProviders, args)
}
