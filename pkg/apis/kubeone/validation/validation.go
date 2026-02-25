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
	"bytes"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"

	"github.com/Masterminds/semver/v3"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	helm "k8c.io/kubeone/pkg/localhelm"
	"k8c.io/kubeone/pkg/semverutil"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	netutils "k8s.io/utils/net"
	"sigs.k8s.io/yaml"
)

const (
	// MinimumSupportedVersion defines the minimum Kubernetes version supported by KubeOne.
	MinimumSupportedVersion = "1.33"
	// MaximumSupportedVersion defines the maximum Kubernetes version supported by KubeOne.
	MaximumSupportedVersion = "1.35"
)

var (
	// minVersionConstraint defines the minimum Kubernetes version supported by KubeOne
	minVersionConstraint = semverutil.MustParseConstraint(fmt.Sprintf(">= %s", MinimumSupportedVersion))
	// maxVersionConstraint defines the maximum Kubernetes version supported by KubeOne
	maxVersionConstraint = semverutil.MustParseConstraint(fmt.Sprintf("<= %s", MaximumSupportedVersion))
)

// ValidateKubeOneCluster validates the KubeOneCluster object
func ValidateKubeOneCluster(c kubeoneapi.KubeOneCluster) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateName(c.Name, field.NewPath("name"))...)
	allErrs = append(allErrs, ValidateControlPlaneConfig(c.ControlPlane, c.ClusterNetwork, field.NewPath("controlPlane"))...)
	allErrs = append(allErrs, ValidateKubeletConfig(c.KubeletConfig, field.NewPath("kubeletConfig"))...)
	allErrs = append(allErrs, ValidateAPIEndpoint(c.APIEndpoint, field.NewPath("apiEndpoint"))...)
	allErrs = append(allErrs, ValidateCloudProviderSpec(c, field.NewPath("provider"))...)
	allErrs = append(allErrs, ValidateVersionConfig(c.Versions, field.NewPath("versions"))...)
	allErrs = append(allErrs, ValidateKubernetesSupport(c, field.NewPath(""))...)
	allErrs = append(allErrs, ValidateContainerRuntimeConfig(c.ContainerRuntime, c.Versions, field.NewPath("containerRuntime"))...)
	allErrs = append(allErrs, ValidateClusterNetworkConfig(c.ClusterNetwork, c.CloudProvider, field.NewPath("clusterNetwork"))...)
	allErrs = append(allErrs, ValidateStaticWorkersConfig(c.StaticWorkers, c.ControlPlane, c.ClusterNetwork, field.NewPath("staticWorkers"))...)

	if c.MachineController != nil && c.MachineController.Deploy {
		allErrs = append(allErrs, ValidateDynamicWorkerConfig(c.DynamicWorkers, c.CloudProvider, field.NewPath("dynamicWorkers"))...)
	} else if len(c.DynamicWorkers) > 0 {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("dynamicWorkers"),
			"machine-controller deployment is disabled, but the configuration still contains dynamic workers"))
	}

	if c.OperatingSystemManager.Deploy {
		allErrs = append(allErrs, ValidateOperatingSystemManager(c.MachineController, field.NewPath("operatingSystemManager"))...)
	}

	allErrs = append(allErrs, ValidateCABundle(c.CertificateAuthority.Bundle, field.NewPath("certificateAuthority", "bundle"))...)
	allErrs = append(allErrs, ValidateFeatures(c.Features, field.NewPath("features"))...)
	allErrs = append(allErrs, ValidateAddons(c.Addons, field.NewPath("addons"))...)
	allErrs = append(allErrs, ValidateRegistryConfiguration(c.RegistryConfiguration, field.NewPath("registryConfiguration"))...)
	allErrs = append(allErrs, ValidateControlPlaneComponents(c.ControlPlaneComponents, field.NewPath("controlPlaneComponents"))...)

	return allErrs
}

// ValidateName validates the Name of cluster
func ValidateName(name string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	errs := validation.IsDNS1123Subdomain(name)
	for _, err := range errs {
		allErrs = append(allErrs, field.Invalid(fldPath, name, err))
	}

	return allErrs
}

// ValidateControlPlaneConfig validates the ControlPlaneConfig structure
func ValidateControlPlaneConfig(c kubeoneapi.ControlPlaneConfig, clusterNetwork kubeoneapi.ClusterNetworkConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(c.Hosts) > 0 {
		allErrs = append(allErrs, ValidateHostConfig(c.Hosts, clusterNetwork, fldPath.Child("hosts"))...)
	} else {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("hosts"), "",
			".controlPlane.Hosts is a required field. There must be at least one control plane instance in the cluster."))
	}

	return allErrs
}

