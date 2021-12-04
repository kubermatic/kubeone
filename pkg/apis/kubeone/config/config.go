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
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubeonescheme "k8c.io/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1beta1 "k8c.io/kubeone/pkg/apis/kubeone/v1beta1"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	kubeonevalidation "k8c.io/kubeone/pkg/apis/kubeone/validation"
	terraformv1beta1 "k8c.io/kubeone/pkg/terraform/v1beta1"
	terraformv1beta2 "k8c.io/kubeone/pkg/terraform/v1beta2"

	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// KubeOneClusterKind is kind of the KubeOneCluster object
	KubeOneClusterKind = "KubeOneCluster"
)

var (
	// AllowedAPIs contains APIs which are allowed to be used
	AllowedAPIs = map[string]string{
		kubeonev1beta1.SchemeGroupVersion.String(): "",
		kubeonev1beta2.SchemeGroupVersion.String(): "",
	}

	// DeprecatedAPIs contains APIs which are deprecated
	DeprecatedAPIs = map[string]string{
		kubeonev1beta1.SchemeGroupVersion.String(): "",
	}
)

// LoadKubeOneCluster returns the internal representation of the KubeOneCluster object
// parsed from the versioned KubeOneCluster manifest, Terraform output and credentials file
func LoadKubeOneCluster(clusterCfgPath, tfOutputPath, credentialsFilePath string, logger logrus.FieldLogger) (*kubeoneapi.KubeOneCluster, error) {
	if len(clusterCfgPath) == 0 {
		return nil, errors.New("cluster configuration path not provided")
	}

	cluster, err := os.ReadFile(clusterCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read the given cluster configuration file")
	}

	var tfOutput []byte

	switch {
	case tfOutputPath == "-":
		if tfOutput, err = io.ReadAll(os.Stdin); err != nil {
			return nil, errors.Wrap(err, "unable to read terraform output from stdin")
		}
	case isDir(tfOutputPath):
		cmd := exec.Command("terraform", "output", "-json")
		cmd.Dir = tfOutputPath
		if tfOutput, err = cmd.Output(); err != nil {
			return nil, errors.Wrapf(err, "unable to read terraform output from the %q directory", tfOutputPath)
		}
	case len(tfOutputPath) != 0:
		if tfOutput, err = os.ReadFile(tfOutputPath); err != nil {
			return nil, errors.Wrap(err, "unable to read the given terraform output file")
		}
	}

	var credentialsFile []byte
	if len(credentialsFilePath) != 0 {
		credentialsFile, err = os.ReadFile(credentialsFilePath)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read the given credentials file")
		}
	}

	return BytesToKubeOneCluster(cluster, tfOutput, credentialsFile, logger)
}

// BytesToKubeOneCluster parses the bytes of the versioned KubeOneCluster manifests
func BytesToKubeOneCluster(cluster, tfOutput, credentialsFile []byte, logger logrus.FieldLogger) (*kubeoneapi.KubeOneCluster, error) {
	// Get the GVK from the given KubeOneCluster manifest
	typeMeta := runtime.TypeMeta{}
	if err := yaml.Unmarshal(cluster, &typeMeta); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal cluster typeMeta")
	}
	if len(typeMeta.APIVersion) == 0 || len(typeMeta.Kind) == 0 {
		return nil, errors.New("apiVersion and kind must be present in the manifest")
	}
	if typeMeta.Kind != KubeOneClusterKind {
		return nil, errors.Errorf("provided object %q is not KubeOneCluster object", typeMeta.Kind)
	}
	if _, ok := AllowedAPIs[typeMeta.APIVersion]; !ok {
		return nil, errors.Errorf("provided apiVersion %q is not supported", typeMeta.APIVersion)
	}
	if _, ok := DeprecatedAPIs[typeMeta.APIVersion]; ok {
		logger.Warningf("The provided APIVersion %q is deprecated. Please use \"kubeone config migrate\" command to migrate to the latest version.", typeMeta.APIVersion)
	}

	// Parse the cluster bytes depending on the GVK
	switch typeMeta.APIVersion {
	case kubeonev1beta1.SchemeGroupVersion.String():
		v1beta1Cluster := &kubeonev1beta1.KubeOneCluster{}
		if err := runtime.DecodeInto(kubeonescheme.Codecs.UniversalDecoder(), cluster, v1beta1Cluster); err != nil {
			return nil, err
		}
		return DefaultedV1Beta1KubeOneCluster(v1beta1Cluster, tfOutput, credentialsFile, logger)
	case kubeonev1beta2.SchemeGroupVersion.String():
		v1beta2Cluster := &kubeonev1beta2.KubeOneCluster{}
		if err := runtime.DecodeInto(kubeonescheme.Codecs.UniversalDecoder(), cluster, v1beta2Cluster); err != nil {
			return nil, err
		}
		return DefaultedV1Beta2KubeOneCluster(v1beta2Cluster, tfOutput, credentialsFile, logger)
	default:
		return nil, errors.Errorf("invalid api version %q", typeMeta.APIVersion)
	}
}

