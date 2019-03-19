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

package upgrade

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/util"
)

const (
	upgradeKubeletDebianCommand = `
source /etc/os-release
source /etc/kubeone/proxy-env

sudo apt-get update

kube_ver=$(apt-cache madison kubelet | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold kubelet
sudo apt-get install -y --no-install-recommends kubelet=${kube_ver}
sudo apt-mark hold kubelet
`
	upgradeKubeletCentOSCommand = `
source /etc/kubeone/proxy-env

sudo yum install -y --disableexcludes=kubernetes \
			kubelet-{{ .KUBERNETES_VERSION }}-0
`
	upgradeKubeletCoreOSCommand = `
source /etc/kubeone/proxy-env

RELEASE="v{{ .KUBERNETES_VERSION }}"

sudo mkdir -p /opt/bin
cd /opt/bin
sudo curl -L --remote-name-all \
     https://storage.googleapis.com/kubernetes-release/release/${RELEASE}/bin/linux/amd64/kubelet
sudo chmod +x kubelet
`
)

func upgradeKubelet(ctx *util.Context, node *config.HostConfig) error {
	var err error

	switch node.OperatingSystem {
	case "ubuntu", "debian":
		err = upgradeKubeletDebian(ctx)

	case "coreos":
		err = upgradeKubeletCoreOS(ctx)

	case "centos":
		err = upgradeKubeletCentOS(ctx)

	default:
		err = errors.Errorf("'%s' is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func upgradeKubeletDebian(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(upgradeKubeletDebianCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
	})

	return errors.WithStack(err)
}

func upgradeKubeletCentOS(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(upgradeKubeletCentOSCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
	})

	return errors.WithStack(err)
}

func upgradeKubeletCoreOS(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(upgradeKubeletCoreOSCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
	})

	return errors.WithStack(err)
}