// ValidateAPIEndpoint validates the APIEndpoint structure
func ValidateAPIEndpoint(a kubeoneapi.APIEndpoint, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(a.Host) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("host"), ".apiEndpoint.host is a required field"))
	}
	if a.Port <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("port"), a.Port, "apiEndpoint.port must be greater than 0"))
	}
	if a.Port > 65535 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("port"), a.Port, "apiEndpoint.Port must be lower than 65535"))
	}

	visited := make(map[string]bool)
	for _, altName := range a.AlternativeNames {
		if visited[altName] {
			allErrs = append(allErrs, field.Invalid(fldPath, altName, "duplicates are not allowed in alternative names"))

			break
		}
		visited[altName] = true
	}

	return allErrs
}

// ValidateCloudProviderSpec validates the CloudProviderSpec structure
//
//nolint:gocyclo
func ValidateCloudProviderSpec(cluster kubeoneapi.KubeOneCluster, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	providerSpec := cluster.CloudProvider
	networkConfig := cluster.ClusterNetwork

	providerFound := false
	if providerSpec.AWS != nil {
		if networkConfig.IPFamily.IsDualstack() && providerSpec.External && len(providerSpec.CloudConfig) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("cloudConfig"), "cloudConfig is required for dualstack clusters for aws provider"))
		}
		providerFound = true
	}
	if providerSpec.Azure != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("azure"), "only one provider can be used at the same time"))
		}
		if len(providerSpec.CloudConfig) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("cloudConfig"), ".cloudProvider.cloudConfig is required for azure provider"))
		}
		providerFound = true
	}
	if providerSpec.DigitalOcean != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("digitalocean"), "only one provider can be used at the same time"))
		}
		providerFound = true
	}
	if providerSpec.GCE != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("gce"), "only one provider can be used at the same time"))
		}
		providerFound = true
	}
	if providerSpec.Hetzner != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("hetzner"), "only one provider can be used at the same time"))
		}
		providerFound = true
	}
	if providerSpec.Kubevirt != nil {
		kubevirtFld := fldPath.Child("kubevirt")
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(kubevirtFld, "only one provider can be used at the same time"))
		}
		providerFound = true
		if providerSpec.Kubevirt.InfraNamespace == "" {
			allErrs = append(allErrs, field.Required(kubevirtFld.Child("infraNamespace"), "is required for kubevirt provider"))
		}
	}
	if providerSpec.Nutanix != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("nutanix"), "only one provider can be used at the same time"))
		}
		providerFound = true
	}
	if providerSpec.Openstack != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("openstack"), "only one provider can be used at the same time"))
		}
		if len(providerSpec.CloudConfig) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("cloudConfig"), ".cloudProvider.cloudConfig is required for openstack provider"))
		}
		providerFound = true
	}
	if providerSpec.EquinixMetal != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("equinixmetal"), "only one provider can be used at the same time"))
		}
		providerFound = true
	}
	if providerSpec.VMwareCloudDirector != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("vmwareCloudDirector"), "only one provider can be used at the same time"))
		}
		providerFound = true
		if providerSpec.External {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("external"), "external cloud provider is not supported for VMware Cloud Director clusters"))
		}
	}
	if providerSpec.Vsphere != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("vsphere"), "only one provider can be used at the same time"))
		}
		if len(providerSpec.CloudConfig) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("cloudConfig"), ".cloudProvider.cloudConfig is required for vSphere provider"))
		}
		if providerSpec.External && !providerSpec.DisableBundledCSIDrivers && len(providerSpec.CSIConfig) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("csiConfig"), ".cloudProvider.csiConfig is required for vSphere provider"))
		}
		providerFound = true
	}
	if providerSpec.None != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("none"), "only one provider can be used at the same time"))
		}
		if cluster.MachineController != nil && cluster.MachineController.Deploy {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("none"), "machine-controller requires a cloud provider"))
		}
		if cluster.OperatingSystemManager != nil && cluster.OperatingSystemManager.Deploy {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("none"), "operating-system-manager requires a cloud provider"))
		}
		providerFound = true
	}

	if !providerFound {
		allErrs = append(allErrs, field.Invalid(fldPath, "", "provider must be specified"))
	}

	if providerSpec.DisableBundledCSIDrivers && len(providerSpec.CSIConfig) > 0 {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("csiConfig"), ".cloudProvider.csiConfig is mutually exclusive with .cloudProvider.disableBundledCSIDrivers"))
	}

	if providerSpec.Vsphere == nil && len(providerSpec.CSIConfig) > 0 {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("csiConfig"), ".cloudProvider.csiConfig is currently supported only for vsphere clusters"))
	}

	return allErrs
}

