package upgrade

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
)

const (
	upgradeKubeadmDebianCommand = `
source /etc/os-release
source /etc/kubeone/proxy-env

sudo apt-get update

kube_ver=$(apt-cache madison kubelet | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold kubeadm
sudo apt-get install kubeadm=${kube_ver}
sudo apt-mark hold kubeadm
`
	upgradeKubeadmCentOSCommand = `
source /etc/kubeone/proxy-env

sudo yum install -y --disableexcludes=kubernetes \
			kubeadm-{{ .KUBERNETES_VERSION }}-0
`
	upgradeKubeadmCoreOSCommand = `
source /etc/kubeone/proxy-env

RELEASE="v{{ .KUBERNETES_VERSION }}"

sudo mkdir -p /opt/bin
cd /opt/bin
sudo curl -L --remote-name-all \
     https://storage.googleapis.com/kubernetes-release/release/${RELEASE}/bin/linux/amd64/kubeadm
sudo chmod +x kubeadm
`
)

func upgradeKubeadm(ctx *util.Context, node *config.HostConfig) error {
	var err error

	switch node.OperatingSystem {
	case "ubuntu", "debian":
		err = upgradeKubeadmDebian(ctx)

	case "coreos":
		err = upgradeKubeadmCoreOS(ctx)

	case "centos":
		err = upgradeKubeadmCentOS(ctx)

	default:
		err = errors.Errorf("'%s' is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func upgradeKubeadmDebian(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(upgradeKubeadmDebianCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
	})

	return errors.WithStack(err)
}

func upgradeKubeadmCentOS(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(upgradeKubeadmCentOSCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
	})

	return errors.WithStack(err)
}

func upgradeKubeadmCoreOS(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(upgradeKubeadmCoreOSCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
	})

	return errors.WithStack(err)
}
