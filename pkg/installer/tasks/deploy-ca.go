package tasks

import (
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/manifest"
)

type DeployCATask struct{}

func (t *DeployCATask) Execute(ctx *Context) error {
	var err error

	ctx.Logger.Infoln("Deploying PKI…")

	for idx, node := range ctx.Manifest.Hosts {
		logger := ctx.Logger.WithFields(logrus.Fields{
			"node": node.Address,
		})

		err = t.executeNode(ctx, node, idx, logger)
		if err != nil {
			break
		}
	}

	return err
}

func (t *DeployCATask) executeNode(ctx *Context, node manifest.HostManifest, nodeIndex int, logger logrus.FieldLogger) error {
	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.Address, err)
	}

	logger.Infoln("Uploading files…")
	err = ctx.Configuration.UploadTo(conn, ctx.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to upload: %v", err)
	}

	// sudo with local binary directories manually added to path. Needed because some
	// distros don't correctly set up path in non-interactive sessions, e.g. RHEL
	logger.Infoln("Setting up certificates and restarting kubelet…")
	command, err := makeShellCommand(`
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
sudo systemctl restart kubelet
`, map[string]string{
		"WORK_DIR":   ctx.WorkDir,
		"NODE_INDEX": strconv.Itoa(nodeIndex),
	})
	if err != nil {
		return fmt.Errorf("failed to construct shell script: %v", err)
	}

	_, stderr, _, err := conn.Exec(command)
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr)
	}

	return err
}