// ValidateVersionConfig validates the VersionConfig structure
func ValidateVersionConfig(version kubeoneapi.VersionConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	v, err := semver.NewVersion(version.Kubernetes)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("kubernetes"), version, ".versions.kubernetes is not a semver string"))

		return allErrs
	}

	if strings.HasPrefix(version.Kubernetes, "v") {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("kubernetes"), version, ".versions.kubernetes can't start with a leading 'v'"))
	}

	if valid, errs := minVersionConstraint.Validate(v); !valid {
		for _, err := range errs {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("kubernetes"), version, fmt.Sprintf("kubernetes version does not satisfy version constraint '%s': %s. You need to use an older KubeOne version to upgrade your cluster to a supported version. Please refer to the Compatibility section of docs for more details.", minVersionConstraint, err.Error())))
		}
	}

	if valid, errs := maxVersionConstraint.Validate(v); !valid {
		for _, err := range errs {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("kubernetes"), version, fmt.Sprintf("kubernetes version does not satisfy version constraint '%s': %s. This version is not yet supported. Please refer to the Compatibility section of docs for more details.", maxVersionConstraint, err.Error())))
		}
	}

	return allErrs
}

func ValidateKubernetesSupport(c kubeoneapi.KubeOneCluster, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if strings.Contains(c.Versions.Kubernetes, "-eks-") {
		return append(allErrs, field.Invalid(fldPath.Child("versions").Child("kubernetes"), c.Versions.Kubernetes, "Amazon EKS-D clusters are not supported by KubeOne 1.4+"))
	}

	v, err := semver.NewVersion(c.Versions.Kubernetes)
	if err != nil {
		return append(allErrs, field.Invalid(fldPath.Child("versions").Child("kubernetes"), c.Versions.Kubernetes, ".versions.kubernetes is not a semver string"))
	}

	if !c.CloudProvider.External {
		switch {
		case c.CloudProvider.AWS != nil && v.Minor() >= 27:
			// The in-tree cloud provider for AWS has been removed in Kubernetes 1.27.
			allErrs = append(allErrs, field.Invalid(fldPath.Child("cloudProvider").Child("external"), c.CloudProvider.External, "kubernetes 1.27 and newer doesn't support in-tree cloud provider with aws"))
		case c.CloudProvider.Azure != nil && v.Minor() >= 29:
			// The in-tree cloud provider for Azure has been removed in Kubernetes 1.29.
			allErrs = append(allErrs, field.Invalid(fldPath.Child("cloudProvider").Child("external"), c.CloudProvider.External, "kubernetes 1.29 and newer doesn't support in-tree cloud provider with azure"))
		case c.CloudProvider.GCE != nil && v.Minor() >= 29:
			// The in-tree cloud provider for GCE has been removed in Kubernetes 1.29.
			allErrs = append(allErrs, field.Invalid(fldPath.Child("cloudProvider").Child("external"), c.CloudProvider.External, "kubernetes 1.29 and newer doesn't support in-tree cloud provider with gce"))
		case c.CloudProvider.Openstack != nil && v.Minor() >= 26:
			// The in-tree cloud provider for OpenStack has been removed in Kubernetes 1.26.
			allErrs = append(allErrs, field.Invalid(fldPath.Child("cloudProvider").Child("external"), c.CloudProvider.External, "kubernetes 1.26 and newer doesn't support in-tree cloud provider with openstack"))
		case c.CloudProvider.Vsphere != nil && v.Minor() >= 25:
			// We require external CCM/CSI on vSphere starting with Kubernetes 1.25
			// because the in-tree volume plugin requires the CSI driver to be
			// deployed for Kubernetes 1.25 and newer.
			// Existing clusters running the in-tree cloud provider must be migrated
			// to the external CCM/CSI before upgrading to Kubernetes 1.25.
			allErrs = append(allErrs, field.Invalid(fldPath.Child("cloudProvider").Child("external"), c.CloudProvider.External, "kubernetes 1.25 and newer doesn't support in-tree cloud provider with vsphere"))
		}
	}

	return allErrs
}

func ValidateContainerRuntimeConfig(cr kubeoneapi.ContainerRuntimeConfig, _ kubeoneapi.VersionConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allCRs := []any{
		cr.Containerd,
	}

	var found bool
	for _, x := range allCRs {
		if !reflect.ValueOf(x).IsNil() {
			if found {
				allErrs = append(allErrs, field.Invalid(fldPath, x, "only 1 container runtime can be activated"))
			}
			found = true
		}
	}

	return allErrs
}

