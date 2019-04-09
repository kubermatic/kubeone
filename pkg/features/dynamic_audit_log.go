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
	kubeadmv1beta1 "github.com/kubermatic/kubeone/pkg/apis/kubeadm/v1beta1"
	"github.com/kubermatic/kubeone/pkg/config"
)

const (
	auditDynamicConfigurationFlag = "audit-dynamic-configuration"
	runtimeConfigFlag             = "runtime-config"
	auditRegistrationAPI          = "auditregistration.k8s.io/v1alpha1=true"
)

func activateKubeadmDynamicAuditLogs(feature config.DynamicAuditLog, clusterConfig *kubeadmv1beta1.ClusterConfiguration) {
	if !feature.Enable {
		return
	}

	if clusterConfig.APIServer.ExtraArgs == nil {
		clusterConfig.APIServer.ExtraArgs = make(map[string]string)
	}

	clusterConfig.APIServer.ExtraArgs[auditDynamicConfigurationFlag] = "true"

	if _, ok := clusterConfig.APIServer.ExtraArgs[runtimeConfigFlag]; ok {
		clusterConfig.APIServer.ExtraArgs[runtimeConfigFlag] += "," + auditRegistrationAPI
	} else {
		clusterConfig.APIServer.ExtraArgs[runtimeConfigFlag] = auditRegistrationAPI
	}

	if clusterConfig.FeatureGates == nil {
		clusterConfig.FeatureGates = map[string]bool{}
	}
	clusterConfig.FeatureGates["DynamicAuditing"] = true
}
