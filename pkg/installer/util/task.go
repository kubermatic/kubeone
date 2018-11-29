package util

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// NodeTask is a task that is specifically tailored to run on a single node.
type NodeTask func(ctx *Context, node config.HostConfig, conn ssh.Connection) error

// RunTaskOnNodes runs the given task on the given selection of hosts.
func RunTaskOnNodes(ctx *Context, nodes []config.HostConfig, task NodeTask) error {
	var (
		err  error
		conn ssh.Connection
	)

	for _, node := range nodes {
		context := ctx.Clone()
		context.Logger = context.Logger.WithField("node", node.PublicAddress)

		// connect to the host (and do not close connection
		// because we want to re-use it for future tasks)
		conn, err = context.Connector.Connect(node)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
		}

		err = task(context, node, conn)
		if err != nil {
			break
		}
	}

	return err
}

// RunTaskOnAllNodes runs the given task on all hosts.
func RunTaskOnAllNodes(ctx *Context, task NodeTask) error {
	return RunTaskOnNodes(ctx, ctx.Cluster.Hosts, task)
}

// RunTaskOnLeader runs the given task on the leader host.
func RunTaskOnLeader(ctx *Context, task NodeTask) error {
	hosts := []config.HostConfig{
		ctx.Cluster.Leader(),
	}

	return RunTaskOnNodes(ctx, hosts, task)
}

// RunTaskOnFollowers runs the given task on the follower hosts.
func RunTaskOnFollowers(ctx *Context, task NodeTask) error {
	return RunTaskOnNodes(ctx, ctx.Cluster.Followers(), task)
}
