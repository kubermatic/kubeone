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
	"testing"

	"k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/testhelper"
)

type genClusterOpts func(*kubeone.KubeOneCluster)

func withContainerd(cls *kubeone.KubeOneCluster) {
	cls.ContainerRuntime.Containerd = &kubeone.ContainerRuntimeContainerd{}
	cls.ContainerRuntime.Docker = nil
}

func withDocker(cls *kubeone.KubeOneCluster) {
	cls.ContainerRuntime.Containerd = nil
	cls.ContainerRuntime.Docker = &kubeone.ContainerRuntimeDocker{}
}

func withKubeVersion(ver string) genClusterOpts {
	return func(cls *kubeone.KubeOneCluster) {
		cls.Versions.Kubernetes = ver
	}
}

func withProxy(proxy string) genClusterOpts {
	return func(cls *kubeone.KubeOneCluster) {
		cls.Proxy.HTTPS = proxy
		cls.Proxy.HTTP = proxy
	}
}

func withRegistry(registry string) genClusterOpts {
	return func(cls *kubeone.KubeOneCluster) {
		cls.RegistryConfiguration = &kubeone.RegistryConfiguration{
			OverwriteRegistry: registry,
		}
	}
}

func withInsecureRegistry(registry string) genClusterOpts {
	return func(cls *kubeone.KubeOneCluster) {
		cls.RegistryConfiguration = &kubeone.RegistryConfiguration{
			OverwriteRegistry: registry,
			InsecureRegistry:  true,
		}
	}
}

func withDefaultAssetConfiguration(cls *kubeone.KubeOneCluster) {
	cls.AssetConfiguration = kubeone.AssetConfiguration{
		Kubernetes: kubeone.ImageAsset{
			ImageRepository: "k8s.gcr.io",
		},
		CNI: kubeone.BinaryAsset{
			URL: "http://127.0.0.1/cni.tar.gz",
		},
		NodeBinaries: kubeone.BinaryAsset{
			URL: "http://127.0.0.1/node.tar.gz",
		},
		Kubectl: kubeone.BinaryAsset{
			URL: "http://127.0.0.1/kubectl.tar.gz",
		},
	}
}

func genCluster(opts ...genClusterOpts) kubeone.KubeOneCluster {
	cls := &kubeone.KubeOneCluster{
		Versions: kubeone.VersionConfig{
			Kubernetes: "1.17.4",
		},
		SystemPackages: &kubeone.SystemPackages{
			ConfigureRepositories: true,
		},
		Proxy: kubeone.ProxyConfig{
			HTTP:    "http://http.proxy",
			HTTPS:   "http://https.proxy",
			NoProxy: ".local",
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
		cluster kubeone.KubeOneCluster
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := KubeadmDebian(&tt.args.cluster, false)
			if err != tt.err {
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
		name                     string
		insecureRegistry         string
		generateContainerdConfig bool
		err                      error
	}{
		{
			name:                     "simple",
			generateContainerdConfig: true,
		},
		{
			name:                     "flatcat",
			generateContainerdConfig: false,
		},
		{
			name:                     "insecureRegistry",
			insecureRegistry:         "some.registry",
			generateContainerdConfig: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := MigrateToContainerd(tt.insecureRegistry, tt.generateContainerdConfig)
			if err != tt.err {
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
		cluster kubeone.KubeOneCluster
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := KubeadmCentOS(&tt.args.cluster, tt.args.force)
			if err != tt.err {
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
		cluster kubeone.KubeOneCluster
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := KubeadmAmazonLinux(&tt.args.cluster, tt.args.force)
			if err != tt.err {
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
		cluster kubeone.KubeOneCluster
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := KubeadmFlatcar(&tt.args.cluster)
			if err != tt.err {
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

	got, err := UpgradeKubeadmAndCNIFlatcar("v1.17.4")
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

	got, err := UpgradeKubeletAndKubectlFlatcar("v1.17.4")
	if err != nil {
		t.Errorf("UpgradeKubeletAndKubectlFlatcar() error = %v", err)
		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}
