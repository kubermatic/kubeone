package kube110

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/sirupsen/logrus"
)

func generateCA(ctx *util.Context) error {
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

sudo kubeadm alpha phase certs ca --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo kubeadm alpha phase certs etcd-ca --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo kubeadm alpha phase certs sa --config=./{{ .WORK_DIR }}/cfg/master.yaml
sudo rsync -av /etc/kubernetes/pki/ ./{{ .WORK_DIR }}/pki/
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

	if ctx.BackupFile != "" {
		logger.Infoln("Creating local backup…")

		err = ctx.Configuration.Backup(ctx.BackupFile)
		if err != nil {
			// do not stop in case of failed backups, the user can
			// always create the backup themselves if needed
			logger.Warnf("Failed to create backup: %v", err)
		}
	}

	return nil
}
