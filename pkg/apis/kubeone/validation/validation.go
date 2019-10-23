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

package validation

import (
	"net"

	"github.com/Masterminds/semver"

	"github.com/kubermatic/kubeone/pkg/apis/kubeone"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateKubeOneCluster validates the KubeOneCluster object
func ValidateKubeOneCluster(c kubeone.KubeOneCluster) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateCloudProviderSpec(c.CloudProvider, field.NewPath("provider"))...)

	if c.Name == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("name"), c.Name, "no cluster name specified"))
	}
	if len(c.Hosts) > 0 {
		allErrs = append(allErrs, ValidateHostConfig(c.Hosts, field.NewPath("hosts"))...)
	} else {
		allErrs = append(allErrs, field.Invalid(field.NewPath("hosts"), c.Hosts, "no host specified"))
	}

	if c.MachineController != nil && c.MachineController.Deploy {
		allErrs = append(allErrs, ValidateMachineControllerConfig(c.MachineController, c.CloudProvider.Name, field.NewPath("machineController"))...)
		allErrs = append(allErrs, ValidateWorkerConfig(c.Workers, field.NewPath("workers"))...)
	} else if len(c.Workers) > 0 {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("workers"), "machine-controller deployment is disabled, but configuration still contains worker definitions"))
	}

	allErrs = append(allErrs, ValidateVersionConfig(c.Versions, field.NewPath("versions"))...)
	allErrs = append(allErrs, ValidateClusterNetworkConfig(c.ClusterNetwork, field.NewPath("clusterNetwork"))...)
	allErrs = append(allErrs, ValidateFeatures(c.Features, field.NewPath("features"))...)

	return allErrs
}

// ValidateCloudProviderSpec checks the CloudProviderSpec structure for errors
func ValidateCloudProviderSpec(p kubeone.CloudProviderSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	switch p.Name {
	case kubeone.CloudProviderNameAWS:
	case kubeone.CloudProviderNameAzure:
		if p.CloudConfig == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, p.CloudConfig, "`cloudProvider.cloudConfig` is required for azure provider"))
		}
	case kubeone.CloudProviderNameOpenStack:
		if p.CloudConfig == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, p.CloudConfig, "`cloudProvider.cloudConfig` is required for openstack provider"))
		}
	case kubeone.CloudProviderNameHetzner:
	case kubeone.CloudProviderNameDigitalOcean:
	case kubeone.CloudProviderNamePacket:
	case kubeone.CloudProviderNameVSphere:
		if p.CloudConfig == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, p.CloudConfig, "`cloudProvider.cloudConfig` is required for vsphere provider"))
		}
	case kubeone.CloudProviderNameGCE:
	case kubeone.CloudProviderNameNone:
	default:
		allErrs = append(allErrs, field.Invalid(fldPath, p.Name, "unknown provider name"))
	}

	return allErrs
}

// ValidateHostConfig validates the HostConfig structure
func ValidateHostConfig(hosts []kubeone.HostConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for _, h := range hosts {
		if len(h.PublicAddress) == 0 {
			allErrs = append(allErrs, field.Invalid(fldPath, h.PublicAddress, "no public IP/address given"))
		}
		if len(h.PrivateAddress) == 0 {
			allErrs = append(allErrs, field.Invalid(fldPath, h.PrivateAddress, "no private IP/address givevn"))
		}
		if len(h.SSHPrivateKeyFile) == 0 && len(h.SSHAgentSocket) == 0 {
			allErrs = append(allErrs, field.Invalid(fldPath, h.SSHPrivateKeyFile, "neither SSH private key nor agent socket given, don't know how to authenticate"))
			allErrs = append(allErrs, field.Invalid(fldPath, h.SSHAgentSocket, "neither SSH private key nor agent socket given, don't know how to authenticate"))
		}
		if len(h.SSHUsername) == 0 {
			allErrs = append(allErrs, field.Invalid(fldPath, h.SSHUsername, "no SSH username given"))
		}
	}

	return allErrs
}

// ValidateVersionConfig validates the VersionConfig structure
func ValidateVersionConfig(version kubeone.VersionConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	v, err := semver.NewVersion(version.Kubernetes)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, version, "failed to parse kubernetes version"))
		return allErrs
	}

	if v.Major() != 1 || v.Minor() < 13 {
		allErrs = append(allErrs, field.Invalid(fldPath, version, "kubernetes versions lower than 1.13 are not supported"))
	}

	return allErrs
}

