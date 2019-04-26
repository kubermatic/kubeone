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
	"io/ioutil"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	kubeonescheme "github.com/kubermatic/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1alpha1 "github.com/kubermatic/kubeone/pkg/apis/kubeone/v1alpha1"
	"github.com/kubermatic/kubeone/pkg/apis/kubeone/validation"
	"github.com/kubermatic/kubeone/pkg/util/credentials"
)

// SetKubeOneClusterDynamicDefaults sets the dynamic defaults for a given KubeOneCluster object
func SetKubeOneClusterDynamicDefaults(cfg *kubeoneapi.KubeOneCluster) error {
	if err := SetKubeOneClusterCredentials(cfg); err != nil {
		return errors.Wrap(err, "unable to set dynamic defaults for a given KubeOneCluster object")
	}
	return nil
}

// SetKubeOneClusterCredentials populates credentials used for machine-controller and external CCM
func SetKubeOneClusterCredentials(cfg *kubeoneapi.KubeOneCluster) error {
	// Only populate credentials if machine-controller is deployed or cloud provider is external
	if (cfg.MachineController != nil && !cfg.MachineController.Deploy) && !cfg.CloudProvider.External {
		return nil
	}

	creds, err := credentials.ProviderCredentials(cfg.CloudProvider.Name)
	if err != nil {
		return errors.Wrap(err, "unable to fetch cloud provider credentials")
	}
	cfg.Credentials = creds

	return nil
}

// SourceKubeOneClusterFromTerraformOutput sources information about the cluster from the Terraform output
// func SourceKubeOneClusterFromTerraformOutput(terraformOutput []byte, cluster *kubeonev1alpha1.KubeOneCluster) error {
// 	var (
// 		tfConfig *terraform.Config
// 		err      error
// 	)

// 	if tfConfig, err = terraform.NewConfigFromJSON(terraformOutput); err != nil {
// 		return errors.Wrap(err, "failed to parse Terraform config")
// 	}

// 	return tfConfig.Apply(cluster)
// }

// DefaultedKubeOneCluster takes a versioned KubeOneCluster object and optionally a Terraform output, and converts versioned object to the
// internal API type, defaults and validates the object.
func DefaultedKubeOneCluster(versionedCluster *kubeonev1alpha1.KubeOneCluster, tfOutput []byte) (*kubeoneapi.KubeOneCluster, error) {
	internalCfg := &kubeoneapi.KubeOneCluster{}

	// if tfOutput != nil {
	// 	if err := SourceKubeOneClusterFromTerraformOutput(tfOutput, versionedCluster); err != nil {
	// 		return nil, errors.Wrap(err, "unable to source information about cluster from a given terraform output")
	// 	}
	// }

	// Default and convert to the internal API type
	kubeonescheme.Scheme.Default(versionedCluster)
	if err := kubeonescheme.Scheme.Convert(versionedCluster, internalCfg, nil); err != nil {
		return nil, errors.Wrap(err, "unable to convert versioned to internal cluster object")
	}

	// Apply the dynamic defaults
	if err := SetKubeOneClusterDynamicDefaults(internalCfg); err != nil {
		return nil, err
	}

	// Validate the configuration
	if err := validation.ValidateKubeOneCluster(*internalCfg).ToAggregate(); err != nil {
		return nil, errors.Wrap(err, "unable to validate the given KubeOneCluster object")
	}

	return internalCfg, nil
}

// LoadKubeOneCluster takes a KubeOneCluster manifests and optionally a Terraform output, and
// returns a KubeOneCluster object
func LoadKubeOneCluster(clusterCfgPath, tfOutputPath string) (*kubeoneapi.KubeOneCluster, error) {
	if len(clusterCfgPath) == 0 {
		return nil, errors.New("cluster configuration path not provided")
	}

	cluster, err := ioutil.ReadFile(clusterCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read the given cluster configuration file")
	}

	var tfOutput []byte
	if len(tfOutputPath) > 0 {
		tfOutput, err = ioutil.ReadFile(tfOutputPath)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read the given terraform output file")
		}
	}

	return BytesToKubeOneCluster(cluster, tfOutput)
}

// BytesToKubeOneCluster parses the bytes slice of cluster configuration and optionally terraform output
// and returns a KubeOneCluster object
func BytesToKubeOneCluster(cluster, tfOutput []byte) (*kubeoneapi.KubeOneCluster, error) {
	initCfg := &kubeonev1alpha1.KubeOneCluster{}
	if err := runtime.DecodeInto(kubeonescheme.Codecs.UniversalDecoder(), cluster, initCfg); err != nil {
		return nil, err
	}

	return DefaultedKubeOneCluster(initCfg, tfOutput)
}
