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
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	yaml "gopkg.in/yaml.v2"

	kubeonev1alpha1 "github.com/kubermatic/kubeone/pkg/apis/kubeone/v1alpha1"

	kyaml "sigs.k8s.io/yaml"
)

var update = flag.Bool("update", false, "update .golden files")

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
			newConfigYAML, err := MigrateToKubeOneClusterAPI(filepath.Join("testdata", tc.name+".yaml"))
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

			compareOutput(t, tc.name, buffer.Bytes(), *update)
		})
	}
}

func compareOutput(t *testing.T, name string, output []byte, update bool) {
	golden, err := filepath.Abs(filepath.Join("testdata", name+".yaml.golden"))
	if err != nil {
		t.Fatalf("failed to get absolute path to goldan file: %v", err)
	}
	if update {
		if writeErr := ioutil.WriteFile(golden, output, 0644); writeErr != nil {
			t.Fatalf("failed to write updated fixture: %v", err)
		}
	}
	expected, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Fatalf("failed to read .golden file: %v", err)
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(expected)),
		B:        difflib.SplitLines(string(output)),
		FromFile: "Fixture",
		ToFile:   "Current",
		Context:  3,
	}
	diffStr, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		t.Fatal(err)
	}

	if diffStr != "" {
		t.Errorf("got diff between expected and actual result: \n%s\n", diffStr)
	}
}
