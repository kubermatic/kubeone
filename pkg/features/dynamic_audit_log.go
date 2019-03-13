package features

import (
	kubeadmv1beta1 "github.com/kubermatic/kubeone/pkg/apis/kubeadm/v1beta1"
)

const (
	auditDynamicConfigurationFlag = "audit-dynamic-configuration"
	runtimeConfigFlag             = "runtime-config"
	auditRegistrationAPI          = "auditregistration.k8s.io/v1alpha1=true"
)

func activateKubeadmDynamicAuditLogs(activate bool, clusterConfig *kubeadmv1beta1.ClusterConfiguration) {
	if !activate {
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
