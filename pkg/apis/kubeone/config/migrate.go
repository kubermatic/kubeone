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
	"os"
	"reflect"

	"gopkg.in/yaml.v2"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	kubeonev1beta3 "k8c.io/kubeone/pkg/apis/kubeone/v1beta3"
	"k8c.io/kubeone/pkg/fail"
	terraformv1beta2 "k8c.io/kubeone/pkg/terraform/v1beta2"
	"k8c.io/kubeone/pkg/yamled"

	kyaml "sigs.k8s.io/yaml"
)

// MigrateV1beta2V1beta3 migrates KubeOneCluster v1beta2 object to v1beta3
func MigrateV1beta2V1beta3(clusterFilePath string, tfOutput []byte) ([]byte, error) {
	originalManifest, err := loadClusterConfig(clusterFilePath)
	if err != nil {
		return nil, fail.Runtime(err, "loading cluster config to migrate")
	}

	var (
		buffer                 bytes.Buffer
		v1beta2KubeOneCluster  = kubeonev1beta2.NewKubeOneCluster()
		v1beta3KubeOneCluster  = kubeonev1beta3.NewKubeOneCluster()
		internalKubeOneCluster = new(kubeoneapi.KubeOneCluster)
	)

	if err = yaml.NewEncoder(&buffer).Encode(originalManifest.Root()); err != nil {
		return nil, fail.Config(err, "marshaling v1beta2 KubeOneCluster")
	}

	if err = kyaml.UnmarshalStrict(buffer.Bytes(), v1beta2KubeOneCluster); err != nil {
		return nil, fail.Runtime(err, "testing unmarshal v1beta2 KubeOneCluster")
	}

	if tfOutput != nil {
		tfConfig, tferr := terraformv1beta2.NewConfigFromJSON(tfOutput)
		if tferr != nil {
			return nil, tferr
		}
		if err = tfConfig.Apply(v1beta2KubeOneCluster); err != nil {
			return nil, err
		}
	}

	if err = scheme.Scheme.Convert(v1beta2KubeOneCluster, internalKubeOneCluster, kubeoneapi.SchemeGroupVersion); err != nil {
		return nil, fail.Config(err, "converting v1beta2/KubeOneCluster into internal KubeOneCluster")
	}

	if err = scheme.Scheme.Convert(internalKubeOneCluster, v1beta3KubeOneCluster, kubeonev1beta3.SchemeGroupVersion); err != nil {
		return nil, fail.Config(err, "converting internal KubeOneCluster into v1beta3/KubeOneCluster")
	}

	conversionsTab := []struct {
		path      yamled.Path
		convertor func(yamled.Path)
	}{
		{
			path: yamled.Path{"apiVersion"},
			convertor: func(p yamled.Path) {
				originalManifest.Set(p, kubeonev1beta3.SchemeGroupVersion.String())
			},
		},
		{
			path: yamled.Path{"addons"},
			convertor: func(p yamled.Path) {
				// we moved helmReleases inside the addons
				if originalManifest.Has(p) || originalManifest.Has(yamled.Path{"helmReleases"}) {
					ybuf, _ := kyaml.Marshal(v1beta3KubeOneCluster.Addons)
					addons, _ := yamled.Load(bytes.NewBuffer(ybuf))
					originalManifest.Set(p, addons)
				}
			},
		},
		{
			path: yamled.Path{"addons"},
			convertor: func(p yamled.Path) {
				// cleanup addons from all the nil/zero/invalid values
				originalManifest.Walk(p, func(key yamled.Path, value any) {
					refval := reflect.ValueOf(value)

					//nolint:exhaustive
					switch refval.Kind() {
					case reflect.Pointer, reflect.Map, reflect.Slice:
						if refval.IsNil() {
							originalManifest.Remove(key)
						}
					case reflect.Invalid:
						originalManifest.Remove(key)
					default:
						if refval.IsZero() {
							originalManifest.Remove(key)
						}
					}
				})
			},
		},
		{
			path: yamled.Path{"helmReleases"},
			convertor: func(p yamled.Path) {
				originalManifest.Remove(p)
			},
		},
		{
			path: yamled.Path{"apiEndpoint"},
			convertor: func(p yamled.Path) {
				if v1beta2KubeOneCluster.APIEndpoint.Host == "" {
					if len(v1beta2KubeOneCluster.ControlPlane.Hosts) > 0 {
						defaultAPIEndpoint := v1beta2KubeOneCluster.ControlPlane.Hosts[0].PublicAddress
						if defaultAPIEndpoint == "" {
							defaultAPIEndpoint = v1beta2KubeOneCluster.ControlPlane.Hosts[0].PrivateAddress
						}
						originalManifest.Set(append(p, "host"), defaultAPIEndpoint)
					}
				}
			},
		},
	}

	for _, conv := range conversionsTab {
		conv.convertor(conv.path)
		var buf bytes.Buffer
		_ = yaml.NewEncoder(&buf).Encode(originalManifest)
		originalManifest, err = yamled.Load(&buf)
		if err != nil {
			return nil, err
		}
	}

	return yaml.Marshal(originalManifest)
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
