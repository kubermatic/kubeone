/*
Copyright 2025 The KubeOne Authors.

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
			name: "force",
			args: args{
				cluster: genCluster(),
			},
		},
		{
			name: "overwrite registry",
			args: args{
				cluster: genCluster(
					withRegistry("127.0.0.1:5000"),
				),
			},
		},
		{
			name: "with containerd",
			args: args{
				cluster: genCluster(),
			},
		},
		{
			name: "with containerd with insecure registry",
			args: args{
				cluster: genCluster(
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

func TestRemoveBinariesFlatcar(t *testing.T) {
	t.Parallel()

	got, err := RemoveBinariesFlatcar()
	if err != nil {
		t.Errorf("RemoveBinariesFlatcar() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeadmAndCNIFlatcar(t *testing.T) {
	t.Parallel()

	c := genCluster(
		withKubeVersion("1.26.0"),
		withInsecureRegistry("127.0.0.1:5000"),
	)
	got, err := UpgradeKubeadmAndCNIFlatcar(&c)
	if err != nil {
		t.Errorf("UpgradeKubeadmAndCNIFlatcar() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}

func TestUpgradeKubeletAndKubectlFlatcar(t *testing.T) {
	t.Parallel()

	c := genCluster(
		withKubeVersion("1.26.0"),
		withInsecureRegistry("127.0.0.1:5000"),
	)
	got, err := UpgradeKubernetesBinariesFlatcar(&c)
	if err != nil {
		t.Errorf("UpgradeKubeletAndKubectlFlatcar() error = %v", err)

		return
	}

	testhelper.DiffOutput(t, testhelper.FSGoldenName(t), got, *updateFlag)
}
