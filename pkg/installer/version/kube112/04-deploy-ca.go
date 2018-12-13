package kube112

import (
	"fmt"
	"strconv"

	"github.com/kubermatic/kubeone/pkg/config"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func downloadCA(ctx *util.Context) error {
	ctx.Logger.Infoln("Generating PKI…")

	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, _ *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Running kubeadm…")

		_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, `
mkdir -p ./{{ .WORK_DIR }}/pki/etcd
sudo cp /etc/kubernetes/pki/ca.crt ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/ca.key ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/sa.key ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/sa.pub ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/front-proxy-ca.crt ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/front-proxy-ca.key ./{{ .WORK_DIR }}/pki/
#sudo cp /etc/kubernetes/pki/etcd/ca.crt ./{{ .WORK_DIR }}/pki/etcd/ca.crt
#sudo cp /etc/kubernetes/pki/etcd/ca.key ./{{ .WORK_DIR }}/pki/etcd/ca.key
#sudo cp /etc/kubernetes/admin.conf ./{{ .WORK_DIR }}/pki/

sudo chown -R "$USER:$USER" ./{{ .WORK_DIR }}
`, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
		})
		if err != nil {
			return err
		}

		ctx.Logger.Infoln("Downloading PKI files…")

		err = ctx.Configuration.Download(conn, ctx.WorkDir+"/pki", "pki")
		if err != nil {
			return fmt.Errorf("failed to download PKI files: %v", err)
		}

		if ctx.BackupFile != "" {
			ctx.Logger.Infoln("Creating local backup…")

			err = ctx.Configuration.Backup(ctx.BackupFile)
			if err != nil {
				// do not stop in case of failed backups, the user can
				// always create the backup themselves if needed
				ctx.Logger.Warnf("Failed to create backup: %v", err)
			}
		}

		return nil
	})
}

func deployCA(ctx *util.Context) error {
	ctx.Logger.Infoln("Deploying PKI…")

	return util.RunTaskOnFollowers(ctx, deployCAOnNode)
}

func deployCAOnNode(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Uploading files…")
	err := ctx.Configuration.UploadTo(conn, ctx.WorkDir)
	return err
	if err != nil {
		return fmt.Errorf("failed to upload: %v", err)
	}

	ctx.Logger.Infoln("Setting up certificates and restarting kubelet…")

	_, _, _, err = util.RunShellCommand(conn, ctx.Verbose, `
sudo rsync -av ./{{ .WORK_DIR }}/pki/ /etc/kubernetes/pki/
rm -rf ./{{ .WORK_DIR }}/pki
sudo chown -R root:root /etc/kubernetes
sudo mkdir -p /etc/kubernetes/manifests
sudo kubeadm alpha phase certs all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo kubeadm alpha phase kubelet config write-to-disk --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo kubeadm alpha phase kubelet write-env-file --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo kubeadm alpha phase kubeconfig kubelet --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo systemctl daemon-reload
sudo systemctl restart kubelet
`, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
		"NODE_ID":  strconv.Itoa(node.ID),
	})

	return err
}
