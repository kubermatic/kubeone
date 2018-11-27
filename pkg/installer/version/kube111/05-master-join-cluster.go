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

	return util.RunTaskOnFollowers(ctx, joinNodesMasterCluster)
}

func joinNodesMasterCluster(ctx *util.Context, node config.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Finalizing cluster…")

	leader := ctx.Cluster.Leader()

	_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, `
for tries in $(seq 1 60); do
    # Waiting for kubelet to spawn etcd before joining it a cluster.
    sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf get -n kube-system pod etcd-{{ .LEADER_HOSTNAME }} && break
    sleep 1
done

sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf exec -n kube-system etcd-{{ .LEADER_HOSTNAME }} -- etcdctl --ca-file /etc/kubernetes/pki/etcd/ca.crt --cert-file /etc/kubernetes/pki/etcd/peer.crt --key-file /etc/kubernetes/pki/etcd/peer.key --endpoints=https://{{ .LEADER_ADDRESS }}:2379 member add {{ .NODE_HOSTNAME }} https://{{ .NODE_ADDRESS }}:2380

sudo kubeadm alpha phase etcd local --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo kubeadm alpha phase kubeconfig all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo kubeadm alpha phase controlplane all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo kubeadm alpha phase mark-master --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`, util.TemplateVariables{
		"WORK_DIR":        ctx.WorkDir,
		"NODE_ID":         strconv.Itoa(node.ID),
		"LEADER_HOSTNAME": leader.Hostname,
		"NODE_HOSTNAME":   node.Hostname,
		"LEADER_ADDRESS":  leader.PrivateAddress,
		"NODE_ADDRESS":    node.PrivateAddress,
	})
	if err != nil {
		return err
	}

	return wait(ctx, 30*time.Second)
}
