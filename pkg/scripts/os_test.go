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

func withContainerd(cls *kubeoneapi.KubeOneCluster) {
	cls.ContainerRuntime.Containerd = &kubeoneapi.ContainerRuntimeContainerd{}
	cls.ContainerRuntime.Docker = nil
}

func withDocker(cls *kubeoneapi.KubeOneCluster) {
	cls.ContainerRuntime.Containerd = nil
	cls.ContainerRuntime.Docker = &kubeoneapi.ContainerRuntimeDocker{}
}

func withKubeVersion(ver string) genClusterOpts {
	return func(cls *kubeoneapi.KubeOneCluster) {
		cls.Versions.Kubernetes = ver
	}
}

func withNutanixCloudProvider(cls *kubeoneapi.KubeOneCluster) {
	cls.CloudProvider = kubeoneapi.CloudProviderSpec{
		Nutanix: &kubeoneapi.NutanixSpec{},
	}
}

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
			ImageRepository: "k8s.gcr.io",
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
			Kubernetes: "1.17.4",
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

func TestKubeadmDebian(t *testing.T) {
	t.Parallel()

	type args struct {
		cluster kubeoneapi.KubeOneCluster
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "simple",
			args: args{
				cluster: genCluster(withDocker),
			},
		},
		{
			name: "overwrite registry",
			args: args{
				cluster: genCluster(withDocker, withRegistry("127.0.0.1:5000")),
			},
		},
		{
			name: "overwrite registry insecure",
			args: args{
				cluster: genCluster(withDocker, withInsecureRegistry("127.0.0.1:5000")),
			},
		},
		{
			name: "with containerd",
			args: args{
				cluster: genCluster(withContainerd),
			},
		},
		{
			name: "with containerd with insecure registry",
			args: args{
				cluster: genCluster(withContainerd, withInsecureRegistry("127.0.0.1:5000")),
			},
		},
		{
			name: "nutanix cluster",
			args: args{
				cluster: genCluster(withNutanixCloudProvider),
			},
		},
		{
			name: "cilium cluster",
			args: args{
				cluster: genCluster(withCiliumCNI),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := KubeadmDebian(&tt.args.cluster, false)
			if !errors.Is(err, tt.err) {
				t.Errorf("KubeadmDebian() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cls := genCluster(
				withInsecureRegistry(tt.insecureRegistry),
				withContainerd,
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

func TestKubeadmCentOS(t *testing.T) {
	t.Parallel()

	type args struct {
		cluster kubeoneapi.KubeOneCluster
		force   bool
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "simple",
			args: args{
				cluster: genCluster(withDocker),
			},
		},
		{
			name: "proxy",
			args: args{
				cluster: genCluster(withDocker, withProxy("http://my-proxy.tld")),
			},
		},
		{
			name: "force",
			args: args{
				cluster: genCluster(withDocker),
				force:   true,
			},
		},
		{
			name: "v1.16.1",
			args: args{
				cluster: genCluster(withDocker, withKubeVersion("1.16.1")),
			},
		},
		{
			name: "overwrite registry",
			args: args{
				cluster: genCluster(withDocker, withRegistry("127.0.0.1:5000")),
			},
		},
		{
			name: "overwrite registry insecure",
			args: args{
				cluster: genCluster(withDocker, withInsecureRegistry("127.0.0.1:5000")),
			},
		},
		{
			name: "with containerd",
			args: args{
				cluster: genCluster(withContainerd),
			},
		},
		{
			name: "with containerd with insecure registry",
			args: args{
				cluster: genCluster(withContainerd, withInsecureRegistry("127.0.0.1:5000")),
			},
		},
		{
			name: "nutanix cluster",
			args: args{
				cluster: genCluster(withNutanixCloudProvider),
			},
		},
		{
			name: "cilium cluster",
			args: args{
				cluster: genCluster(withCiliumCNI),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := KubeadmCentOS(&tt.args.cluster, tt.args.force)
			if !errors.Is(err, tt.err) {
				t.Errorf("KubeadmCentOS() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestKubeadmAmazonLinux(t *testing.T) {
	t.Parallel()

	type args struct {
		cluster kubeoneapi.KubeOneCluster
		force   bool
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "simple",
			args: args{
				cluster: genCluster(
					withDocker,
					withDefaultAssetConfiguration,
				),
			},
		},
		{
			name: "proxy",
			args: args{
				cluster: genCluster(
					withDocker,
					withProxy("http://my-proxy.tld"),
					withDefaultAssetConfiguration,
				),
			},
		},
		{
			name: "force",
			args: args{
				cluster: genCluster(
					withDocker,
					withDefaultAssetConfiguration,
				),
				force: true,
			},
		},
		{
			name: "v1.16.1",
			args: args{
				cluster: genCluster(
					withDocker,
					withKubeVersion("1.16.1"),
					withDefaultAssetConfiguration,
				),
			},
		},
		{
			name: "overwrite registry",
			args: args{
				cluster: genCluster(
					withDocker,
					withRegistry("127.0.0.1:5000"),
					withDefaultAssetConfiguration,
				),
			},
		},
		{
			name: "overwrite registry insecure",
			args: args{
				cluster: genCluster(
					withDocker,
					withInsecureRegistry("127.0.0.1:5000"),
					withDefaultAssetConfiguration,
				),
			},
		},
		{
			name: "with containerd",
			args: args{
				cluster: genCluster(withContainerd),
			},
		},
		{
			name: "with containerd with insecure registry",
			args: args{
				cluster: genCluster(withContainerd, withInsecureRegistry("127.0.0.1:5000")),
			},
		},
		{
			name: "with cilium",
			args: args{
				cluster: genCluster(withCiliumCNI),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := KubeadmAmazonLinux(&tt.args.cluster, tt.args.force)
			if !errors.Is(err, tt.err) {
				t.Errorf("KubeadmAmazonLinux() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestKubeadmFlatcar(t *testing.T) {
	t.Parallel()

	type args struct {
		cluster kubeoneapi.KubeOneCluster
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "simple",
			args: args{
				cluster: genCluster(withDocker),
			},
		},
		{
			name: "force",
			args: args{
				cluster: genCluster(withDocker),
			},
		},
		{
			name: "overwrite registry",
			args: args{
				cluster: genCluster(
					withDocker,
					withRegistry("127.0.0.1:5000"),
				),
			},
		},
		{
			name: "overwrite registry insecure",
			args: args{
				cluster: genCluster(
					withDocker,
					withInsecureRegistry("127.0.0.1:5000"),
				),
			},
		},
		{
			name: "with containerd",
			args: args{
				cluster: genCluster(withContainerd),
			},
		},
		{
			name: "with containerd with insecure registry",
			args: args{
				cluster: genCluster(
					withContainerd,
					withInsecureRegistry("127.0.0.1:5000"),
				),
			},
		},
		{
			name: "with cilium",
			args: args{
				cluster: genCluster(withCiliumCNI),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := KubeadmFlatcar(&tt.args.cluster)
			if !errors.Is(err, tt.err) {
				t.Errorf("KubeadmFlatcar() error = %v, wantErr %v", err, tt.err)

				return
			}

			testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
		})
	}
}

func TestRemoveBinariesDebian(t *testing.T) {
	t.Parallel()

	got, err := RemoveBinariesDebian()
	if err != nil {
		t.Errorf("RemoveBinariesDebian() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestRemoveBinariesCentOS(t *testing.T) {
	t.Parallel()

	got, err := RemoveBinariesCentOS()
	if err != nil {
		t.Errorf("RemoveBinariesCentOS() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestRemoveBinariesAmazonLinux(t *testing.T) {
	t.Parallel()

	got, err := RemoveBinariesAmazonLinux()
	if err != nil {
		t.Errorf("RemoveBinariesAmazonLinux() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestRemoveBinariesFlatcar(t *testing.T) {
	t.Parallel()

	got, err := RemoveBinariesFlatcar()
	if err != nil {
		t.Errorf("RemoveBinariesFlatcar() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeadmAndCNIDebian(t *testing.T) {
	t.Parallel()

	cls := genCluster(withDocker)
	got, err := UpgradeKubeadmAndCNIDebian(&cls)
	if err != nil {
		t.Errorf("UpgradeKubeadmAndCNIDebian() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeadmAndCNICentOS(t *testing.T) {
	t.Parallel()

	cls := genCluster(withDocker)
	got, err := UpgradeKubeadmAndCNICentOS(&cls)
	if err != nil {
		t.Errorf("UpgradeKubeadmAndCNICentOS() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeadmAndCNIAmazonLinux(t *testing.T) {
	t.Parallel()

	cls := genCluster(withDocker, withDefaultAssetConfiguration)
	got, err := UpgradeKubeadmAndCNIAmazonLinux(&cls)
	if err != nil {
		t.Errorf("UpgradeKubeadmAndCNIAmazonLinux() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeadmAndCNIFlatcar(t *testing.T) {
	t.Parallel()

	c := genCluster(
		withKubeVersion("v1.17.4"),
		withContainerd,
		withInsecureRegistry("127.0.0.1:5000"),
	)
	got, err := UpgradeKubeadmAndCNIFlatcar(&c)
	if err != nil {
		t.Errorf("UpgradeKubeadmAndCNIFlatcar() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeletAndKubectlDebian(t *testing.T) {
	t.Parallel()

	cls := genCluster(withDocker)
	got, err := UpgradeKubeletAndKubectlDebian(&cls)
	if err != nil {
		t.Errorf("UpgradeKubeletAndKubectlDebian() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeletAndKubectlCentOS(t *testing.T) {
	t.Parallel()

	cls := genCluster(withDocker)
	got, err := UpgradeKubeletAndKubectlCentOS(&cls)
	if err != nil {
		t.Errorf("UpgradeKubeletAndKubectlCentOS() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeletAndKubectlAmazonLinux(t *testing.T) {
	t.Parallel()

	cls := genCluster(withDocker, withDefaultAssetConfiguration)
	got, err := UpgradeKubeletAndKubectlAmazonLinux(&cls)
	if err != nil {
		t.Errorf("UpgradeKubeletAndKubectlAmazonLinux() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeletAndKubectlFlatcar(t *testing.T) {
	t.Parallel()

	c := genCluster(
		withKubeVersion("v1.17.4"),
		withContainerd,
		withInsecureRegistry("127.0.0.1:5000"),
	)
	got, err := UpgradeKubeletAndKubectlFlatcar(&c)
	if err != nil {
		t.Errorf("UpgradeKubeletAndKubectlFlatcar() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}
