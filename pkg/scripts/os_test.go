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

package scripts

import (
	"errors"
	"testing"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/testhelper"
)

type genClusterOpts func(*kubeoneapi.KubeOneCluster)

func withKubeVersion(ver string) genClusterOpts {
	return func(cls *kubeoneapi.KubeOneCluster) {
		cls.Versions.Kubernetes = ver
	}
}

// func withNutanixCloudProvider(cls *kubeoneapi.KubeOneCluster) {
// 	cls.CloudProvider = kubeoneapi.CloudProviderSpec{
// 		Nutanix: &kubeoneapi.NutanixSpec{},
// 	}
// }

func withCiliumCNI(cls *kubeoneapi.KubeOneCluster) {
	cls.ClusterNetwork.CNI = &kubeoneapi.CNI{
		Cilium: &kubeoneapi.CiliumSpec{},
	}
}

func withProxy(proxy string) genClusterOpts {
	return func(cls *kubeoneapi.KubeOneCluster) {
		cls.Proxy.HTTPS = proxy
		cls.Proxy.HTTP = proxy
	}
}

func withRegistry(registry string) genClusterOpts {
	return func(cls *kubeoneapi.KubeOneCluster) {
		cls.RegistryConfiguration = &kubeoneapi.RegistryConfiguration{
			OverwriteRegistry: registry,
		}
	}
}

func withInsecureRegistry(registry string) genClusterOpts {
	return func(cls *kubeoneapi.KubeOneCluster) {
		cls.RegistryConfiguration = &kubeoneapi.RegistryConfiguration{
			OverwriteRegistry: registry,
			InsecureRegistry:  true,
		}
	}
}

func withDefaultAssetConfiguration(cls *kubeoneapi.KubeOneCluster) {
	cls.AssetConfiguration = kubeoneapi.AssetConfiguration{
		Kubernetes: kubeoneapi.ImageAsset{
			ImageRepository: "registry.k8s.io",
		},
		CNI: kubeoneapi.BinaryAsset{
			URL: "http://127.0.0.1/cni.tar.gz",
		},
		NodeBinaries: kubeoneapi.BinaryAsset{
			URL: "http://127.0.0.1/node.tar.gz",
		},
		Kubectl: kubeoneapi.BinaryAsset{
			URL: "http://127.0.0.1/kubectl.tar.gz",
		},
	}
}

func genCluster(opts ...genClusterOpts) kubeoneapi.KubeOneCluster {
	cls := &kubeoneapi.KubeOneCluster{
		Versions: kubeoneapi.VersionConfig{
			Kubernetes: "1.30.0",
		},
		ContainerRuntime: kubeoneapi.ContainerRuntimeConfig{
			Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
		},
		SystemPackages: &kubeoneapi.SystemPackages{
			ConfigureRepositories: true,
		},
		Proxy: kubeoneapi.ProxyConfig{
			HTTP:    "http://http.proxy",
			HTTPS:   "http://https.proxy",
			NoProxy: ".local",
		},
		LoggingConfig: kubeoneapi.LoggingConfig{
			ContainerLogMaxSize:  "100Mi",
			ContainerLogMaxFiles: 5,
		},
	}

	for _, fn := range opts {
		fn(cls)
	}

	return *cls
}

func TestMigrateToContainerd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		insecureRegistry string
		mirrors          []string
		osName           kubeoneapi.OperatingSystemName
		err              error
	}{
		{
			name: "simple",
		},
		{
			name:   "flatcar",
			osName: kubeoneapi.OperatingSystemNameFlatcar,
		},
		{
			name:             "insecureRegistry",
			insecureRegistry: "some.registry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cls := genCluster(
				withInsecureRegistry(tt.insecureRegistry),
			)

			got, err := MigrateToContainerd(&cls, &kubeoneapi.HostConfig{OperatingSystem: tt.osName})
			if !errors.Is(err, tt.err) {
				t.Errorf("MigrateToContainerd() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}
