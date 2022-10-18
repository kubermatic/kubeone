/*
Copyright 2022 The KubeOne Authors.

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

package cmd

import (
	"errors"
	"flag"
	"sort"
	"testing"

	"k8c.io/kubeone/pkg/testhelper"
)

var updateFlag = flag.Bool("update", false, "update testdata files")

func Test_genKubeOneClusterYAML(t *testing.T) {
	type testArgs struct {
		providerName string
		err          error
	}

	validProvidersNames := []string{}
	for provName := range validProviders {
		validProvidersNames = append(validProvidersNames, provName)
	}
	sort.Strings(validProvidersNames)
	tests := []testArgs{}

	for _, provName := range validProvidersNames {
		tests = append(tests, testArgs{
			providerName: provName,
		})
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.providerName, func(t *testing.T) {
			params := &genKubeOneClusterYAMLParams{
				providerName:      tt.providerName,
				clusterName:       "example",
				kubernetesVersion: "v1.24.4",
				validProviders:    validProviders,
			}

			got, err := genKubeOneClusterYAML(params)
			if !errors.Is(err, tt.err) {
				t.Errorf("genKubeOneClusterYAML() unexpected error: %v, expected err %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), string(got), *updateFlag)
		})
	}
}
