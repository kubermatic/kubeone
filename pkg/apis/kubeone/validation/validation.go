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
	"reflect"
	"strings"

	"github.com/Masterminds/semver/v3"

	"k8c.io/kubeone/pkg/addons"
	"k8c.io/kubeone/pkg/apis/kubeone"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateKubeOneCluster validates the KubeOneCluster object
func ValidateKubeOneCluster(c kubeone.KubeOneCluster) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateName(c.Name, field.NewPath("name"))...)
	allErrs = append(allErrs, ValidateControlPlaneConfig(c.ControlPlane, field.NewPath("controlPlane"))...)
	allErrs = append(allErrs, ValidateAPIEndpoint(c.APIEndpoint, field.NewPath("apiEndpoint"))...)
	allErrs = append(allErrs, ValidateCloudProviderSpec(c.CloudProvider, field.NewPath("provider"))...)
	allErrs = append(allErrs, ValidateVersionConfig(c.Versions, field.NewPath("versions"))...)
	allErrs = append(allErrs, ValidateKubernetesSupport(c, field.NewPath(""))...)
	allErrs = append(allErrs, ValidateContainerRuntimeConfig(c.ContainerRuntime, c.Versions, field.NewPath("containerRuntime"))...)
	allErrs = append(allErrs, ValidateClusterNetworkConfig(c.ClusterNetwork, field.NewPath("clusterNetwork"))...)
	allErrs = append(allErrs, ValidateStaticWorkersConfig(c.StaticWorkers, field.NewPath("staticWorkers"))...)

	if c.MachineController != nil && c.MachineController.Deploy {
		allErrs = append(allErrs, ValidateDynamicWorkerConfig(c.DynamicWorkers, field.NewPath("dynamicWorkers"))...)
	} else if len(c.DynamicWorkers) > 0 {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("dynamicWorkers"),
			"machine-controller deployment is disabled, but the configuration still contains dynamic workers"))
	}

	allErrs = append(allErrs, ValidateCABundle(c.CABundle, field.NewPath("caBundle"))...)
	allErrs = append(allErrs, ValidateFeatures(c.Features, c.Versions, field.NewPath("features"))...)
	allErrs = append(allErrs, ValidateAddons(c.Addons, field.NewPath("addons"))...)
	allErrs = append(allErrs, ValidateRegistryConfiguration(c.RegistryConfiguration, field.NewPath("registryConfiguration"))...)
	allErrs = append(allErrs,
		ValidateContainerRuntimeVSRegistryConfiguration(
			c.ContainerRuntime,
			field.NewPath("containerRuntime"),
			c.RegistryConfiguration,
			field.NewPath("registryConfiguration"),
		)...)

	return allErrs
}

func ValidateContainerRuntimeVSRegistryConfiguration(
	cr kubeone.ContainerRuntimeConfig,
	crFldPath *field.Path,
	rc *kubeone.RegistryConfiguration,
	rcFldPath *field.Path,
) field.ErrorList {
	allErrs := field.ErrorList{}

	switch {
	case rc == nil:
	case cr.Containerd != nil && cr.Containerd.Registries != nil:
		containerdRegistriesField := crFldPath.Child("containerd", "registries")
		allErrs = append(allErrs, field.Invalid(
			containerdRegistriesField,
			"",
			fmt.Sprintf("can't have both %s and %s set", rcFldPath.String(), containerdRegistriesField.String()),
		))
	case cr.Docker != nil && cr.Docker.RegistryMirrors != nil:
		dockerRegistryMirrorsField := crFldPath.Child("docker", "registryMirrors")
		allErrs = append(allErrs, field.Invalid(
			dockerRegistryMirrorsField,
			"",
			fmt.Sprintf("can't have both %s and %s set", rcFldPath.String(), dockerRegistryMirrorsField.String()),
		))
	}

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
func ValidateControlPlaneConfig(c kubeone.ControlPlaneConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(c.Hosts) > 0 {
		allErrs = append(allErrs, ValidateHostConfig(c.Hosts, fldPath.Child("hosts"))...)
	} else {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("hosts"), "",
			".controlPlane.Hosts is a required field. There must be at least one control plane instance in the cluster."))
	}

	return allErrs
}

