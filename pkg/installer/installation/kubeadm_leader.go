package installation

import (
	"strconv"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/util"
)

const (
	kubeadmCertCommand = `
if [[ -d ./{{ .WORK_DIR }}/pki ]]; then
       sudo rsync -av ./{{ .WORK_DIR }}/pki/ /etc/kubernetes/pki/
       rm -rf ./{{ .WORK_DIR }}/pki
fi
sudo kubeadm init phase certs all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`
	kubeadmInitCommand = `
if [[ -f /etc/kubernetes/admin.conf ]]; then exit 0; fi
sudo kubeadm init --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`
)

func kubeadmCertsOnLeader(ctx *util.Context) error {
	ctx.Logger.Infoln("Configuring certs and etcd on first controller…")
	return ctx.RunTaskOnLeader(kubeadmCertsExecutor)
}

func kubeadmCertsOnFollower(ctx *util.Context) error {
	ctx.Logger.Infoln("Configuring certs and etcd on consecutive controller…")
	return ctx.RunTaskOnFollowers(kubeadmCertsExecutor, true)
}

func kubeadmCertsExecutor(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {

	ctx.Logger.Infoln("Ensuring Certificates…")
	_, _, err := ctx.Runner.Run(kubeadmCertCommand, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
		"NODE_ID":  strconv.Itoa(node.ID),
	})
	return err
}

func initKubernetesLeader(ctx *util.Context) error {
	ctx.Logger.Infoln("Initializing Kubernetes on leader…")

	return ctx.RunTaskOnLeader(func(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Running kubeadm…")

		_, _, err := ctx.Runner.Run(kubeadmInitCommand, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
			"NODE_ID":  strconv.Itoa(node.ID),
		})

		return err
	})
}
