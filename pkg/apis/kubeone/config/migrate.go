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
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	kubeonev1beta3 "k8c.io/kubeone/pkg/apis/kubeone/v1beta3"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/yamled"

	kyaml "sigs.k8s.io/yaml"
)

// MigrateV1beta2V1beta3 migrates KubeOneCluster v1beta2 object to v1beta3
func MigrateV1beta2V1beta3(clusterFilePath string) (*kubeonev1beta3.KubeOneCluster, error) {
	doc, err := loadClusterConfig(clusterFilePath)
	if err != nil {
		return nil, fail.Runtime(err, "loading cluster config to migrate")
	}

	// Check is kubeone.k8c.io/v1beta2 config provided
	oldAPIVersion, apiVersionExists := doc.GetString(yamled.Path{"apiVersion"})
	if !apiVersionExists {
		return nil, fail.Config(fmt.Errorf("apiVersion not present in the manifest"), "checking apiVersion presence")
	}

	if oldAPIVersion != kubeonev1beta2.SchemeGroupVersion.String() {
		return nil, fail.Config(fmt.Errorf("migration is available only for %q API, but %q is given", kubeonev1beta2.SchemeGroupVersion.String(), oldAPIVersion), "checking apiVersion compatibility")
	}

	// Ensure kind is KubeOneCluster
	kind, kindExists := doc.GetString(yamled.Path{"kind"})
	if !kindExists {
		return nil, fail.ConfigValidation(fmt.Errorf("kind not present in the manifest"))
	}
	if kind != KubeOneClusterKind {
		return nil, fail.ConfigValidation(fmt.Errorf("migration is available only for kind %q, but %q is given", KubeOneClusterKind, kind))
	}

	var (
		buffer           bytes.Buffer
		oldManifest      = kubeonev1beta2.NewKubeOneCluster()
		newManifest      = kubeonev1beta3.NewKubeOneCluster()
		internalManifest = new(kubeoneapi.KubeOneCluster)
	)

	if err = yaml.NewEncoder(&buffer).Encode(doc.Root()); err != nil {
		return nil, fail.Config(err, "marshaling v1beta2 KubeOneCluster")
	}

	if err = kyaml.UnmarshalStrict(buffer.Bytes(), oldManifest); err != nil {
		return nil, fail.Runtime(err, "testing unmarshal v1beta2 KubeOneCluster")
	}

	if err = scheme.Scheme.Convert(oldManifest, internalManifest, kubeoneapi.SchemeGroupVersion); err != nil {
		return nil, fail.Config(err, "converting v1beta2/KubeOneCluster into internal KubeOneCluster")
	}

	if err = scheme.Scheme.Convert(internalManifest, newManifest, kubeonev1beta3.SchemeGroupVersion); err != nil {
		return nil, fail.Config(err, "converting internal KubeOneCluster into v1beta3/KubeOneCluster")
	}

	return newManifest, nil
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
