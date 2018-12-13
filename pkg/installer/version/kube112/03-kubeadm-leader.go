package kube112

import (
	"fmt"
	"strconv"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

const (
	kubeadmCertCommand = `
if [[ -d ./{{ .WORK_DIR }}/pki ]]; then
	sudo rsync -av ./{{ .WORK_DIR }}/pki/ /etc/kubernetes/pki/
	rm -rf ./{{ .WORK_DIR }}/pki
fi
sudo chown -R root:root /etc/kubernetes
sudo kubeadm alpha phase certs all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo kubeadm alpha phase etcd local --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`
	kubeadmInitCommand = `
if [[ ! -f /etc/kubernetes/admin.conf ]]; then
	sudo mv /etc/systemd/system/kubelet.service.d/10-kubeadm.conf.disabled \
			/etc/systemd/system/kubelet.service.d/10-kubeadm.conf
	sudo systemctl daemon-reload
	sudo systemctl stop kubelet
	sudo kubeadm init --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml \
		--ignore-preflight-errors=FileAvailable--etc-kubernetes-manifests-etcd.yaml,Port-10250,Port-2379,DirAvailable--var-lib-etcd
else
	echo "skip init, already initialized"
fi
`
)

func kubeadmCertsOnLeader(ctx *util.Context) error {
	return util.RunTaskOnLeader(ctx, kubeadmCertsExecutor)
}

func kubeadmCertsOnFollower(ctx *util.Context) error {
	return util.RunTaskOnFollowers(ctx, kubeadmCertsExecutor)
}

func kubeadmCertsExecutor(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Ensuring Certificates…")
	_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, kubeadmCertCommand, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
		"NODE_ID":  strconv.Itoa(node.ID),
	})
	return err
}

func initKubernetesLeader(ctx *util.Context) error {
	ctx.Logger.Infoln("Initializing Kubernetes on leader…")

	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Running kubeadm…")

		stdout, stderr, _, err := util.RunShellCommand(conn, ctx.Verbose, kubeadmInitCommand, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
			"NODE_ID":  strconv.Itoa(node.ID),
		})
		if err != nil {
			return fmt.Errorf("error: %v, stdout: %s, stderr: %s", err, stdout, stderr)
		}

		return nil
	})
}
