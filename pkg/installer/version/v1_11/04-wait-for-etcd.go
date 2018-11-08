package v1_11

import (
	"fmt"
	"time"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"

	"github.com/kubermatic/kubeone/pkg/manifest"
)

func WaitForEtcd(ctx *util.Context) error {
	ctx.Logger.Infoln("Waiting for etcd…")

	return util.RunTaskOnNodes(ctx, waitForEtcdOnNode)
}

func waitForEtcdOnNode(ctx *util.Context, node manifest.HostManifest, _ int, conn ssh.Connection) error {
	var err error

	command := fmt.Sprintf(`
sudo curl -s --max-time 3 --fail \
     --cert /etc/kubernetes/pki/etcd/peer.crt \
     --key /etc/kubernetes/pki/etcd/peer.key \
     --cacert /etc/kubernetes/pki/etcd/ca.crt \
	  https://%s:2379/health`,
		node.PrivateAddress)

	ctx.Logger.Infoln("Waiting…")
	for remaining := 100; remaining >= 0; remaining-- {
		_, _, _, err := util.RunCommand(conn, command, ctx.Verbose)
		if err == nil {
			break
		}
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("etcd did not come up, giving up")
	}

	return nil
}
