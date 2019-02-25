package upgrade

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
)

const (
	upgradeKubeletDebianCommand = `
source /etc/os-release
source /etc/kubeone/proxy-env

sudo apt-get update

kube_ver=$(apt-cache madison kubelet | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold kubelet
sudo apt-get install kubelet=${kube_ver}
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
