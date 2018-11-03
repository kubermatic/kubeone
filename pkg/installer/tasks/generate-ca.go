package tasks

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type GenerateCATask struct{}

func (t *GenerateCATask) Execute(ctx *Context) error {
	ctx.Logger.Infoln("Generating PKI…")

	node := ctx.Manifest.Hosts[0]
	logger := ctx.Logger.WithFields(logrus.Fields{
		"node":   node.Address,
		"master": true,
	})

	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.Address, err)
	}

	logger.Infoln("Running kubeadm…")
	// sudo with local binary directories manually added to path. Needed because some
	// distros don't correctly set up path in non-interactive sessions, e.g. RHEL
	command, err := makeShellCommand(`
set -xeu pipefail

export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"

sudo kubeadm alpha phase certs ca --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo kubeadm alpha phase certs etcd-ca --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo kubeadm alpha phase certs sa --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo rsync -av /etc/kubernetes/pki/ ./{{ .WORK_DIR }}/pki/
sudo chown -R "$USER:$USER" ./{{ .WORK_DIR }}
`, map[string]string{
		"WORK_DIR": ctx.WorkDir,
	})
	if err != nil {
		return fmt.Errorf("failed to construct shell script: %v", err)
	}

	_, stderr, _, err := conn.Exec(command)
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr)
	}

	logger.Infoln("Downloading PKI files…")

	err = ctx.Configuration.Download(conn, ctx.WorkDir+"/pki", "pki")
	if err != nil {
		return fmt.Errorf("failed to download PKI files: %v", err)
	}

	return nil
}