// ValidateClusterNetworkConfig validates the ClusterNetworkConfig structure
func ValidateClusterNetworkConfig(c kubeoneapi.ClusterNetworkConfig, prov kubeoneapi.CloudProviderSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateIPFamily(c.IPFamily, prov, fldPath.Child("ipFamily"))...)
	allErrs = append(allErrs, validateCIDRs(c, fldPath)...)
	allErrs = append(allErrs, validateNodeCIDRMaskSize(c, fldPath)...)

	if c.CNI != nil {
		allErrs = append(allErrs, ValidateCNI(c.CNI, fldPath.Child("cni"))...)

		// validated cilium kube-proxy replacement
		if c.CNI.Cilium != nil && c.CNI.Cilium.KubeProxyReplacement && (c.KubeProxy == nil || !c.KubeProxy.SkipInstallation) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("cni"), c.CNI.Cilium.KubeProxyReplacement, ".cilium.kubeProxyReplacement cannot be set with kube-proxy enabled"))
		}
	}
	if c.KubeProxy != nil {
		allErrs = append(allErrs, ValidateKubeProxy(c.KubeProxy, fldPath.Child("kubeProxy"))...)
	}

	return allErrs
}

func validateIPFamily(ipFamily kubeoneapi.IPFamily, prov kubeoneapi.CloudProviderSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if ipFamily == kubeoneapi.IPFamilyIPv6 || ipFamily == kubeoneapi.IPFamilyIPv6IPv4 {
		allErrs = append(allErrs, field.Forbidden(fldPath, "ipv6 and ipv6+ipv4 ip families are currently not supported"))
	}

	if ipFamily == kubeoneapi.IPFamilyIPv4IPv6 && prov.AWS == nil && prov.None == nil && prov.Vsphere == nil {
		allErrs = append(allErrs, field.Forbidden(fldPath, "dualstack is currently supported only on AWS, vSphere and baremetal (none)"))
	}

	return allErrs
}

func validateNodeCIDRMaskSize(c kubeoneapi.ClusterNetworkConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	validateNodeCIDRMaskSize := func(nodeCIDRMaskSize *int, podCIDR string, fldPath *field.Path) {
		if nodeCIDRMaskSize == nil {
			allErrs = append(allErrs, field.Invalid(fldPath, nodeCIDRMaskSize, "node CIDR mask size must be set"))

			return
		}

		_, podCIDRNet, err := net.ParseCIDR(podCIDR)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath, podCIDR, fmt.Sprintf("couldn't parse CIDR %q: %v", podCIDR, err)))

			return
		}
		podCIDRMaskSize, _ := podCIDRNet.Mask.Size()

		if podCIDRMaskSize >= *nodeCIDRMaskSize {
			allErrs = append(allErrs, field.Invalid(fldPath, nodeCIDRMaskSize,
				fmt.Sprintf("node CIDR mask size (%d) must be longer than the mask size of the pod CIDR (%q)", *nodeCIDRMaskSize, podCIDR)))

			return
		}
	}

	switch c.IPFamily {
	case kubeoneapi.IPFamilyIPv4:
		validateNodeCIDRMaskSize(c.NodeCIDRMaskSizeIPv4, c.PodSubnet, fldPath.Child("nodeCIDRMaskSizeIPv4"))
	case kubeoneapi.IPFamilyIPv6:
		validateNodeCIDRMaskSize(c.NodeCIDRMaskSizeIPv6, c.PodSubnetIPv6, fldPath.Child("nodeCIDRMaskSizeIPv6"))
	case kubeoneapi.IPFamilyIPv4IPv6, kubeoneapi.IPFamilyIPv6IPv4:
		validateNodeCIDRMaskSize(c.NodeCIDRMaskSizeIPv4, c.PodSubnet, fldPath.Child("nodeCIDRMaskSizeIPv4"))
		validateNodeCIDRMaskSize(c.NodeCIDRMaskSizeIPv6, c.PodSubnetIPv6, fldPath.Child("nodeCIDRMaskSizeIPv6"))
	}

	return allErrs
}

