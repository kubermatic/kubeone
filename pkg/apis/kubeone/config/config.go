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

package config

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubeonescheme "k8c.io/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	kubeonev1beta3 "k8c.io/kubeone/pkg/apis/kubeone/v1beta3"
	kubeonevalidation "k8c.io/kubeone/pkg/apis/kubeone/validation"
	"k8c.io/kubeone/pkg/containerruntime"
	"k8c.io/kubeone/pkg/fail"
	terraformv1beta2 "k8c.io/kubeone/pkg/terraform/v1beta2"
	terraformv1beta3 "k8c.io/kubeone/pkg/terraform/v1beta3"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

const (
	// KubeOneClusterKind is kind of the KubeOneCluster object
	KubeOneClusterKind                  = "KubeOneCluster"
	controlPlaneComponentsWarning       = "Usage of the .controlPlaneComponents feature is at your own risk since options configured via this feature cannot properly be validated by KubeOne"
	flagsAndFeatureGateOverridesWarning = "\t- %s only covers %s. Some features might also need additional configuration for other components."
)

var (
	// AllowedAPIs contains APIs which are allowed to be used
	AllowedAPIs = map[string]string{
		kubeonev1beta2.SchemeGroupVersion.String(): "",
		// kubeonev1beta3.SchemeGroupVersion.String(): "",
	}

	// DeprecatedAPIs contains APIs which are deprecated
	DeprecatedAPIs = map[string]string{
		// kubeonev1beta2.SchemeGroupVersion.String(): "",
	}
)

// LoadKubeOneCluster returns the internal representation of the KubeOneCluster object
// parsed from the versioned KubeOneCluster manifest, Terraform output and credentials file
func LoadKubeOneCluster(clusterCfgPath, tfOutputPath, credentialsFilePath string, logger logrus.FieldLogger) (*kubeoneapi.KubeOneCluster, error) {
	if len(clusterCfgPath) == 0 {
		return nil, fail.Runtime(fmt.Errorf("is not provided"), "cluster configuration path")
	}

	cfgAbsPath, err := filepath.Abs(clusterCfgPath)
	if err != nil {
		return nil, err
	}
	cfgBaseDir := filepath.Dir(cfgAbsPath)

	cluster, err := os.ReadFile(clusterCfgPath)
	if err != nil {
		return nil, fail.Runtime(err, "reading cluster configuration")
	}

	tfOutput, err := TFOutput(tfOutputPath)
	if err != nil {
		return nil, err
	}

	var credentialsFile []byte
	if len(credentialsFilePath) != 0 {
		credentialsFile, err = os.ReadFile(credentialsFilePath)
		if err != nil {
			return nil, fail.Runtime(err, "reading credentials file")
		}
	}

	return BytesToKubeOneCluster(cluster, tfOutput, credentialsFile, logger, cfgBaseDir)
}

func TFOutput(tfOutputPath string) ([]byte, error) {
	var (
		tfOutput []byte
		err      error
	)

	switch {
	case tfOutputPath == "-":
		if tfOutput, err = io.ReadAll(os.Stdin); err != nil {
			return nil, fail.Runtime(err, "reading terraform output from stdin")
		}
	case isDir(tfOutputPath):
		cmd := exec.Command("terraform", "output", "-json")
		cmd.Dir = tfOutputPath
		if tfOutput, err = cmd.Output(); err != nil {
			return nil, fail.Runtime(err, "reading terraform output")
		}
	case len(tfOutputPath) != 0:
		if tfOutput, err = os.ReadFile(tfOutputPath); err != nil {
			return nil, fail.Runtime(err, "reading terraform output file")
		}
	}

	return tfOutput, nil
}