// ValidateMachineControllerConfig validates the MachineControllerConfig structure
func ValidateMachineControllerConfig(m *kubeone.MachineControllerConfig, cloudProviderName kubeone.CloudProviderName, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// If ProviderName is not None default to cloud provider and ensure user have not
	// manually provided machine-controller provider different than cloud provider.
	// If ProviderName is None, take user input or default to None.
	if cloudProviderName != kubeone.CloudProviderNameNone {
		if m.Provider != cloudProviderName {
			allErrs = append(allErrs, field.Invalid(fldPath, m.Provider, "cloud provider must be same as machine-controller provider"))
		}
	} else if cloudProviderName == kubeone.CloudProviderNameNone && m.Provider == "" {
		allErrs = append(allErrs, field.Invalid(fldPath, m.Provider, "machine-controller deployed but no provider selected"))
	}

	return allErrs
}

// ValidateWorkerConfig validates the WorkerConfig structure
func ValidateWorkerConfig(workerset []kubeone.WorkerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for _, w := range workerset {
		if w.Name == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, w.Name, "no name given"))
		}
		if w.Replicas == nil || *w.Replicas < 1 {
			allErrs = append(allErrs, field.Invalid(fldPath, w.Replicas, "replicas must be specified and >= 1"))
		}
	}

	return allErrs
}

// ValidateClusterNetworkConfig validates the ClusterNetworkConfig structure
func ValidateClusterNetworkConfig(c kubeone.ClusterNetworkConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if c.PodSubnet != "" {
		if _, _, err := net.ParseCIDR(c.PodSubnet); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath, c.PodSubnet, "invalid pod subnet specified"))
		}
	}

	if c.ServiceSubnet != "" {
		if _, _, err := net.ParseCIDR(c.ServiceSubnet); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath, c.ServiceSubnet, "invalid service subnet specified"))
		}
	}

	if c.CNI != nil {
		allErrs = append(allErrs, ValidateCNI(c.CNI, fldPath.Child("cni"))...)
	}

	return allErrs
}

// ValidateCNI validates CNI structure
func ValidateCNI(c *kubeone.CNI, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	switch c.Provider {
	case kubeone.CNIProviderCanal:
	case kubeone.CNIProviderWeaveNet:
	default:
		allErrs = append(allErrs, field.Invalid(fldPath, c.Provider, "unknown CNI provider"))
	}

	if c.Encrypted && c.Provider != kubeone.CNIProviderWeaveNet {
		allErrs = append(allErrs, field.Invalid(fldPath, c, "only `weave-net` cni provider support `encrypted: true`"))
	}

	return allErrs
}

// ValidateFeatures validates the Features structure
func ValidateFeatures(f kubeone.Features, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if f.StaticAuditLog != nil && f.StaticAuditLog.Enable {
		allErrs = append(allErrs, ValidateStaticAuditLogConfig(f.StaticAuditLog.Config, fldPath.Child("staticAuditLog"))...)
	}
	if f.OpenIDConnect != nil && f.OpenIDConnect.Enable {
		allErrs = append(allErrs, ValidateOIDCConfig(f.OpenIDConnect.Config, fldPath.Child("openidConnect"))...)
	}
	return allErrs
}

func ValidateStaticAuditLogConfig(s kubeone.StaticAuditLogConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if s.PolicyFilePath == "" {
		allErrs = append(allErrs, field.Invalid(fldPath, s.PolicyFilePath, "staticAuditLog.config.policyFilePath can't be empty"))
	}
	if s.LogPath == "" {
		allErrs = append(allErrs, field.Invalid(fldPath, s.LogPath, "staticAuditLog.config.logPath can't be empty"))
	}
	if s.LogMaxAge <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, s.LogMaxAge, "staticAuditLog.config.logMaxAge must be greater than 0"))
	}
	if s.LogMaxBackup <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, s.LogMaxBackup, "staticAuditLog.config.logMaxBackup must be greater than 0"))
	}
	if s.LogMaxSize <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, s.LogMaxSize, "staticAuditLog.config.logMaxSize must be greater than 0"))
	}

	return allErrs
}

// ValidateOIDCConfig validates the OpenID Connect configuration
func ValidateOIDCConfig(o kubeone.OpenIDConnectConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if o.IssuerURL == "" {
		allErrs = append(allErrs, field.Invalid(fldPath, o.IssuerURL, "openidConnect.config.issuer_url can't be empty"))
	}
	if o.ClientID == "" {
		allErrs = append(allErrs, field.Invalid(fldPath, o.ClientID, "openidConnect.config.client_id can't be empty"))
	}

	return allErrs
}
