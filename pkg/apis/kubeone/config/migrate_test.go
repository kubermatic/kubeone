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
	"path/filepath"
	"testing"

	yaml "gopkg.in/yaml.v2"

	kubeonev1beta1 "k8c.io/kubeone/pkg/apis/kubeone/v1beta1"
	"k8c.io/kubeone/pkg/testhelper"

	kyaml "sigs.k8s.io/yaml"
)

var update = flag.Bool("update", false, "update .golden files")

func TestMigrateOldConfig(t *testing.T) {
	testcases := []struct {
		name string
	}{
		{
			name: "config-aws",
		},
		{
			name: "config-azure",
		},
		{
			name: "config-digitalocean",
		},
		{
			name: "config-gce",
		},
		{
			name: "config-hetzner",
		},
		{
			name: "config-hetzner-networkid",
		},
		{
			name: "config-openstack",
		},
		{
			name: "config-packet",
		},
		{
			name: "config-vsphere",
		},
		{
			name: "config-none",
		},
		{
			name: "config-full",
		},
		{
			name: "config-full-example",
		},
		{
			name: "config-cni-canal",
		},
		{
			name: "config-cni-canal-encrypted",
		},
		{
			name: "config-cni-weave-net",
		},
		{
			name: "config-cni-weave-net-encrypted",
		},
		{
			name: "config-cni-external",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			newConfigYAML, err := MigrateOldConfig(filepath.Join("testdata", tc.name+"-v1alpha1.yaml"))
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
			newConfig := &kubeonev1beta1.KubeOneCluster{}
			err = kyaml.UnmarshalStrict(buffer.Bytes(), &newConfig)
			if err != nil {
				t.Errorf("failed to decode new config: %v", err)
			}

			testhelper.DiffOutput(t, tc.name+"-v1beta1.golden", buffer.String(), *update)
		})
	}
}