func validateCIDRs(c kubeoneapi.ClusterNetworkConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	invalidFamilyErr := func(node, subnet string, ipFamily kubeoneapi.IPFamily) *field.Error {
		return field.Invalid(fldPath.Child(node), subnet, fmt.Sprintf(".clusterNetwork.%s must be valid %q subnet.", node, ipFamily))
	}

	validateCIDR := func(node, subnet string, ipFamily kubeoneapi.IPFamily) {
		switch ipFamily {
		case kubeoneapi.IPFamilyIPv4:
			if !netutils.IsIPv4CIDRString(subnet) {
				allErrs = append(allErrs, invalidFamilyErr(node, subnet, ipFamily))
			}
		case kubeoneapi.IPFamilyIPv6:
			if !netutils.IsIPv6CIDRString(subnet) {
				allErrs = append(allErrs, invalidFamilyErr(node, subnet, ipFamily))
			}
		case kubeoneapi.IPFamilyIPv4IPv6, kubeoneapi.IPFamilyIPv6IPv4:
			// just to make linter happy
		}
	}

	switch c.IPFamily {
	case kubeoneapi.IPFamilyIPv4:
		validateCIDR("podSubnet", c.PodSubnet, kubeoneapi.IPFamilyIPv4)
		validateCIDR("serviceSubnet", c.ServiceSubnet, kubeoneapi.IPFamilyIPv4)
	case kubeoneapi.IPFamilyIPv6:
		validateCIDR("podSubnetIPv6", c.PodSubnetIPv6, kubeoneapi.IPFamilyIPv6)
		validateCIDR("serviceSubnetIPv6", c.ServiceSubnetIPv6, kubeoneapi.IPFamilyIPv6)
	case kubeoneapi.IPFamilyIPv4IPv6, kubeoneapi.IPFamilyIPv6IPv4:
		validateCIDR("podSubnet", c.PodSubnet, kubeoneapi.IPFamilyIPv4)
		validateCIDR("serviceSubnet", c.ServiceSubnet, kubeoneapi.IPFamilyIPv4)
		validateCIDR("podSubnetIPv6", c.PodSubnetIPv6, kubeoneapi.IPFamilyIPv6)
		validateCIDR("serviceSubnetIPv6", c.ServiceSubnetIPv6, kubeoneapi.IPFamilyIPv6)
	default:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("ipFamily"), c.IPFamily, "unknown ipFamily"))
	}

	return allErrs
}

func ValidateKubeProxy(kbPrxConf *kubeoneapi.KubeProxyConfig, fldPath *field.Path) field.ErrorList {
	var (
		allErrs     field.ErrorList
		configFound bool
	)

	if kbPrxConf.IPTables != nil {
		configFound = true
	}

	if kbPrxConf.IPVS != nil {
		if configFound {
			allErrs = append(allErrs, field.Invalid(fldPath, "", "should have only 1, ether iptables or ipvs or none"))
		}
	}

	return allErrs
}

// ValidateCNI validates the CNI structure
func ValidateCNI(c *kubeoneapi.CNI, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	cniFound := false
	if c.Canal != nil {
		cniFound = true
		if c.Canal.MTU == 0 {
			allErrs = append(allErrs,
				field.Invalid(fldPath.Child("canal").Child("mtu"), c.Canal.MTU, "invalid value"))
		}
	}
	if c.Cilium != nil {
		if cniFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("cilium"), "only one cni plugin can be used at the same time"))
		}
		cniFound = true
	}
	if c.WeaveNet != nil {
		if cniFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("weaveNet"), "only one cni plugin can be used at the same time"))
		}
		cniFound = true
	}
	if c.External != nil {
		if cniFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("external"), "only one cni plugin can be used at the same time"))
		}
		cniFound = true
	}

	if !cniFound {
		allErrs = append(allErrs, field.Invalid(fldPath, "", "cni plugin must be specified"))
	}

	return allErrs
}

// ValidateStaticWorkersConfig validates the StaticWorkersConfig structure
func ValidateStaticWorkersConfig(staticWorkers kubeoneapi.StaticWorkersConfig, controlPlane kubeoneapi.ControlPlaneConfig, clusterNetwork kubeoneapi.ClusterNetworkConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(staticWorkers.Hosts) > 0 {
		allErrs = append(allErrs, ValidateHostConfig(staticWorkers.Hosts, clusterNetwork, fldPath.Child("hosts"))...)
	}

	for idx, worker := range staticWorkers.Hosts {
		for _, cp := range controlPlane.Hosts {
			if cp.Hostname != "" && worker.Hostname != "" && cp.Hostname == worker.Hostname {
				allErrs = append(allErrs,
					field.Invalid(
						fldPath.Child("hosts").Index(idx),
						"Hostname",
						fmt.Sprintf("hostname %q already used for control-plane node", worker.Hostname),
					),
				)
			}

			if cp.PrivateAddress == worker.PrivateAddress {
				allErrs = append(allErrs,
					field.Invalid(
						fldPath.Child("hosts").Index(idx),
						"PrivateAddress",
						fmt.Sprintf("private IP address %q already used for control-plane node", worker.PrivateAddress),
					),
				)
			}

			if cp.PublicAddress == worker.PublicAddress {
				allErrs = append(allErrs,
					field.Invalid(
						fldPath.Child("hosts").Index(idx),
						"PublicAddress",
						fmt.Sprintf("public IP address %q already used for control-plane node", worker.PublicAddress),
					),
				)
			}
		}
	}

	return allErrs
}

