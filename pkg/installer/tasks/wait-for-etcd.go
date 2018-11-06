package tasks

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/manifest"
)

type WaitForEtcdTask struct{}

func (t *WaitForEtcdTask) Execute(ctx *Context) error {
	var err error

	ctx.Logger.Infoln("Waiting for etcd…")

	for _, node := range ctx.Manifest.Hosts {
		logger := ctx.Logger.WithFields(logrus.Fields{
			"node": node.PublicAddress,
		})

		err = t.executeNode(ctx, node, logger)
		if err != nil {
			break
		}
	}

	return err
}

func (t *WaitForEtcdTask) executeNode(ctx *Context, node manifest.HostManifest, logger logrus.FieldLogger) error {
	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	command := fmt.Sprintf(`
sudo curl -s --max-time 3 --fail \
     --cert /etc/kubernetes/pki/etcd/peer.crt \
     --key /etc/kubernetes/pki/etcd/peer.key \
     --cacert /etc/kubernetes/pki/etcd/ca.crt \
	  https://%s:2379/health`,
		node.PrivateAddress)

	logger.Infoln("Waiting…")
	for remaining := 100; remaining >= 0; remaining-- {
		_, _, _, err := runCommand(conn, command, ctx.Verbose)
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
