package kube110

import (
	"fmt"
	"strconv"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func deployCA(ctx *util.Context) error {
	ctx.Logger.Infoln("Deploying PKI…")

	return util.RunTaskOnAllNodes(ctx, deployCAOnNode)
}

func deployCAOnNode(ctx *util.Context, node config.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Uploading files…")
	err := ctx.Configuration.UploadTo(conn, ctx.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to upload: %v", err)
	}

	ctx.Logger.Infoln("Setting up certificates and restarting kubelet…")

	_, _, _, err = util.RunShellCommand(conn, ctx.Verbose, `
sudo rsync -av ./{{ .WORK_DIR }}/pki/ /etc/kubernetes/pki/
rm -rf ./{{ .WORK_DIR }}/pki
sudo chown -R root:root /etc/kubernetes/pki
sudo mkdir -p /etc/kubernetes/manifests
sudo cp ./{{ .WORK_DIR }}/etcd/etcd_{{ .NODE_ID }}.yaml /etc/kubernetes/manifests/etcd.yaml
sudo kubeadm alpha phase certs etcd-healthcheck-client --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo kubeadm alpha phase certs etcd-peer --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo kubeadm alpha phase certs etcd-server --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo kubeadm alpha phase kubeconfig kubelet --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo systemctl daemon-reload
sudo systemctl restart kubelet
`, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
		"NODE_ID":  strconv.Itoa(node.ID),
	})

	return err
}