// ValidateDynamicWorkerConfig validates the DynamicWorkerConfig structure
func ValidateDynamicWorkerConfig(workerset []kubeoneapi.DynamicWorkerConfig, prov kubeoneapi.CloudProviderSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for _, w := range workerset {
		if w.Name == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("name"), ".dynamicWorkers.name is a required field"))
		}
		if w.Replicas == nil || *w.Replicas < 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("replicas"), w.Replicas, ".dynamicWorkers.replicas must be specified and >= 0"))
		}
		if w.Config.Network != nil && w.Config.Network.IPFamily != "" {
			allErrs = append(allErrs, validateIPFamily(w.Config.Network.IPFamily, prov, fldPath.Child("network", "ipFamily"))...)
		}
	}

	return allErrs
}

func ValidateCABundle(caBundle string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	caPEM := bytes.TrimSpace([]byte(caBundle))
	if len(caPEM) == 0 {
		return allErrs
	}

	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(caPEM); !ok {
		allErrs = append(allErrs, field.Invalid(fldPath, "", "can't parse caBundle"))
	}

	return allErrs
}

// ValidateFeatures validates the Features structure
func ValidateFeatures(f kubeoneapi.Features, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if f.CoreDNS != nil && f.CoreDNS.Replicas != nil && *f.CoreDNS.Replicas < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("coreDNS", "replicas"), *f.CoreDNS.Replicas, "coreDNS replicas cannot be < 0"))
	}
	if f.PodNodeSelector != nil && f.PodNodeSelector.Enable {
		allErrs = append(allErrs, ValidatePodNodeSelectorConfig(f.PodNodeSelector.Config, fldPath.Child("podNodeSelector"))...)
	}
	if f.StaticAuditLog != nil && f.StaticAuditLog.Enable {
		allErrs = append(allErrs, ValidateStaticAuditLogConfig(f.StaticAuditLog.Config, fldPath.Child("staticAuditLog"))...)
	}
	if f.WebhookAuditLog != nil && f.WebhookAuditLog.Enable {
		allErrs = append(allErrs, ValidateWebhookAuditLogConfig(f.WebhookAuditLog.Config, fldPath.Child("webhookAuditLog"))...)
	}
	if f.OpenIDConnect != nil && f.OpenIDConnect.Enable {
		allErrs = append(allErrs, ValidateOIDCConfig(f.OpenIDConnect.Config, fldPath.Child("openidConnect"))...)
	}

	return allErrs
}

// ValidatePodNodeSelectorConfig validates the PodNodeSelectorConfig structure
func ValidatePodNodeSelectorConfig(n kubeoneapi.PodNodeSelectorConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(n.ConfigFilePath) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("configFilePath"), ".podNodeSelector.config.configFilePath is a required field"))
	}

	return allErrs
}

// ValidateStaticAuditLogConfig validates the StaticAuditLogConfig structure
func ValidateStaticAuditLogConfig(s kubeoneapi.StaticAuditLogConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(s.PolicyFilePath) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("policyFilePath"), ".staticAuditLog.config.policyFilePath is a required field"))
	}
	if len(s.LogPath) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("logPath"), ".staticAuditLog.config.logPath is a required field"))
	}
	if s.LogMaxAge <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("logMaxAge"), s.LogMaxAge, ".staticAuditLog.config.logMaxAge must be greater than 0"))
	}
	if s.LogMaxBackup <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("logMaxBackup"), s.LogMaxBackup, ".staticAuditLog.config.logMaxBackup must be greater than 0"))
	}
	if s.LogMaxSize <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("logMaxSize"), s.LogMaxSize, ".staticAuditLog.config.logMaxSize must be greater than 0"))
	}

	return allErrs
}

// ValidateWebhookAuditLogConfig validates the WebhookAuditLogConfig structure
func ValidateWebhookAuditLogConfig(s kubeoneapi.WebhookAuditLogConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(s.PolicyFilePath) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("policyFilePath"), ".staticAuditLog.config.policyFilePath is a required field"))
	}
	if len(s.ConfigFilePath) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("configFilePath"), ".webhookAuditLog.config.configFilePath is a required field"))
	}

	return allErrs
}

// ValidateOIDCConfig validates the OpenIDConnectConfig structure
func ValidateOIDCConfig(o kubeoneapi.OpenIDConnectConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(o.IssuerURL) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("issuerURL"), ".openidConnect.config.issuerURL is a required field"))
	}
	if len(o.ClientID) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("clientID"), ".openidConnect.config.clientID is a required field"))
	}

	return allErrs
}

