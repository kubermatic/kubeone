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
	"errors"
	"flag"
	"path/filepath"
	"testing"

	yaml "gopkg.in/yaml.v2"

	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/testhelper"

	kyaml "sigs.k8s.io/yaml"
)

var update = flag.Bool("update", false, "update .golden files")

func TestMigrateOldConfig(t *testing.T) {
	testcases := []struct {
		name string
		err  string
	}{
		{
			name: "config-addons-1",
		},
		{
			name: "config-addons-2",
		},
		{
			name: "config-addons-3",
		},
		{
			name: "config-addons-4",
		},
		{
			name: "config-addons-5",
		},
		{
			name: "config-addons-6",
		},
		{
			name: "config-assetconfig",
			err:  "the AssetConfiguration API has been removed from the v1beta2 API, please check the docs for information on how to migrate",
		},
		{
			name: "config-aws",
		},
		{
			name: "config-full",
		},
		{
			name: "config-packet",
		},
		{
			name: "config-podpresets-1",
		},
		{
			name: "config-podpresets-2",
		},
		{
			name: "config-podpresets-3",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			newConfigYAML, err := MigrateOldConfig(filepath.Join("testdata", tc.name+"-v1beta1.yaml"))
			if err != nil {
				errMsg := err.Error()

				var cfgErr fail.ConfigError
				if errors.As(err, &cfgErr) {
					errMsg = cfgErr.Err.Error()
				}

				if errMsg == tc.err {
					return
				}
				t.Errorf("error converting old config: %v", err)
			}

			// Convert new config to yaml
			var buffer bytes.Buffer
			err = yaml.NewEncoder(&buffer).Encode(newConfigYAML)
			if err != nil {
				t.Errorf("unable to decode yaml: %v", err)
			}

			// Validate new config by unmarshaling
			newConfig := kubeonev1beta2.NewKubeOneCluster()
			err = kyaml.UnmarshalStrict(buffer.Bytes(), &newConfig)
			if err != nil {
				t.Errorf("failed to decode new config: %v", err)
			}

			testhelper.DiffOutput(t, tc.name+"-v1beta2.golden", buffer.String(), *update)
		})
	}
}