// ValidateAPIEndpoint validates the APIEndpoint structure
func ValidateAPIEndpoint(a kubeone.APIEndpoint, fldPath *field.Path) field.ErrorList {
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
		} else {
			visited[altName] = true
		}
	}

	return allErrs
}

// ValidateCloudProviderSpec validates the CloudProviderSpec structure
func ValidateCloudProviderSpec(p kubeone.CloudProviderSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	providerFound := false
	if p.AWS != nil {
		providerFound = true
	}
	if p.Azure != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("azure"), "only one provider can be used at the same time"))
		}
		if len(p.CloudConfig) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("cloudConfig"), ".cloudProvider.cloudConfig is required for azure provider"))
		}
		providerFound = true
	}
	if p.DigitalOcean != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("digitalocean"), "only one provider can be used at the same time"))
		}
		providerFound = true
	}
	if p.GCE != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("gce"), "only one provider can be used at the same time"))
		}
		providerFound = true
	}
	if p.Hetzner != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("hetzner"), "only one provider can be used at the same time"))
		}
		providerFound = true
	}
	if p.Openstack != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("openstack"), "only one provider can be used at the same time"))
		}
		if len(p.CloudConfig) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("cloudConfig"), ".cloudProvider.cloudConfig is required for openstack provider"))
		}
		providerFound = true
	}
	if p.EquinixMetal != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("equinixmetal"), "only one provider can be used at the same time"))
		}
		providerFound = true
	}
	if p.Vsphere != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("vsphere"), "only one provider can be used at the same time"))
		}
		if len(p.CloudConfig) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("cloudConfig"), ".cloudProvider.cloudConfig is required for vSphere provider"))
		}
		providerFound = true
	}
	if p.None != nil {
		if providerFound {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("none"), "only one provider can be used at the same time"))
		}
		providerFound = true
	}

	if !providerFound {
		allErrs = append(allErrs, field.Invalid(fldPath, "", "provider must be specified"))
	}

	if len(p.CSIConfig) > 0 {
		if p.Vsphere == nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("csiConfig"), "", ".cloudProvider.csiConfig is currently supported only for vsphere clusters"))
		}
		if !p.External {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("csiConfig"), "", ".cloudProvider.csiConfig is supported only for clusters using external cloud provider (.cloudProvider.external)"))
		}
	}

	return allErrs
}

// ValidateVersionConfig validates the VersionConfig structure
func ValidateVersionConfig(version kubeone.VersionConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	v, err := semver.NewVersion(version.Kubernetes)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("kubernetes"), version, ".versions.kubernetes is not a semver string"))
		return allErrs
	}
	if v.Major() != 1 || v.Minor() < 19 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("kubernetes"), version, "kubernetes versions lower than 1.19 are not supported. You need to use an older KubeOne version to upgrade your cluster to v1.19. Please refer to the Compatibility section of docs for more details."))
	}
	if strings.HasPrefix(version.Kubernetes, "v") {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("kubernetes"), version, ".versions.kubernetes can't start with a leading 'v'"))
	}

	return allErrs
}