// ValidateAddons validates the Addons configuration
func ValidateAddons(o *kubeoneapi.Addons, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateHelmReleases(o.OnlyHelmReleases(), fldPath)...)

	return allErrs
}

func validateHelmReleases(helmReleases []kubeoneapi.HelmRelease, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for i, hr := range helmReleases {
		fldPath := fldPath.Index(i) //nolint:govet
		if hr.Chart == "" && hr.ChartURL == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("chart"), hr.Chart))
		}

		if hr.Auth != nil {
			if hr.Auth.Username != "" && hr.Auth.Password == "" {
				allErrs = append(allErrs, field.Required(fldPath.Child("auth").Child("password"), "password is required when username is set"))
			}
			if hr.Auth.Password != "" && hr.Auth.Username == "" {
				allErrs = append(allErrs, field.Required(fldPath.Child("auth").Child("username"), "username is required when password is set"))
			}
		}

		if hr.Namespace == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("namespace"), hr.Namespace))
		}

		if hr.RepoURL == "" && hr.ChartURL == "" {
			_, err := helm.GetChartNameFromChartYAML(hr.Chart)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("chart"), hr.Chart, fmt.Sprintf("invalid local chart: %v", err)))
			}
		}

		if hr.ChartURL != "" && hr.ReleaseName == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("releaseName"), "since chartURL is given directly, releaseName is required"))
		}

		for idx, helmValues := range hr.Values {
			fldIdentity := fldPath.Child("values").Index(idx)

			if helmValues.ValuesFile != "" {
				err := func() error {
					valFile, err := os.Open(helmValues.ValuesFile)
					if valFile != nil {
						defer valFile.Close()
					}

					return err
				}()
				if err != nil {
					allErrs = append(allErrs,
						field.Invalid(fldIdentity.Child("valuesFile"), hr.Values[idx].ValuesFile, fmt.Sprintf("file is invalid: %v", err)),
					)
				}
			}

			if helmValues.Inline != nil {
				obj := map[string]any{}
				err := yaml.Unmarshal(helmValues.Inline, &obj)
				if err != nil {
					allErrs = append(allErrs,
						field.Invalid(fldIdentity.Child("inline"), hr.Values[idx].Inline, fmt.Sprintf("inline is not a valid YAML: %v", err)),
					)
				}
			}
		}
	}

	return allErrs
}

// ValidateHostConfig validates the HostConfig structure
func ValidateHostConfig(hosts []kubeoneapi.HostConfig, clusterNetwork kubeoneapi.ClusterNetworkConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	leaderFound := false
	for idx, host := range hosts {
		hostFldPath := fldPath.Index(idx)

		if leaderFound && host.IsLeader {
			allErrs = append(allErrs, field.Invalid(hostFldPath.Child("isLeader"), host.IsLeader, "only one leader is allowed"))
		}
		if host.IsLeader {
			leaderFound = true
		}
		if len(host.PublicAddress) == 0 {
			allErrs = append(allErrs, field.Required(hostFldPath.Child("publicAddress"), "no public IP/address given"))
		}
		if (clusterNetwork.IPFamily == kubeoneapi.IPFamilyIPv6 || clusterNetwork.IPFamily == kubeoneapi.IPFamilyIPv4IPv6 || clusterNetwork.IPFamily == kubeoneapi.IPFamilyIPv6IPv4) && len(host.IPv6Addresses) == 0 {
			allErrs = append(allErrs, field.Required(hostFldPath.Child("ipFamily"), "no IPv6 address given"))
		}
		if len(host.PrivateAddress) == 0 {
			allErrs = append(allErrs, field.Required(hostFldPath.Child("privateAddress"), "no private IP/address givevn"))
		}
		if len(host.SSHPrivateKeyFile) == 0 && len(host.SSHAgentSocket) == 0 {
			allErrs = append(allErrs, field.Invalid(hostFldPath.Child("sshPrivateKeyFile"), host.SSHPrivateKeyFile, "neither SSH private key nor agent socket given, don't know how to authenticate"))
			allErrs = append(allErrs, field.Invalid(hostFldPath.Child("sshAgentSocket"), host.SSHAgentSocket, "neither SSH private key nor agent socket given, don't know how to authenticate"))
		}
		if len(host.SSHUsername) == 0 {
			allErrs = append(allErrs, field.Required(hostFldPath.Child("sshUsername"), "no SSH username given"))
		}
		if !host.OperatingSystem.IsValid() {
			allErrs = append(allErrs, field.Invalid(hostFldPath.Child("operatingSystem"), host.OperatingSystem, "invalid operatingSystem provided"))
		}
		allErrs = append(allErrs, ValidateKubeletConfig(host.Kubelet, hostFldPath.Child("kubelet"))...)
		allErrs = append(allErrs, validateLabels(host.Annotations, hostFldPath.Child("annotations"))...)
		allErrs = append(allErrs, validateLabels(host.Labels, hostFldPath.Child("labels"))...)
		for _, taint := range host.Taints {
			if taint.Key == "node-role.kubernetes.io/master" {
				allErrs = append(allErrs, field.Forbidden(hostFldPath.Child("taints"), fmt.Sprintf("%q taint is forbidden for clusters running Kubernetes 1.25+", "node-role.kubernetes.io/master")))
			}
		}
	}

	return allErrs
}

