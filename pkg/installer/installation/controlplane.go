package installation

import (
	"strconv"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/util"
)

func joinControlplaneNode(ctx *util.Context) error {
	ctx.Logger.Infoln("Joining controlplane node…")
	return ctx.RunTaskOnFollowers(joinControlPlaneNodeInternal, false)
}

func joinControlPlaneNodeInternal(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	_, _, err := ctx.Runner.Run(`
if [[ -f /etc/kubernetes/kubelet.conf ]]; then exit 0; fi

sudo kubeadm join \
	--config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml \
	--experimental-control-plane
`, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
		"NODE_ID":  strconv.Itoa(node.ID),
	})
	return err
}