func ValidateKubernetesSupport(c kubeone.KubeOneCluster, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if strings.Contains(c.Versions.Kubernetes, "-eks-") {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("versions").Child("kubernetes"), c.Versions.Kubernetes, "Amazon EKS-D clusters are not supported by KubeOne 1.4+"))
		return allErrs
	}

	v, err := semver.NewVersion(c.Versions.Kubernetes)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("versions").Child("kubernetes"), c.Versions.Kubernetes, ".versions.kubernetes is not a semver string"))
		return allErrs
	}

	if v.Minor() >= 23 {
		switch {
		case c.CloudProvider.Vsphere != nil:
			allErrs = append(allErrs, field.Invalid(fldPath.Child("versions").Child("kubernetes"), c.Versions.Kubernetes, "kubernetes versions 1.23.0 and newer are currently not supported for vsphere clusters"))
		case c.ClusterNetwork.KubeProxy != nil && c.ClusterNetwork.KubeProxy.IPVS != nil:
			if c.ClusterNetwork.CNI != nil && c.ClusterNetwork.CNI.Canal != nil {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("versions").Child("kubernetes"), c.Versions.Kubernetes, "kubernetes versions 1.23.0 and newer are currently not supported for clusters running kube-proxy in ipvs mode with Canal CNI"))
			} else if c.ClusterNetwork.CNI != nil && c.ClusterNetwork.CNI.External != nil && c.Addons != nil {
				for _, addon := range c.Addons.Addons {
					if addon.Name == "calico-vxlan" {
						allErrs = append(allErrs, field.Invalid(fldPath.Child("versions").Child("kubernetes"), c.Versions.Kubernetes, "kubernetes versions 1.23.0 and newer are currently not supported for clusters running kube-proxy in ipvs mode with Calico CNI"))
					}
				}
			}
		}
	}

	return allErrs
}

func ValidateContainerRuntimeConfig(cr kubeone.ContainerRuntimeConfig, versions kubeone.VersionConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allCRs := []interface{}{
		cr.Docker,
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

	if cr.Docker != nil {
		kubeVer, _ := semver.NewVersion(versions.Kubernetes)
		gteKube122Condition, _ := semver.NewConstraint(">= 1.22")
		if gteKube122Condition.Check(kubeVer) {
			allErrs = append(allErrs, field.Invalid(fldPath, cr.Docker, "kubernetes v1.22+ require containerd container runtime"))
		}
	}

	return allErrs
}

// ValidateClusterNetworkConfig validates the ClusterNetworkConfig structure
func ValidateClusterNetworkConfig(c kubeone.ClusterNetworkConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(c.PodSubnet) > 0 {
		if _, _, err := net.ParseCIDR(c.PodSubnet); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("podSubnet"), c.PodSubnet, ".clusterNetwork.podSubnet must be a valid CIDR string"))
		}
	}
	if len(c.ServiceSubnet) > 0 {
		if _, _, err := net.ParseCIDR(c.ServiceSubnet); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("serviceSubnet"), c.ServiceSubnet, ".clusterNetwork.serviceSubnet must be a valid CIDR string"))
		}
	}

	if c.CNI != nil {
		allErrs = append(allErrs, ValidateCNI(c.CNI, fldPath.Child("cni"))...)

		// validated cilium kube-proxy replacement
		if c.CNI.Cilium != nil && c.CNI.Cilium.KubeProxyReplacement != kubeone.KubeProxyReplacementDisabled && (c.KubeProxy == nil || !c.KubeProxy.SkipInstallation) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("cni"), c.CNI.Cilium.KubeProxyReplacement, ".cilium.kubeProxyReplacement cannot be set with kube-proxy enabled"))
		}
	}
	if c.KubeProxy != nil {
		allErrs = append(allErrs, ValidateKubeProxy(c.KubeProxy, fldPath.Child("kubeProxy"))...)
	}

	return allErrs
}

func ValidateKubeProxy(kbPrxConf *kubeone.KubeProxyConfig, fldPath *field.Path) field.ErrorList {
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
func ValidateCNI(c *kubeone.CNI, fldPath *field.Path) field.ErrorList {
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
func ValidateStaticWorkersConfig(staticWorkers kubeone.StaticWorkersConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(staticWorkers.Hosts) > 0 {
		allErrs = append(allErrs, ValidateHostConfig(staticWorkers.Hosts, fldPath.Child("hosts"))...)
	}

	return allErrs
}

// ValidateDynamicWorkerConfig validates the DynamicWorkerConfig structure
func ValidateDynamicWorkerConfig(workerset []kubeone.DynamicWorkerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for _, w := range workerset {
		if w.Name == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("name"), ".dynamicWorkers.name is a required field"))
		}
		if w.Replicas == nil || *w.Replicas < 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("replicas"), w.Replicas, ".dynamicWorkers.replicas must be specified and >= 0"))
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
func ValidateFeatures(f kubeone.Features, versions kubeone.VersionConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if f.PodNodeSelector != nil && f.PodNodeSelector.Enable {
		allErrs = append(allErrs, ValidatePodNodeSelectorConfig(f.PodNodeSelector.Config, fldPath.Child("podNodeSelector"))...)
	}
	if f.StaticAuditLog != nil && f.StaticAuditLog.Enable {
		allErrs = append(allErrs, ValidateStaticAuditLogConfig(f.StaticAuditLog.Config, fldPath.Child("staticAuditLog"))...)
	}
	if f.OpenIDConnect != nil && f.OpenIDConnect.Enable {
		allErrs = append(allErrs, ValidateOIDCConfig(f.OpenIDConnect.Config, fldPath.Child("openidConnect"))...)
	}
	return allErrs
}

