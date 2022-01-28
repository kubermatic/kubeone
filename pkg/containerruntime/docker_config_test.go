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

package containerruntime

import (
	"encoding/json"
	"testing"

	"k8c.io/kubeone/pkg/apis/kubeone"
)

func Test_marshalDockerConfig(t *testing.T) {
	tests := []struct {
		name             string
		cluster          *kubeone.KubeOneCluster
		expectedMaxSize  string
		expectedMaxFiles string
	}{
		{
			name:             "Should be convert 100Mi to 100m",
			cluster:          genCluster(withContainerLogMaxSize("100Mi")),
			expectedMaxSize:  "100m",
			expectedMaxFiles: "0",
		},
		{
			name:             "Should be convert 100Ki to 100k",
			cluster:          genCluster(withContainerLogMaxSize("100Ki")),
			expectedMaxSize:  "100k",
			expectedMaxFiles: "0",
		},
		{
			name:             "Should be convert 100Gi to 100g",
			cluster:          genCluster(withContainerLogMaxSize("100Gi")),
			expectedMaxSize:  "100g",
			expectedMaxFiles: "0",
		},
		{
			name:             "Should set max-files to 10 and max-size to 100m",
			cluster:          genCluster(withContainerLogMaxSize("100m"), withContainerLogMaxFiles(10)),
			expectedMaxSize:  "100m",
			expectedMaxFiles: "10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalDockerConfig(tt.cluster)
			if err != nil {
				t.Errorf("marshalDockerConfig() error = %v,", err)
			}
			cfg := dockerConfig{}
			err = json.Unmarshal([]byte(got), &cfg)
			if err != nil {
				t.Errorf("marshalDockerConfig() error = %v,", err)
			}
			maxSize := cfg.LogOpts["max-size"]
			if maxSize != tt.expectedMaxSize {
				t.Errorf("marshalDockerConfig() got = %v, want %v", got, tt.expectedMaxSize)
			}

			maxFiles := cfg.LogOpts["max-files"]
			if maxFiles != tt.expectedMaxFiles {
				t.Errorf("marshalDockerConfig() got = %v, want %v", got, tt.expectedMaxFiles)
			}
		})
	}
}

func withContainerLogMaxSize(logSize string) clusterOpts {
	return func(cls *kubeone.KubeOneCluster) {
		cls.LoggingConfig.ContainerLogMaxSize = logSize
	}
}

func withContainerLogMaxFiles(logFiles int32) clusterOpts {
	return func(cls *kubeone.KubeOneCluster) {
		cls.LoggingConfig.ContainerLogMaxFiles = logFiles
	}
}
