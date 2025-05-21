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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubeonescheme "k8c.io/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	kubeonev1beta3 "k8c.io/kubeone/pkg/apis/kubeone/v1beta3"
	kubeonevalidation "k8c.io/kubeone/pkg/apis/kubeone/validation"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/fail"
	terraformv1beta2 "k8c.io/kubeone/pkg/terraform/v1beta2"
	terraformv1beta3 "k8c.io/kubeone/pkg/terraform/v1beta3"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/providerconfig"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
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

	return BytesToKubeOneCluster(cluster, tfOutput, credentialsFilePath, logger, cfgBaseDir)
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		// TODO: replace terraform exec with direct file read and decode
		cmd := exec.CommandContext(ctx, "terraform", "output", "-json")
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
func BytesToKubeOneCluster(cluster, tfOutput []byte, credentialsFilePath string, logger logrus.FieldLogger, baseDir string) (*kubeoneapi.KubeOneCluster, error) {
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
	if err := SetKubeOneClusterDynamicDefaults(internalCluster, credentialsFilePath, baseDir); err != nil {
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
func SetKubeOneClusterDynamicDefaults(cluster *kubeoneapi.KubeOneCluster, credentialsFilePath string, baseDir string) error {
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
		cluster.CABundle = cluster.CertificateAuthority.Bundle //nolint:staticcheck
	}

	// Parse the credentials file
	credentials := make(map[string]string)

	if len(credentialsFilePath) != 0 {
		credentialsFile, err := os.ReadFile(credentialsFilePath)
		if err != nil {
			return fail.Runtime(err, "reading credentials file")
		}

		if err := yaml.Unmarshal(credentialsFile, &credentials); err != nil {
			return fail.Config(err, "YAML unmarshalling credentials file")
		}
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

	// Default the AssetsConfiguration internal API
	cluster.DefaultAssetConfiguration()

	var err error

	if cluster.ControlPlane.NodeSets != nil {
		// We have to partially validate early to be able to set defaults
		if err := kubeonevalidation.ValidateCloudProviderSpec(*cluster, field.NewPath("provider")).ToAggregate(); err != nil {
			return fail.ConfigValidation(err)
		}

		switch {
		case cluster.CloudProvider.Hetzner != nil:
			setDefaultHetznerControlPlane(cluster.Name, cluster.CloudProvider.Hetzner.ControlPlane)

			if cluster.APIEndpoint.Host == "" {
				cluster.APIEndpoint, err = getOrCreateHetznerLB(cluster, credentialsFilePath)
				if err != nil {
					return err
				}
			}

			machines, err := generateHetznerControlPlaneMachines(cluster.ControlPlane.NodeSets, cluster.Versions.Kubernetes)
			if err != nil {
				return err
			}
			_ = machines

			// * find or create []Machines using machine controller libs
			// * generate cluster.ControlPlane.Hosts based on []Machines
		default:
			return fail.ConfigError{
				Op:  "cloud provider checking",
				Err: errors.New("configured cloud provider does not support managed control plane nodes"),
			}
		}
	}

	return nil
}

func setDefaultHetznerControlPlane(clusterName string, hzCP *kubeoneapi.HetznerControlPlane) {
	hzCP.LoadBalancer.Name = defaults(
		hzCP.LoadBalancer.Name,
		clusterName+"-kubeapi",
	)
	hzCP.LoadBalancer.Type = defaults(
		hzCP.LoadBalancer.Type,
		"lb11",
	)
	hzCP.LoadBalancer.Location = defaults(
		hzCP.LoadBalancer.Location,
		"nbg1",
	)
	hzCP.LoadBalancer.PublicIP = defaults(
		hzCP.LoadBalancer.PublicIP,
		ptr.To(true),
	)
	if hzCP.LoadBalancer.Labels == nil {
		hzCP.LoadBalancer.Labels = map[string]string{}
	}
	maps.Copy(
		hzCP.LoadBalancer.Labels,
		map[string]string{
			"kubeone_cluster_name": hzCP.LoadBalancer.Name,
			"kubeone_role":         "api",
		},
	)
}

func generateHetznerControlPlaneMachines(nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	var machines []clusterv1alpha1.Machine

	for _, node := range nodeSet {
		for idx := range node.Replicas {
			osSpecRaw, err := json.Marshal(node.OperatingSystemSpec)
			if err != nil {
				return nil, err
			}

			providerSpec := providerconfig.Config{
				SSHPublicKeys: node.SSH.PublicKeys,
				CloudProvider: providerconfig.CloudProviderHetzner,
				CloudProviderSpec: runtime.RawExtension{
					Raw: node.CloudProviderSpec,
				},
				OperatingSystem: providerconfig.OperatingSystem(node.OperatingSystem),
				OperatingSystemSpec: runtime.RawExtension{
					Raw: osSpecRaw,
				},
			}

			providerSpecRaw, err := json.Marshal(providerSpec)
			if err != nil {
				return nil, err
			}

			machines = append(machines, clusterv1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%d", node.Name, idx),
				},
				Spec: clusterv1alpha1.MachineSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      node.NodeSettings.Labels,
						Annotations: node.NodeSettings.Annotations,
					},
					Taints: node.NodeSettings.Taints,
					Versions: clusterv1alpha1.MachineVersionInfo{
						Kubelet: kubeletVersion,
					},
					ProviderSpec: clusterv1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{
							Raw: providerSpecRaw,
						},
					},
				},
			})
		}
	}

	return machines, nil
}