// BytesToKubeOneCluster parses the bytes of the versioned KubeOneCluster manifests
func BytesToKubeOneCluster(cluster, tfOutput, credentialsFile []byte, logger logrus.FieldLogger, baseDir string) (*kubeoneapi.KubeOneCluster, error) {
	// Get the GVK from the given KubeOneCluster manifest
	typeMeta := runtime.TypeMeta{}
	if err := yaml.Unmarshal(cluster, &typeMeta); err != nil {
		return nil, fail.Config(err, "unmarshal cluster typeMeta")
	}
	if len(typeMeta.APIVersion) == 0 || len(typeMeta.Kind) == 0 {
		return nil, fail.ConfigValidation(fmt.Errorf("apiVersion and kind must be present in the manifest"))
	}
	if typeMeta.Kind != KubeOneClusterKind {
		return nil, fail.ConfigValidation(fmt.Errorf("provided object %q is not KubeOneCluster object", typeMeta.Kind))
	}
	if _, ok := AllowedAPIs[typeMeta.APIVersion]; !ok {
		return nil, fail.ConfigValidation(fmt.Errorf("provided apiVersion %q is not supported", typeMeta.APIVersion))
	}
	if _, ok := DeprecatedAPIs[typeMeta.APIVersion]; ok {
		logger.Warningf(`The provided APIVersion %q is deprecated. Please use "kubeone config migrate" command to migrate to the latest version.`, typeMeta.APIVersion)
	}

	var (
		internalCluster *kubeoneapi.KubeOneCluster
		err             error
	)

	// Parse the cluster bytes depending on the GVK
	switch typeMeta.APIVersion {
	case kubeonev1beta2.SchemeGroupVersion.String():
		v1beta2Cluster := kubeonev1beta2.NewKubeOneCluster()
		if err = runtime.DecodeInto(kubeonescheme.Codecs.UniversalDecoder(), cluster, v1beta2Cluster); err != nil {
			return nil, fail.Config(err, fmt.Sprintf("decoding %s", v1beta2Cluster.GroupVersionKind()))
		}

		internalCluster, err = DefaultedV1Beta2KubeOneCluster(v1beta2Cluster, tfOutput)
		if err != nil {
			return nil, err
		}
	// case kubeonev1beta3.SchemeGroupVersion.String():
	// 	v1beta3Cluster := kubeonev1beta3.NewKubeOneCluster()
	// 	if err := runtime.DecodeInto(kubeonescheme.Codecs.UniversalDecoder(), cluster, v1beta3Cluster); err != nil {
	// 		return nil, fail.Config(err, fmt.Sprintf("decoding %s", v1beta3Cluster.GroupVersionKind()))
	// 	}

	// 	internalCluster, err = DefaultedV1Beta3KubeOneCluster(v1beta3Cluster, tfOutput, credentialsFile, logger)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	default:
		return nil, fail.Config(fmt.Errorf("invalid api version %q", typeMeta.APIVersion), "api version")
	}

	// Apply the dynamic defaults
	if err := SetKubeOneClusterDynamicDefaults(internalCluster, credentialsFile, baseDir); err != nil {
		return nil, err
	}

	// Validate the configuration
	if err := kubeonevalidation.ValidateKubeOneCluster(*internalCluster).ToAggregate(); err != nil {
		return nil, fail.ConfigValidation(err)
	}

	// Check for deprecated fields/features for a cluster
	checkClusterConfiguration(*internalCluster, logger)

	return internalCluster, nil
}

// DefaultedV1Beta2KubeOneCluster converts a v1beta2 KubeOneCluster object to an internal representation of KubeOneCluster
// object while sourcing information from Terraform output, applying default values and validating the KubeOneCluster
// object
func DefaultedV1Beta2KubeOneCluster(versionedCluster *kubeonev1beta2.KubeOneCluster, tfOutput []byte) (*kubeoneapi.KubeOneCluster, error) {
	if tfOutput != nil {
		tfConfig, err := terraformv1beta2.NewConfigFromJSON(tfOutput)
		if err != nil {
			return nil, err
		}
		if err := tfConfig.Apply(versionedCluster); err != nil {
			return nil, err
		}
	}

	internalCluster := &kubeoneapi.KubeOneCluster{}

	kubeonescheme.Scheme.Default(versionedCluster)
	if err := kubeonescheme.Scheme.Convert(versionedCluster, internalCluster, nil); err != nil {
		return nil, fail.Config(err, fmt.Sprintf("converting %s to internal object", versionedCluster.GroupVersionKind()))
	}

	return internalCluster, nil
}

// DefaultedV1Beta3KubeOneCluster converts a v1beta3 KubeOneCluster object to an internal representation of KubeOneCluster
// object while sourcing information from Terraform output, applying default values and validating the KubeOneCluster
// object
func DefaultedV1Beta3KubeOneCluster(versionedCluster *kubeonev1beta3.KubeOneCluster, tfOutput []byte) (*kubeoneapi.KubeOneCluster, error) {
	if tfOutput != nil {
		tfConfig, err := terraformv1beta3.NewConfigFromJSON(tfOutput)
		if err != nil {
			return nil, err
		}
		if err := tfConfig.Apply(versionedCluster); err != nil {
			return nil, err
		}
	}

	internalCluster := &kubeoneapi.KubeOneCluster{}

	kubeonescheme.Scheme.Default(versionedCluster)
	if err := kubeonescheme.Scheme.Convert(versionedCluster, internalCluster, nil); err != nil {
		return nil, fail.Config(err, fmt.Sprintf("converting %s to internal object", versionedCluster.GroupVersionKind()))
	}

	return internalCluster, nil
}

