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
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	kubeonescheme "github.com/kubermatic/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1alpha1 "github.com/kubermatic/kubeone/pkg/apis/kubeone/v1alpha1"
	"github.com/kubermatic/kubeone/pkg/apis/kubeone/validation"
	"github.com/kubermatic/kubeone/pkg/terraform"

	"k8s.io/apimachinery/pkg/runtime"
)

// SourceKubeOneClusterFromTerraformOutput sources information about the cluster from the Terraform output
func SourceKubeOneClusterFromTerraformOutput(terraformOutput []byte, cluster *kubeonev1alpha1.KubeOneCluster) error {
	tfConfig, err := terraform.NewConfigFromJSON(terraformOutput)
	if err != nil {
		return errors.Wrap(err, "failed to parse Terraform config")
	}
	return tfConfig.Apply(cluster)
}

// SetKubeOneClusterDynamicDefaults sets the dynamic defaults for a given KubeOneCluster object
func SetKubeOneClusterDynamicDefaults(cfg *kubeoneapi.KubeOneCluster, credentialsFile []byte) error {
	// Parse the credentials file
	credentials := make(map[string]string)
	err := yaml.Unmarshal(credentialsFile, &credentials)
	if err != nil {
		return errors.Wrap(err, "unable to convert credentials file to yaml")
	}

	// Cloud-Config
	if cc, ok := credentials["cloudConfig"]; ok {
		cfg.CloudProvider.CloudConfig = cc
	}

	return nil
}

// DefaultedKubeOneCluster converts a versioned KubeOneCluster object to an internal representation of KubeOneCluster
// object while sourcing information from Terraform output, applying default values and validating the KubeOneCluster
// object
func DefaultedKubeOneCluster(versionedCluster *kubeonev1alpha1.KubeOneCluster, tfOutput, credentialsFile []byte) (*kubeoneapi.KubeOneCluster, error) {
	internalCfg := &kubeoneapi.KubeOneCluster{}

	if tfOutput != nil {
		if err := SourceKubeOneClusterFromTerraformOutput(tfOutput, versionedCluster); err != nil {
			return nil, errors.Wrap(err, "unable to source information about cluster from a given terraform output")
		}
	}

	// Default and convert to the internal API type
	kubeonescheme.Scheme.Default(versionedCluster)
	if err := kubeonescheme.Scheme.Convert(versionedCluster, internalCfg, nil); err != nil {
		return nil, errors.Wrap(err, "unable to convert versioned to internal cluster object")
	}

	// Apply the dynamic defaults
	err := SetKubeOneClusterDynamicDefaults(internalCfg, credentialsFile)
	if err != nil {
		return nil, err
	}

	// Validate the configuration
	if err := validation.ValidateKubeOneCluster(*internalCfg).ToAggregate(); err != nil {
		return nil, errors.Wrap(err, "unable to validate the given KubeOneCluster object")
	}

	return internalCfg, nil
}

// LoadKubeOneCluster returns the KubeOneCluster object parsed from the KubeOneCluster configuration file and
// optionally Terraform output
func LoadKubeOneCluster(clusterCfgPath, tfOutputPath, credentialsFilePath string, logger *logrus.Logger) (*kubeoneapi.KubeOneCluster, error) {
	if len(clusterCfgPath) == 0 {
		return nil, errors.New("cluster configuration path not provided")
	}

	cluster, err := ioutil.ReadFile(clusterCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read the given cluster configuration file")
	}

	var tfOutput []byte

	switch {
	case tfOutputPath == "-":
		if tfOutput, err = ioutil.ReadAll(os.Stdin); err != nil {
			return nil, errors.Wrap(err, "unable to read terraform output from stdin")
		}
	case isDir(tfOutputPath):
		cmd := exec.Command("terraform", "output", "-json")
		cmd.Dir = tfOutputPath
		logger.Debugln("Executing `terraform output -json` to query terraform state")
		if tfOutput, err = cmd.Output(); err != nil {
			return nil, errors.Wrapf(err, "unable to read terraform output from the %q directory", tfOutputPath)
		}
	case len(tfOutputPath) != 0:
		if tfOutput, err = ioutil.ReadFile(tfOutputPath); err != nil {
			return nil, errors.Wrap(err, "unable to read the given terraform output file")
		}
	}

	var credentialsFile []byte
	if len(credentialsFilePath) != 0 {
		credentialsFile, err = ioutil.ReadFile(credentialsFilePath)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read the given credentials file")
		}
	}

	return BytesToKubeOneCluster(cluster, tfOutput, credentialsFile)
}

func isDir(dirname string) bool {
	stat, statErr := os.Stat(dirname)
	return statErr == nil && stat.Mode().IsDir()
}

// BytesToKubeOneCluster returns the KubeOneCluster object parsed from the KubeOneCluster manifest and optionally
// Terraform output
func BytesToKubeOneCluster(cluster, tfOutput, credentialsFile []byte) (*kubeoneapi.KubeOneCluster, error) {
	initCfg := &kubeonev1alpha1.KubeOneCluster{}
	if err := runtime.DecodeInto(kubeonescheme.Codecs.UniversalDecoder(), cluster, initCfg); err != nil {
		return nil, err
	}

	return DefaultedKubeOneCluster(initCfg, tfOutput, credentialsFile)
}
