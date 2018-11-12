package kube110

import (
	"fmt"
	"strconv"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/manifest"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func deployCA(ctx *util.Context) error {
	ctx.Logger.Infoln("Deploying PKI…")

	return util.RunTaskOnNodes(ctx, deployCAOnNode)
}

func deployCAOnNode(ctx *util.Context, node manifest.HostManifest, nodeIndex int, conn ssh.Connection) error {
	ctx.Logger.Infoln("Uploading files…")
	err := ctx.Configuration.UploadTo(conn, ctx.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to upload: %v", err)
	}

	// sudo with local binary directories manually added to path. Needed because some
	// distros don't correctly set up path in non-interactive sessions, e.g. RHEL
	ctx.Logger.Infoln("Setting up certificates and restarting kubelet…")

	_, _, _, err = util.RunShellCommand(conn, ctx.Verbose, `
set -xeu pipefail

export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"

sudo rsync -av ./{{ .WORK_DIR }}/pki/ /etc/kubernetes/pki/
rm -rf ./{{ .WORK_DIR }}/pki
sudo chown -R root:root /etc/kubernetes/pki
sudo mkdir -p /etc/kubernetes/manifests
sudo cp ./{{ .WORK_DIR }}/etcd/etcd_{{ .NODE_INDEX }}.yaml /etc/kubernetes/manifests/etcd.yaml
sudo kubeadm alpha phase certs etcd-healthcheck-client --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo kubeadm alpha phase certs etcd-peer --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo kubeadm alpha phase certs etcd-server --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo kubeadm alpha phase kubeconfig kubelet --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo systemctl daemon-reload
sudo systemctl restart kubelet
`, util.TemplateVariables{
		"WORK_DIR":   ctx.WorkDir,
		"NODE_INDEX": strconv.Itoa(nodeIndex),
	})

	return err
}
