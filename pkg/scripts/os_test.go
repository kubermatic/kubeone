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

func genCluster(opts ...genClusterOpts) kubeone.KubeOneCluster {
	c := &kubeone.KubeOneCluster{
		Versions: kubeone.VersionConfig{
			Kubernetes: "v1.17.4",
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
		fn(c)
	}

	return *c
}

func TestKubeadmDebian(t *testing.T) {
	t.Parallel()

	type args struct {
		cluster       kubeone.KubeOneCluster
		dockerVersion string
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "simple",
			args: args{
				dockerVersion: "18.0.6",
				cluster:       genCluster(),
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
				cluster: genCluster(),
			},
		},
		{
			name: "proxy",
			args: args{
				cluster: genCluster(withProxy("http://my-proxy.tld")),
			},
		},
		{
			name: "force",
			args: args{
				cluster: genCluster(),
				force:   true,
			},
		},
		{
			name: "v1.16.1",
			args: args{
				cluster: genCluster(withKubeVersion("v1.16.1")),
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

func TestKubeadmCoreOS(t *testing.T) {
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
				cluster: genCluster(),
			},
		},
		{
			name: "force",
			args: args{
				cluster: genCluster(),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := KubeadmCoreOS(&tt.args.cluster)
			if err != tt.err {
				t.Errorf("KubeadmCoreOS() error = %v, wantErr %v", err, tt.err)
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

func TestRemoveBinariesCoreOS(t *testing.T) {
	t.Parallel()

	got, err := RemoveBinariesCoreOS()
	if err != nil {
		t.Errorf("RemoveBinariesCoreOS() error = %v", err)
		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeadmAndCNIDebian(t *testing.T) {
	t.Parallel()

	cls := genCluster()
	got, err := UpgradeKubeadmAndCNIDebian(&cls)
	if err != nil {
		t.Errorf("UpgradeKubeadmAndCNIDebian() error = %v", err)
		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeadmAndCNICentOS(t *testing.T) {
	t.Parallel()

	cls := genCluster()
	got, err := UpgradeKubeadmAndCNICentOS(&cls)
	if err != nil {
		t.Errorf("UpgradeKubeadmAndCNICentOS() error = %v", err)
		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeadmAndCNICoreOS(t *testing.T) {
	t.Parallel()

	got, err := UpgradeKubeadmAndCNICoreOS("v1.17.4")
	if err != nil {
		t.Errorf("UpgradeKubeadmAndCNICoreOS() error = %v", err)
		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeletAndKubectlDebian(t *testing.T) {
	t.Parallel()

	cls := genCluster()
	got, err := UpgradeKubeletAndKubectlDebian(&cls)
	if err != nil {
		t.Errorf("UpgradeKubeletAndKubectlDebian() error = %v", err)
		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeletAndKubectlCentOS(t *testing.T) {
	t.Parallel()

	cls := genCluster()
	got, err := UpgradeKubeletAndKubectlCentOS(&cls)
	if err != nil {
		t.Errorf("UpgradeKubeletAndKubectlCentOS() error = %v", err)
		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeletAndKubectlCoreOS(t *testing.T) {
	t.Parallel()

	got, err := UpgradeKubeletAndKubectlCoreOS("v1.17.4")
	if err != nil {
		t.Errorf("UpgradeKubeletAndKubectlCoreOS() error = %v", err)
		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}
