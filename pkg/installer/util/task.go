package util

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/manifest"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// NodeTask is a task that is specifically tailored to run on a single node.
type NodeTask func(ctx *Context, node manifest.HostManifest, nodeIndex int, conn ssh.Connection) error

// RunTaskOnNodes runs the given task on all hosts.
func RunTaskOnNodes(ctx *Context, task NodeTask) error {
	var (
		err  error
		conn ssh.Connection
	)

	for idx, node := range ctx.Manifest.Hosts {
		context := ctx.Clone()
		context.Logger = context.Logger.WithField("node", node.PublicAddress)

		// connect to the host (and do not close connection
		// because we want to re-use it for future tasks)
		conn, err = context.Connector.Connect(node)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
		}

		err = task(context, node, idx, conn)
		if err != nil {
			break
		}
	}

	return err
}
