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

	"github.com/kubermatic/kubeone/pkg/apis/kubeone"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateKubeOneCluster validates the KubeOneCluster object
func ValidateKubeOneCluster(c kubeone.KubeOneCluster) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateProviderConfig(c.Spec.Provider, field.NewPath("provider"))...)

	if len(c.Spec.Hosts) > 0 {
		allErrs = append(allErrs, ValidateHostConfig(c.Spec.Hosts, field.NewPath("hosts"))...)
	} else {
		allErrs = append(allErrs, field.Invalid(field.NewPath("hosts"), c.Spec.Hosts, "no host specified"))
	}

	if *c.Spec.MachineController.Deploy {
		allErrs = append(allErrs, ValidateMachineControllerConfig(c.Spec.MachineController, c.Spec.Provider.Name, field.NewPath("machineController"))...)
		allErrs = append(allErrs, ValidateWorkerConfig(c.Spec.Workers, field.NewPath("workers"))...)
	} else if len(c.Spec.Workers) > 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("workers"), c.Spec.Workers, "machine-controller deployment is disabled, but configuration still contains worker definitions"))
	}

	allErrs = append(allErrs, ValidateClusterNetworkConfig(c.Spec.ClusterNetwork, field.NewPath("clusterNetwork"))...)
	allErrs = append(allErrs, ValidateFeatures(c.Spec.Features, field.NewPath("features"))...)

	return allErrs
}

// ValidateProviderConfig checks the ProviderConfig for errors
func ValidateProviderConfig(p kubeone.ProviderConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	switch p.Name {
	case kubeone.ProviderNameAWS:
	case kubeone.ProviderNameOpenStack:
		if p.CloudConfig == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, p.CloudConfig, "`provider.cloud_config` is required for openstack provider"))
		}
	case kubeone.ProviderNameHetzner:
	case kubeone.ProviderNameDigitalOcean:
	case kubeone.ProviderNameVSphere:
	case kubeone.ProviderNameGCE:
	case kubeone.ProviderNameNone:
	default:
		allErrs = append(allErrs, field.Invalid(fldPath, p.Name, "unknown provider name"))
	}

	return allErrs
}

// TODO(xmudrii): hosts == 0

// ValidateHostConfig validates the HostConfig structure
func ValidateHostConfig(hosts []*kubeone.HostConfig, fldPath *field.Path) field.ErrorList {
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

// ValidateMachineControllerConfig validates the MachineControllerConfig structure
func ValidateMachineControllerConfig(m kubeone.MachineControllerConfig, cloudProviderName kubeone.ProviderName, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// If ProviderName is not None default to cloud provider and ensure user have not
	// manually provided machine-controller provider different than cloud provider.
	// If ProviderName is None, take user input or default to None.
	if cloudProviderName != kubeone.ProviderNameNone {
		if m.Provider != cloudProviderName {
			allErrs = append(allErrs, field.Invalid(fldPath, m.Provider, "cloud provider must be same as machine-controller provider"))
		}
	} else if cloudProviderName == kubeone.ProviderNameNone && m.Provider == "" {
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

	return allErrs
}

// ValidateFeatures validates the Features structure
func ValidateFeatures(f kubeone.Features, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if f.OpenIDConnect != nil && f.OpenIDConnect.Enable {
		allErrs = append(allErrs, ValidateOIDCConfig(f.OpenIDConnect.Config, fldPath.Child("openidConnect"))...)
	}
	allErrs = append(allErrs)
	return allErrs
}

// ValidateOIDCConfig validates the OpenID Connect configuration
func ValidateOIDCConfig(o *kubeone.OpenIDConnectConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if o.IssuerURL == "" {
		allErrs = append(allErrs, field.Invalid(fldPath, o.IssuerURL, "openid_connect.config.issuer_url can't be empty"))
	}
	if o.ClientID == "" {
		allErrs = append(allErrs, field.Invalid(fldPath, o.ClientID, "openid_connect.config.client_id can't be empty"))
	}

	return allErrs
}