// DefaultedV1Beta1KubeOneCluster converts a v1beta1 KubeOneCluster object to an internal representation of KubeOneCluster
// object while sourcing information from Terraform output, applying default values and validating the KubeOneCluster
// object
func DefaultedV1Beta1KubeOneCluster(versionedCluster *kubeonev1beta1.KubeOneCluster, tfOutput, credentialsFile []byte, logger logrus.FieldLogger) (*kubeoneapi.KubeOneCluster, error) {
	if tfOutput != nil {
		tfConfig, err := terraformv1beta1.NewConfigFromJSON(tfOutput)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse Terraform config")
		}
		if err := tfConfig.Apply(versionedCluster); err != nil {
			return nil, errors.Wrap(err, "failed to apply Terraform config to the KubeOneCluster object")
		}
	}

	internalCluster := &kubeoneapi.KubeOneCluster{}

	kubeonescheme.Scheme.Default(versionedCluster)
	if err := kubeonescheme.Scheme.Convert(versionedCluster, internalCluster, nil); err != nil {
		return nil, errors.Wrap(err, "failed to convert versioned cluster object to internal object")
	}

	// Apply the dynamic defaults
	err := SetKubeOneClusterDynamicDefaults(internalCluster, credentialsFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply dynamic defaults")
	}

	// Validate the configuration
	if err := kubeonevalidation.ValidateKubeOneCluster(*internalCluster).ToAggregate(); err != nil {
		return nil, errors.Wrap(err, "unable to validate the given KubeOneCluster object")
	}

	// Check for deprecated fields/features for a cluster
	checkClusterForDeprecations(*internalCluster, logger)

	return internalCluster, nil
}

// DefaultedV1Beta2KubeOneCluster converts a v1beta2 KubeOneCluster object to an internal representation of KubeOneCluster
// object while sourcing information from Terraform output, applying default values and validating the KubeOneCluster
// object
func DefaultedV1Beta2KubeOneCluster(versionedCluster *kubeonev1beta2.KubeOneCluster, tfOutput, credentialsFile []byte, logger logrus.FieldLogger) (*kubeoneapi.KubeOneCluster, error) {
	if tfOutput != nil {
		tfConfig, err := terraformv1beta2.NewConfigFromJSON(tfOutput)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse Terraform config")
		}
		if err := tfConfig.Apply(versionedCluster); err != nil {
			return nil, errors.Wrap(err, "failed to apply Terraform config to the KubeOneCluster object")
		}
	}

	internalCluster := &kubeoneapi.KubeOneCluster{}

	kubeonescheme.Scheme.Default(versionedCluster)
	if err := kubeonescheme.Scheme.Convert(versionedCluster, internalCluster, nil); err != nil {
		return nil, errors.Wrap(err, "failed to convert versioned cluster object to internal object")
	}

	// Apply the dynamic defaults
	err := SetKubeOneClusterDynamicDefaults(internalCluster, credentialsFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply dynamic defaults")
	}

	// Validate the configuration
	if err := kubeonevalidation.ValidateKubeOneCluster(*internalCluster).ToAggregate(); err != nil {
		return nil, errors.Wrap(err, "unable to validate the given KubeOneCluster object")
	}

	// Check for deprecated fields/features for a cluster
	checkClusterForDeprecations(*internalCluster, logger)

	return internalCluster, nil
}

// SetKubeOneClusterDynamicDefaults sets the dynamic defaults for a given KubeOneCluster object
func SetKubeOneClusterDynamicDefaults(cfg *kubeoneapi.KubeOneCluster, credentialsFile []byte) error {
	// Parse the credentials file
	credentials := make(map[string]string)
	err := yaml.Unmarshal(credentialsFile, &credentials)
	if err != nil {
		return errors.Wrap(err, "unable to convert credentials file to yaml")
	}

	// Source cloud-config from the credentials file if it's present
	if cc, ok := credentials["cloudConfig"]; ok {
		cfg.CloudProvider.CloudConfig = cc
	}

	return nil
}

func isDir(dirname string) bool {
	stat, statErr := os.Stat(dirname)
	return statErr == nil && stat.Mode().IsDir()
}

// checkClusterForDeprecations with check clusters for usage of deprecated fields, flags etc. and print a warning if any are found
func checkClusterForDeprecations(c kubeoneapi.KubeOneCluster, logger logrus.FieldLogger) {
	if c.Features.PodSecurityPolicy != nil && c.Features.PodSecurityPolicy.Enable {
		logger.Warnf("PodSecurityPolicy is deprecated and will be removed with Kubernetes 1.25 release")
	}
}
