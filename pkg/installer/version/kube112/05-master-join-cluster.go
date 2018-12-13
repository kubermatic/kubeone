package kube112

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func joinMasterCluster(ctx *util.Context) error {
	ctx.Logger.Infoln("Deploying PKI…")

	if err := util.RunTaskOnFollowers(ctx, deployETCDManifest); err != nil {
		return err
	}

	if err := util.RunTaskOnFollowers(ctx, joinETCDMember); err != nil {
		return err
	}
	return util.RunTaskOnFollowers(ctx, joinNodesMasterCluster)
}

func deployETCDManifest(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, `
sudo kubeadm alpha phase etcd local --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
		"NODE_ID":  strconv.Itoa(node.ID),
	})
	return err
}

func joinETCDMember(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	leader, err := ctx.Cluster.Leader()
	if err != nil {
		return err
	}
	ctx.Logger.Infoln("Waiting for first etcd to come up…")
	if err = util.WaitForPod(conn, ctx.Verbose, "kube-system", fmt.Sprintf("etcd-%s", leader.Hostname), 2*time.Minute); err != nil {
		return err
	}

	ctx.Logger.Infoln("Joining etcd member...")
	_, _, _, err = util.RunShellCommand(conn, ctx.Verbose, `
sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf exec \
  -n kube-system etcd-{{ .LEADER_HOSTNAME }} -- \
  etcdctl \
    --ca-file /etc/kubernetes/pki/etcd/ca.crt \
    --cert-file /etc/kubernetes/pki/etcd/peer.crt \
    --key-file /etc/kubernetes/pki/etcd/peer.key \
    --endpoints=https://{{ .LEADER_ADDRESS }}:2379 \
    member add {{ .NODE_HOSTNAME }} https://{{ .NODE_ADDRESS }}:2380
`, util.TemplateVariables{
		"LEADER_ADDRESS":  leader.PrivateAddress,
		"LEADER_HOSTNAME": leader.Hostname,
		"NODE_HOSTNAME":   node.Hostname,
		"NODE_ADDRESS":    node.PrivateAddress,
	})
	return err

}

func joinNodesMasterCluster(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Joining additional controller..…")
	_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, `
sudo kubeadm alpha phase etcd local --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo kubeadm alpha phase kubeconfig all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo kubeadm alpha phase controlplane all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo kubeadm alpha phase kubelet config annotate-cri --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
sudo kubeadm alpha phase mark-master --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`, util.TemplateVariables{
		"WORK_DIR":      ctx.WorkDir,
		"NODE_ID":       strconv.Itoa(node.ID),
		"NODE_HOSTNAME": node.Hostname,
		"NODE_ADDRESS":  node.PrivateAddress,
	})
	return err
}
