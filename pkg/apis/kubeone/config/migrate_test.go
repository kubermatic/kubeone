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
	"flag"
	"testing"

	"k8c.io/kubeone/pkg/testhelper"
)

var updateFlag = flag.Bool("update", false, "update testdata files")

func TestV1Beta2ToV1Beta3Migration(t *testing.T) {
	tests := []string{
		"simple",
		"just addons",
		"helm",
		"addons and helm",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			got, err := MigrateV1beta2V1beta3(testhelper.TestDataFSName(t, "_v1beta2.yaml"))
			if err != nil {
				t.Fatalf("%s", err)
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), string(got), *updateFlag)
		})
	}
}