// SetKubeOneClusterDynamicDefaults sets the dynamic defaults for a given KubeOneCluster object
func SetKubeOneClusterDynamicDefaults(cluster *kubeoneapi.KubeOneCluster, credentialsFile []byte, baseDir string) error {
	// Set the default cloud config
	SetDefaultsCloudConfig(cluster)

	if cluster.CertificateAuthority.File != "" && !filepath.IsAbs(cluster.CertificateAuthority.File) {
		cluster.CertificateAuthority.File = filepath.Join(baseDir, cluster.CertificateAuthority.File)
	}

	if cluster.CertificateAuthority.File != "" {
		buf, err := os.ReadFile(cluster.CertificateAuthority.File)
		if err != nil {
			return fail.ConfigValidation(err)
		}

		cluster.CertificateAuthority.Bundle = string(buf)
	}

	if cluster.CertificateAuthority.Bundle != "" {
		// Set this for backward compatibility with older addons
		cluster.CABundle = cluster.CertificateAuthority.Bundle
	}

	// Parse the credentials file
	credentials := make(map[string]string)

	if err := yaml.Unmarshal(credentialsFile, &credentials); err != nil {
		return fail.Config(err, "YAML unmarshalling credentials file")
	}

	// Source cloud-config from the credentials file if it's present
	if cc, ok := credentials["cloudConfig"]; ok {
		if cluster.CloudProvider.CloudConfig != "" {
			return fail.NewConfigError("dynamic cloud config", "found cloudConfig in credentials file, in addition to already set in the manifest")
		}

		cluster.CloudProvider.CloudConfig = cc
	}
	// Source csi-config from the credentials file if it's present
	if cc, ok := credentials["csiConfig"]; ok {
		cluster.CloudProvider.CSIConfig = cc
	}

	if ra, ok := credentials["registriesAuth"]; ok {
		if err := setRegistriesAuth(cluster, ra); err != nil {
			return err
		}
	}

	// Defaulting for LoggingConfig.
	// NB: We intentionally default here because LoggingConfig is not available in
	// the v1beta1 API. If we would default in the v1beta2 API instead, this value would
	// be empty when converting from v1beta1 to internal. This means that v1beta1 API
	// users would depend on default values provided by Docker/upstream, which are
	// different than our default values, so we want to avoid this.
	if cluster.LoggingConfig.ContainerLogMaxSize == "" {
		cluster.LoggingConfig.ContainerLogMaxSize = containerruntime.DefaultContainerLogMaxSize
	}
	if cluster.LoggingConfig.ContainerLogMaxFiles == 0 {
		cluster.LoggingConfig.ContainerLogMaxFiles = containerruntime.DefaultContainerLogMaxFiles
	}

	// Default the AssetsConfiguration internal API
	cluster.DefaultAssetConfiguration()

	// Copy MachineAnnotations to NodeAnnotations.
	// MachineAnnotations has been deprecated in favor of NodeAnnotations.
	// This is supposed to handle renaming of MachineAnnotations to
	// NodeAnnotations in non backwards-compatibility breaking way.
	for i, workerset := range cluster.DynamicWorkers {
		// NB: We don't want to allow both MachineAnnotations and NodeAnnotations
		// to be set, so we explicitly handle this scenario here and in validation.
		if len(workerset.Config.MachineAnnotations) > 0 && len(workerset.Config.NodeAnnotations) == 0 {
			cluster.DynamicWorkers[i].Config.NodeAnnotations = cluster.DynamicWorkers[i].Config.MachineAnnotations
			cluster.DynamicWorkers[i].Config.MachineAnnotations = nil
		}
	}

	return nil
}

// SetDefaultsCloudConfig sets default values for the CloudConfig field in the KubeOneCluster object.
// this function assigns a default cloud configuration.
func SetDefaultsCloudConfig(obj *kubeoneapi.KubeOneCluster) {
	if obj.CloudProvider.AWS != nil && obj.CloudProvider.External {
		if obj.CloudProvider.CloudConfig == "" {
			obj.CloudProvider.CloudConfig = defaultAWSCCMCloudConfig(obj.Name, obj.ClusterNetwork.IPFamily)
		}
	}
}

// defaultAWSCCMCloudConfig generates a default cloud configuration for AWS when using the Cloud Controller Manager (CCM).
// The configuration includes the Kubernetes cluster ID and optionally sets NodeIPFamilies based on the IPFamily setting.
func defaultAWSCCMCloudConfig(name string, ipFamily kubeoneapi.IPFamily) string {
	// Initialize the configuration with the global section and cluster ID.
	lines := []string{
		"[global]",
		fmt.Sprintf("KubernetesClusterID=%q", name),
	}

	// Set NodeIPFamilies based on the IP family configuration.
	switch ipFamily {
	case kubeoneapi.IPFamilyIPv4:
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv4"))
	case kubeoneapi.IPFamilyIPv6:
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv6"))
	case kubeoneapi.IPFamilyIPv4IPv6:
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv4"))
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv6"))
	case kubeoneapi.IPFamilyIPv6IPv4:
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv6"))
		lines = append(lines, fmt.Sprintf("NodeIPFamilies=%q", "ipv4"))
	}

	return strings.Join(lines, "\n")
}

