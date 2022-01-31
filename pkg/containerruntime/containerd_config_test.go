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

package containerruntime

import (
	"flag"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/testhelper"
)

var (
	updateFlag = flag.Bool("update", false, "update testdata files")
)

func Test_marshalContainerdConfig(t *testing.T) {
	tests := []struct {
		name    string
		cluster *kubeoneapi.KubeOneCluster
	}{
		{
			name:    "simple",
			cluster: genCluster(),
		},
		{
			name: "override insecure registry",
			cluster: genCluster(withRegistryConfiguration(kubeoneapi.RegistryConfiguration{
				OverwriteRegistry: "some.registry",
				InsecureRegistry:  true,
			})),
		},
		{
			name: "docker.io mirror registry",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"docker.io": {
					Mirrors: []string{"https://custom.secure.registry"},
				},
			})),
		},
		{
			name: "multi registry mirrors",
			cluster: genCluster(withContainerdRegistry(map[string]kubeoneapi.ContainerdRegistry{
				"k8s.gcr.io": {
					Mirrors: []string{"https://some"},
				},
				"*": {
					Mirrors: []string{"https://custom.insecure.registry"},
					TLSConfig: &kubeoneapi.ContainerdTLSConfig{
						InsecureSkipVerify: true,
					},
				},
			})),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalContainerdConfig(tt.cluster)
			if err != nil {
				t.Errorf("marshalContainerdConfig() error = %v", err)
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

type clusterOpts func(*kubeoneapi.KubeOneCluster)

func genCluster(opts ...clusterOpts) *kubeoneapi.KubeOneCluster {
	cls := &kubeoneapi.KubeOneCluster{
		RegistryConfiguration: nil,
		ContainerRuntime: kubeoneapi.ContainerRuntimeConfig{
			Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
		},
	}

	for _, o := range opts {
		o(cls)
	}

	return cls
}

func withRegistryConfiguration(regCfg kubeoneapi.RegistryConfiguration) clusterOpts {
	return func(cls *kubeoneapi.KubeOneCluster) {
		cls.RegistryConfiguration = &regCfg
	}
}

func withContainerdRegistry(regCfg map[string]kubeoneapi.ContainerdRegistry) clusterOpts {
	return func(cls *kubeoneapi.KubeOneCluster) {
		cls.ContainerRuntime.Containerd.Registries = regCfg
	}
}