func getOrCreateHetznerLB(cluster *kubeoneapi.KubeOneCluster, credentialsFilePath string) (kubeoneapi.APIEndpoint, error) {
	if cluster.APIEndpoint.Host != "" {
		return cluster.APIEndpoint, nil
	}

	providerCreds, err := credentials.ProviderCredentials(cluster.CloudProvider, credentialsFilePath, credentials.TypeMC)
	if err != nil {
		return cluster.APIEndpoint, err
	}

	hzclient := hcloud.NewClient(hcloud.WithToken(providerCreds["HCLOUD_TOKEN"]))
	var networkID int64

	networkName := cluster.CloudProvider.Hetzner.NetworkID
	ctx := context.Background()

	networks, _, err := hzclient.Network.List(ctx, hcloud.NetworkListOpts{
		Name: networkName,
	})
	if err != nil {
		return cluster.APIEndpoint, fail.Config(err, "listing hetzner networks")
	}

	if len(networks) == 0 {
		return cluster.APIEndpoint, fail.Config(fmt.Errorf("no network ID found with ID: %s", networkName), "looking up hetzner network")
	}

	networkID = networks[0].ID

	clusterLBName := cluster.CloudProvider.Hetzner.ControlPlane.LoadBalancer.Name
	lbs, _, err := hzclient.LoadBalancer.List(ctx, hcloud.LoadBalancerListOpts{
		Name: clusterLBName,
	})
	if err != nil {
		return cluster.APIEndpoint, fail.Config(err, "listing hetzner loadbalancers")
	}

	var realLB *hcloud.LoadBalancer

	if len(lbs) > 0 {
		log.Printf("ℹ️: loadbalancer already exists with id: %d", lbs[0].ID)
		realLB = lbs[0]
	} else {
		log.Printf("⚠️: no existing loadbalancer found, creating a new one")
		lb, err := createLoadBalancer(
			ctx,
			&hzclient.LoadBalancer,
			cluster,
			networkID,
		)
		if err != nil {
			return cluster.APIEndpoint, fail.Config(err, "creating hetzner loadbalancer")
		}
		realLB = lb
	}

	cluster.APIEndpoint.Host = realLB.PublicNet.IPv4.IP.String()
	cluster.APIEndpoint.Port = 6443

	return cluster.APIEndpoint, nil
}

func createLoadBalancer(
	ctx context.Context,
	client hcloud.ILoadBalancerClient,
	cluster *kubeoneapi.KubeOneCluster,
	networkID int64,
) (*hcloud.LoadBalancer, error) {
	vmsLabelSelector := "kubeone_cluster_name=" + cluster.Name + ",role=api"
	now := time.Now().UTC()
	timestamp := strconv.FormatInt(now.Unix(), 10)
	newLabels := map[string]string{
		"kubeone_own_since_timestamp": timestamp,
	}
	hzlbSpec := cluster.CloudProvider.Hetzner.ControlPlane.LoadBalancer
	labels := make(map[string]string)
	maps.Copy(labels, hzlbSpec.Labels)
	maps.Copy(newLabels, labels)

	createReq := hcloud.LoadBalancerCreateOpts{
		Name:             hzlbSpec.Name,
		LoadBalancerType: &hcloud.LoadBalancerType{Name: hzlbSpec.Type},
		Location:         &hcloud.Location{Name: hzlbSpec.Location},
		Labels:           labels,
		PublicInterface:  hzlbSpec.PublicIP,
		Services: []hcloud.LoadBalancerCreateOptsService{
			{
				Protocol:        hcloud.LoadBalancerServiceProtocolTCP,
				ListenPort:      hcloud.Ptr(6443),
				DestinationPort: hcloud.Ptr(6443),
			},
		},
		Targets: []hcloud.LoadBalancerCreateOptsTarget{
			{
				UsePrivateIP: hcloud.Ptr(true),
				Type:         hcloud.LoadBalancerTargetTypeLabelSelector,
				LabelSelector: hcloud.LoadBalancerCreateOptsTargetLabelSelector{
					Selector: vmsLabelSelector,
				},
			},
		},
		Network: &hcloud.Network{ID: networkID},
	}

	result, _, err := client.Create(ctx, createReq)
	if err != nil {
		return nil, err
	}

	return result.LoadBalancer, nil
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

func defaults[T comparable](input, defaultValue T) T {
	var zero T

	if input != zero {
		return input
	}

	return defaultValue
}