func validateLabels(kv map[string]string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for labelKey, labelValue := range kv {
		if strings.HasSuffix(labelKey, "-") && labelValue != "" {
			allErrs = append(allErrs, field.Invalid(fldPath, labelValue, "key to remove cannot have value"))
		}
	}

	return allErrs
}

func ValidateRegistryConfiguration(r *kubeoneapi.RegistryConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if r == nil {
		return allErrs
	}

	if r.InsecureRegistry && r.OverwriteRegistry == "" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("insecureRegistry"), r.InsecureRegistry, "insecureRegistry requires overwriteRegistry to be configured"))
	}

	return allErrs
}

func ValidateControlPlaneComponents(c *kubeoneapi.ControlPlaneComponents, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if c == nil {
		return allErrs
	}

	if c.ControllerManager != nil && c.ControllerManager.Flags != nil {
		if _, ok := c.ControllerManager.Flags["feature-gates"]; ok {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("controllerManager").Child("flags"), c.ControllerManager.Flags, "specifying feature-gates flag is forbidden here. Use .controlPlaneComponents.controllerManager.featureGates instead"))
		}
	}

	if c.Scheduler != nil && c.Scheduler.Flags != nil {
		if _, ok := c.Scheduler.Flags["feature-gates"]; ok {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("scheduler").Child("flags"), c.Scheduler.Flags, "specifying feature-gates flag is forbidden here. Use .controlPlaneComponents.scheduler.featureGates instead"))
		}
	}

	if c.APIServer != nil && c.APIServer.Flags != nil {
		if _, ok := c.APIServer.Flags["feature-gates"]; ok {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("apiServer").Child("flags"), c.APIServer.Flags, "specifying feature-gates flag is forbidden here. Use .controlPlaneComponents.apiServer.featureGates instead"))
		}
	}

	return allErrs
}

func ValidateAssetConfiguration(a *kubeoneapi.AssetConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if a.Kubernetes.ImageTag != "" {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("imageTag"), "imageTag is forbidden for Kubernetes images"))
	}

	if a.Pause.ImageRepository != "" && a.Pause.ImageTag == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("imageTag"), "imageTag for sandbox (pause) image is required"))
	}
	if a.Pause.ImageRepository == "" && a.Pause.ImageTag != "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("imageRepository"), "imageRepository for sandbox (pause) image is required"))
	}

	found := 0
	if a.CNI.URL != "" {
		found++
	}
	if a.NodeBinaries.URL != "" {
		found++
	}
	if a.Kubectl.URL != "" {
		found++
	}
	if found != 0 && found != 3 {
		allErrs = append(allErrs, field.Invalid(fldPath, "", "all binary assets must be specified (cni, nodeBinaries, kubectl)"))
	}

	return allErrs
}

// ValidateOperatingSystemManager validates the OperatingSystemManager structure
func ValidateOperatingSystemManager(mc *kubeoneapi.MachineControllerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if mc == nil || !mc.Deploy {
		allErrs = append(allErrs, field.Invalid(fldPath, "", "machineController needs to be enabled to use operatingSystemManager"))
	}

	return allErrs
}

func ValidateKubeletConfig(klcfg kubeoneapi.KubeletConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if klcfg.MaxPods != nil && *klcfg.MaxPods <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("maxPods"), klcfg.MaxPods, "maxPods must be a positive number"))
	}
	if v := klcfg.ImageGCHighThresholdPercent; v != nil && (*v < 0 || *v > 100) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("imageGCHighThresholdPercent"), *v, "must be between 0 and 100, inclusive"))
	}
	if v := klcfg.ImageGCLowThresholdPercent; v != nil && (*v < 0 || *v > 100) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("imageGCLowThresholdPercent"), *v, "must be between 0 and 100, inclusive"))
	}

	return allErrs
}
