/*
Copyright 2020 The KubeOne Authors.

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
	"os"

	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	kubeonev1beta3 "k8c.io/kubeone/pkg/apis/kubeone/v1beta3"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/yamled"
)

// MigrateOldConfig migrates KubeOneCluster v1beta1 object to v1beta2
func MigrateOldConfig(clusterFilePath string) (interface{}, error) {
	oldConfig, err := loadClusterConfig(clusterFilePath)
	if err != nil {
		return nil, fail.Runtime(err, "loading cluster config to migrate")
	}

	// Check is kubeone.k8c.io/v1beta2 config provided
	apiVersion, apiVersionExists := oldConfig.GetString(yamled.Path{"apiVersion"})
	if !apiVersionExists {
		return nil, fail.Config(fmt.Errorf("apiVersion not present in the manifest"), "checking apiVersion presence")
	}

	if apiVersion != kubeonev1beta2.SchemeGroupVersion.String() {
		return nil, fail.Config(fmt.Errorf("migration is available only for %q API, but %q is given", kubeonev1beta2.SchemeGroupVersion.String(), apiVersion), "checking apiVersion compatibility")
	}

	// Ensure kind is KubeOneCluster
	kind, kindExists := oldConfig.GetString(yamled.Path{"kind"})
	if !kindExists {
		return nil, fail.ConfigValidation(fmt.Errorf("kind not present in the manifest"))
	}
	if kind != KubeOneClusterKind {
		return nil, fail.ConfigValidation(fmt.Errorf("migration is available only for kind %q, but %q is given", KubeOneClusterKind, kind))
	}

	// The APIVersion has been changed to kubeone.k8c.io/v1beta3
	oldConfig.Set(yamled.Path{"apiVersion"}, kubeonev1beta3.SchemeGroupVersion.String())

	return oldConfig.Root(), nil
}

// loadClusterConfig takes path to the Cluster Config (old API) and returns yamled.Document
func loadClusterConfig(oldConfigPath string) (*yamled.Document, error) {
	f, err := os.Open(oldConfigPath)
	if err != nil {
		return nil, fail.Runtime(err, "open manifest file")
	}
	defer f.Close()

	return yamled.Load(f)
}
