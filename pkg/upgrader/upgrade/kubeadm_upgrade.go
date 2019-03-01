package upgrade

import (
	"github.com/kubermatic/kubeone/pkg/util"
)

const (
	kubeadmUpgradeLeaderCommand = `
sudo kubeadm upgrade apply \
	--config=./{{ .WORK_DIR }}/cfg/master_0.yaml \
	-y {{ .VERSION }}
`
	kubeadmUpgradeFollowerCommand = `
sudo kubeadm upgrade node experimental-control-plane
`
)

func upgradeLeaderControlPlane(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(kubeadmUpgradeLeaderCommand, util.TemplateVariables{
		"VERSION":  ctx.Cluster.Versions.Kubernetes,
		"WORK_DIR": ctx.WorkDir,
	})

	return err
}

func upgradeFollowerControlPlane(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(kubeadmUpgradeFollowerCommand, util.TemplateVariables{})
	return err
}