func setRegistriesAuth(cluster *kubeoneapi.KubeOneCluster, buf string) error {
	var registriesAuth struct {
		runtime.TypeMeta                          `json:",inline"`
		kubeonev1beta2.ContainerRuntimeContainerd `json:",inline"`
	}

	if err := yaml.UnmarshalStrict([]byte(buf), &registriesAuth); err != nil {
		return fail.Config(err, "YAML unmarshal registriesAuth")
	}

	if registriesAuth.APIVersion != kubeonev1beta2.SchemeGroupVersion.String() {
		return fail.ConfigError{
			Op:  "registriesAuth apiVersion checking",
			Err: errors.Errorf("only %q apiVersion is supported", kubeonev1beta2.SchemeGroupVersion.String()),
		}
	}

	containerdConfigKind := reflect.TypeOf(registriesAuth.ContainerRuntimeContainerd).Name()
	if registriesAuth.Kind != containerdConfigKind {
		return fail.ConfigError{
			Op:  "registriesAuth kind checking",
			Err: errors.Errorf("only %q kind is supported", containerdConfigKind),
		}
	}

	if cluster.ContainerRuntime.Containerd == nil {
		return fail.ConfigError{
			Op:  "containerRuntime checking",
			Err: errors.Errorf(".ContainerRuntime.Containerd should be set"),
		}
	}

	if cluster.ContainerRuntime.Containerd.Registries == nil {
		cluster.ContainerRuntime.Containerd.Registries = map[string]kubeoneapi.ContainerdRegistry{}
	}

	for registryName, registryInfo := range registriesAuth.Registries {
		internalRegistry := cluster.ContainerRuntime.Containerd.Registries[registryName]
		internalRegistry.Auth = (*kubeoneapi.ContainerdRegistryAuthConfig)(registryInfo.Auth)
		cluster.ContainerRuntime.Containerd.Registries[registryName] = internalRegistry
	}

	return nil
}

func isDir(dirname string) bool {
	stat, statErr := os.Stat(dirname)

	return statErr == nil && stat.Mode().IsDir()
}

// checkClusterConfiguration checks clusters for usage of alpha, deprecated fields, flags, unrecommended features etc. and print a warning if any are found.
func checkClusterConfiguration(cluster kubeoneapi.KubeOneCluster, logger logrus.FieldLogger) {
	if cluster.CloudProvider.Nutanix != nil {
		logger.Warnf("Nutanix support is considered as alpha, so the implementation might be changed in the future")
		logger.Warnf("Nutanix support is planned to graduate to beta/stable in KubeOne 1.5+")
	}

	if cluster.CloudProvider.Vsphere != nil && !cluster.CloudProvider.External && len(cluster.CloudProvider.CSIConfig) > 0 {
		logger.Warnf(".cloudProvider.csiConfig is provided, but is ignored when used with the in-tree cloud provider")
	}

	checkFlagsAndFeatureGateOverrides(cluster, logger)
}

func checkFlagsAndFeatureGateOverrides(cluster kubeoneapi.KubeOneCluster, logger logrus.FieldLogger) {
	if cluster.ControlPlaneComponents != nil {
		logger.Warn(controlPlaneComponentsWarning)

		if cluster.ControlPlaneComponents.ControllerManager != nil {
			if cluster.ControlPlaneComponents.ControllerManager.Flags != nil || cluster.ControlPlaneComponents.ControllerManager.FeatureGates != nil {
				logger.Warnf(flagsAndFeatureGateOverridesWarning, ".controlPlaneComponents.controllerManager", "kube-controller-manager")
			}
		}
		if cluster.ControlPlaneComponents.Scheduler != nil {
			if cluster.ControlPlaneComponents.Scheduler.Flags != nil || cluster.ControlPlaneComponents.Scheduler.FeatureGates != nil {
				logger.Warnf(flagsAndFeatureGateOverridesWarning, ".controlPlaneComponents.scheduler", "kube-scheduler")
			}
		}
		if cluster.ControlPlaneComponents.APIServer != nil {
			if cluster.ControlPlaneComponents.APIServer.Flags != nil || cluster.ControlPlaneComponents.APIServer.FeatureGates != nil {
				logger.Warnf(flagsAndFeatureGateOverridesWarning, ".controlPlaneComponents.apiServer", "kube-apiserver")
			}
		}
	}
}
