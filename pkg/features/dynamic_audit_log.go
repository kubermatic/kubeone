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
	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/kubeadmargs"
)

const (
	auditDynamicConfigurationFlag = "audit-dynamic-configuration"
	runtimeConfigFlag             = "runtime-config"
	auditRegistrationAPI          = "auditregistration.k8s.io/v1alpha1=true"
)

func updateDynamicAuditLogsKubeadmConfig(feature *kubeoneapi.DynamicAuditLog, args *kubeadmargs.Args) {
	if feature == nil {
		return
	}

	if !feature.Enable {
		return
	}

	args.APIServer.ExtraArgs[auditDynamicConfigurationFlag] = "true"
	args.APIServer.AppendMapStringStringExtraArg(runtimeConfigFlag, auditRegistrationAPI)
	args.FeatureGates["DynamicAuditing"] = true
}
