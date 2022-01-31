/*
Copyright 2021 The KubeOne Authors.

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

import "testing"

func TestRunPrint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		printOptions *printOpts
	}{
		{
			name: "simple config manifest",
			printOptions: &printOpts{
				ClusterName:       "test-cluster",
				KubernetesVersion: "v1.21.0",
				CloudProviderName: "aws",
			},
		},
		{
			name: "full config manifest",
			printOptions: &printOpts{
				FullConfig:        true,
				ClusterName:       "test-cluster",
				KubernetesVersion: "v1.21.0",
				CloudProviderName: "aws",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := runPrint(tt.printOptions)
			if err != nil {
				t.Fatalf("Error generating example manifest = %v", err)
			}
		})
	}
}
