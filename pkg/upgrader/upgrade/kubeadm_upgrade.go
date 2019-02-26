package upgrade

import (
	"github.com/kubermatic/kubeone/pkg/installer/util"
)

const (
	kubeadmUpgradeLeaderCommand = `
if [[ -f /etc/kubernetes/kubelet.conf ]]; then exit 0; fi
sudo kubeadm upgrade apply -y {{ .VERSION }}
`
	kubeadmUpgradeFollowerCommand = `
if [[ -f /etc/kubernetes/kubelet.conf ]]; then exit 0; fi
sudo kubeadm upgrade node experimental-control-plane
`
)

func upgradeLeaderControlPlane(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(kubeadmUpgradeLeaderCommand, util.TemplateVariables{
		"VERSION": ctx.Cluster.Versions.Kubernetes,
	})
	return err
}

func upgradeFollowerControlPlane(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(kubeadmUpgradeFollowerCommand, util.TemplateVariables{})
	return err
}
