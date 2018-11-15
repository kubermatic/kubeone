package kube111

import (
	"fmt"
	"strconv"

	"github.com/kubermatic/kubeone/pkg/config"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/sirupsen/logrus"
)

func downloadCA(ctx *util.Context) error {
	ctx.Logger.Infoln("Generating PKI…")

	node := ctx.Cluster.Hosts[0]
	logger := ctx.Logger.WithFields(logrus.Fields{
		"node": node.PublicAddress,
	})

	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	logger.Infoln("Running kubeadm…")

	_, _, _, err = util.RunShellCommand(conn, ctx.Verbose, `
set -xeu pipefail

export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"

mkdir -p ./{{ .WORK_DIR }}/pki/etcd
sudo cp /etc/kubernetes/pki/ca.crt ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/ca.key ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/sa.key ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/sa.pub ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/front-proxy-ca.crt ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/front-proxy-ca.key ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/etcd/ca.crt ./{{ .WORK_DIR }}/pki/etcd/ca.crt
sudo cp /etc/kubernetes/pki/etcd/ca.key ./{{ .WORK_DIR }}/pki/etcd/ca.key
sudo cp /etc/kubernetes/admin.conf ./{{ .WORK_DIR }}/pki/

sudo chown -R "$USER:$USER" ./{{ .WORK_DIR }}
`, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
	})
	if err != nil {
		return err
	}

	logger.Infoln("Downloading PKI files…")

	err = ctx.Configuration.Download(conn, ctx.WorkDir+"/pki", "pki")
	if err != nil {
		return fmt.Errorf("failed to download PKI files: %v", err)
	}

	return nil
}

func deployCA(ctx *util.Context) error {
	ctx.Logger.Infoln("Deploying PKI…")

	return util.RunTaskOnNodes(ctx, deployCAOnNode)
}

func deployCAOnNode(ctx *util.Context, node config.HostConfig, nodeIndex int, conn ssh.Connection) error {
	if nodeIndex == 0 {
		return nil
	}

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
sudo mv /etc/kubernetes/pki/admin.conf /etc/kubernetes/admin.conf
rm -rf ./{{ .WORK_DIR }}/pki
sudo chown -R root:root /etc/kubernetes
sudo mkdir -p /etc/kubernetes/manifests
sudo kubeadm alpha phase certs all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_INDEX }}.yaml
sudo kubeadm alpha phase kubelet config write-to-disk --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_INDEX }}.yaml
sudo kubeadm alpha phase kubelet write-env-file --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_INDEX }}.yaml
sudo kubeadm alpha phase kubeconfig kubelet --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_INDEX }}.yaml
sudo systemctl daemon-reload
sudo systemctl restart kubelet
`, util.TemplateVariables{
		"WORK_DIR":   ctx.WorkDir,
		"NODE_INDEX": strconv.Itoa(nodeIndex),
	})

	return err
}