// ValidatePodNodeSelectorConfig validates the PodNodeSelectorConfig structure
func ValidatePodNodeSelectorConfig(n kubeone.PodNodeSelectorConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(n.ConfigFilePath) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("configFilePath"), ".podNodeSelector.config.configFilePath is a required field"))
	}

	return allErrs
}

// ValidateStaticAuditLogConfig validates the StaticAuditLogConfig structure
func ValidateStaticAuditLogConfig(s kubeone.StaticAuditLogConfig, fldPath *field.Path) field.ErrorList {
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

// ValidateOIDCConfig validates the OpenIDConnectConfig structure
func ValidateOIDCConfig(o kubeone.OpenIDConnectConfig, fldPath *field.Path) field.ErrorList {
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
func ValidateAddons(o *kubeone.Addons, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if o == nil || !o.Enable {
		return allErrs
	}
	if o.Enable && len(o.Path) == 0 {
		// Addons are enabled, path is empty, and no embedded addon is specified
		if len(o.Addons) == 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("enable"), o.Enable, ".addons.enable cannot be set to true without specifying either custom addon path or embedded addon"))
		}

		// Check if only embedded addons are being used; path is not required for embedded addons
		embeddedAddonsOnly, err := addons.EmbeddedAddonsOnly(o.Addons)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath, "", "failed to read embedded addons directory"))
		} else if !embeddedAddonsOnly {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("path"), "", ".addons.path must be specified when using non-embedded addon(s)"))
		}
	}

	return allErrs
}

// ValidateHostConfig validates the HostConfig structure
func ValidateHostConfig(hosts []kubeone.HostConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	leaderFound := false
	for _, h := range hosts {
		if leaderFound && h.IsLeader {
			allErrs = append(allErrs, field.Invalid(fldPath, h.IsLeader, "only one leader is allowed"))
		}
		if h.IsLeader {
			leaderFound = true
		}
		if len(h.PublicAddress) == 0 {
			allErrs = append(allErrs, field.Required(fldPath, "no public IP/address given"))
		}
		if len(h.PrivateAddress) == 0 {
			allErrs = append(allErrs, field.Required(fldPath, "no private IP/address givevn"))
		}
		if len(h.SSHPrivateKeyFile) == 0 && len(h.SSHAgentSocket) == 0 {
			allErrs = append(allErrs, field.Invalid(fldPath, h.SSHPrivateKeyFile, "neither SSH private key nor agent socket given, don't know how to authenticate"))
			allErrs = append(allErrs, field.Invalid(fldPath, h.SSHAgentSocket, "neither SSH private key nor agent socket given, don't know how to authenticate"))
		}
		if len(h.SSHUsername) == 0 {
			allErrs = append(allErrs, field.Required(fldPath, "no SSH username given"))
		}
	}

	return allErrs
}

func ValidateRegistryConfiguration(r *kubeone.RegistryConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if r == nil {
		return allErrs
	}

	if r.InsecureRegistry && r.OverwriteRegistry == "" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("insecureRegistry"), r.InsecureRegistry, "insecureRegistry requires overwriteRegistry to be configured"))
	}

	return allErrs
}

func ValidateAssetConfiguration(a *kubeone.AssetConfiguration, fldPath *field.Path) field.ErrorList {
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
