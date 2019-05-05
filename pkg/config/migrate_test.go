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
	"bytes"
	"io/ioutil"
	"testing"

	yaml "gopkg.in/yaml.v2"

	kubeonev1alpha1 "github.com/kubermatic/kubeone/pkg/apis/kubeone/v1alpha1"

	kyaml "sigs.k8s.io/yaml"
)

func TestMigrateToKubeOneClusterAPI(t *testing.T) {
	testcases := []struct {
		name string
	}{
		{
			name: "config-aws",
		},
		{
			name: "config-do-external",
		},
		{
			name: "config-os-cloudconfig",
		},
		{
			name: "config-hosts",
		},
		{
			name: "config-features",
		},
		{
			name: "config-features-0.5.0",
		},
		{
			name: "config-apiserver",
		},
		{
			name: "config-network",
		},
		{
			name: "config-proxy",
		},
		{
			name: "config-machine-controller",
		},
		{
			name: "config-workers",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			newConfigYAML, err := MigrateToKubeOneClusterAPI("testcases/" + tc.name + ".yaml")
			if err != nil {
				t.Errorf("error converting old config: %v", err)
			}

			// Convert new config to yaml
			var buffer bytes.Buffer
			err = yaml.NewEncoder(&buffer).Encode(newConfigYAML)
			if err != nil {
				t.Errorf("unable to decode yaml: %v", err)
			}

			// Validate new config by unmarshaling
			newConfig := &kubeonev1alpha1.KubeOneCluster{}
			err = kyaml.UnmarshalStrict(buffer.Bytes(), &newConfig)
			if err != nil {
				t.Errorf("failed to decode new config: %v", err)
			}

			// Read expected config from a .golden file
			expectedConfigBytes, err := ioutil.ReadFile("testcases/" + tc.name + ".yaml.golden")
			if err != nil {
				t.Errorf("error reading golden file: %v", err)
			}

			// TODO(xmudrii): Can we use testify here?
			// Compare new and expected configs
			if !bytes.Equal(buffer.Bytes(), expectedConfigBytes) {
				t.Error("mismatch between new and expected configs")
			}
		})
	}
}
