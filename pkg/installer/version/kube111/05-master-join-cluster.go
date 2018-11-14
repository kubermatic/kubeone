package kube111

import (
	"strconv"
	"time"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func joinMasterCluster(ctx *util.Context) error {
	ctx.Logger.Infoln("Deploying PKI…")

	return util.RunTaskOnNodes(ctx, joinNodesMasterCluster)
}

func joinNodesMasterCluster(ctx *util.Context, node config.HostConfig, nodeIndex int, conn ssh.Connection) error {
	if nodeIndex == 0 {
		return nil
	}

	ctx.Logger.Infoln("Finalizing cluster…")

	_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, `
set -xeu pipefail

export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"

for tries in $(seq 1 60); do
    # Waiting for kubelet to spawn etcd before joining it a cluster.
    sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf get -n kube-system pod etcd-{{ .MASTER_HOSTNAME }} && break
    sleep 1
done

sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf exec -n kube-system etcd-{{ .MASTER_HOSTNAME }} -- etcdctl --ca-file /etc/kubernetes/pki/etcd/ca.crt --cert-file /etc/kubernetes/pki/etcd/peer.crt --key-file /etc/kubernetes/pki/etcd/peer.key --endpoints=https://{{ .MASTER_ADDRESS }}:2379 member add {{ .NODE_HOSTNAME }} https://{{ .NODE_ADDRESS }}:2380

sudo kubeadm alpha phase etcd local --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_INDEX }}.yaml
sudo kubeadm alpha phase kubeconfig all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_INDEX }}.yaml
sudo kubeadm alpha phase controlplane all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_INDEX }}.yaml
sudo kubeadm alpha phase mark-master --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_INDEX }}.yaml
`, util.TemplateVariables{
		"WORK_DIR":        ctx.WorkDir,
		"NODE_INDEX":      strconv.Itoa(nodeIndex),
		"MASTER_HOSTNAME": ctx.Cluster.Hosts[0].Hostname,
		"NODE_HOSTNAME":   ctx.Cluster.Hosts[nodeIndex].Hostname,
		"MASTER_ADDRESS":  ctx.Cluster.Hosts[0].PrivateAddress,
		"NODE_ADDRESS":    ctx.Cluster.Hosts[nodeIndex].PrivateAddress,
	})

	if err := wait(ctx, 30*time.Second); err != nil {
		return err
	}

	return err
}
