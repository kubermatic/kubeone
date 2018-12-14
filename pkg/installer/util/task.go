package util

import (
	"fmt"
	"sync"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// NodeTask is a task that is specifically tailored to run on a single node.
type NodeTask func(ctx *Context, node *config.HostConfig, conn ssh.Connection) error

// RunTaskOnNodes runs the given task on the given selection of hosts.
func RunTaskOnNodes(ctx *Context, nodes []*config.HostConfig, task NodeTask, parallel bool) error {
	var (
		err  error
		conn ssh.Connection
	)

	wg := sync.WaitGroup{}
	hasErrors := false

	for _, node := range nodes {
		context := ctx.Clone()
		context.Logger = context.Logger.WithField("node", node.PublicAddress)

		// connect to the host (and do not close connection
		// because we want to re-use it for future tasks)
		conn, err = context.Connector.Connect(*node)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
		}

		prefix := ""
		if parallel {
			prefix = fmt.Sprintf("[%s] ", node.PublicAddress)
		}

		context.Runner = &Runner{
			Conn:    conn,
			Verbose: ctx.Verbose,
			OS:      node.OperatingSystem,
			Prefix:  prefix,
		}

		if parallel {
			wg.Add(1)
			go func() {
				err = task(context, node, conn)
				if err != nil {
					hasErrors = true
				}
				wg.Done()
			}()
		} else {
			err = task(context, node, conn)
			if err != nil {
				break
			}
		}
	}

	wg.Wait()

	if hasErrors {
		err = fmt.Errorf("at least one of the tasks has encountered an error")
	}

	return err
}

// RunTaskOnAllNodes runs the given task on all hosts.
func RunTaskOnAllNodes(ctx *Context, task NodeTask, parallel bool) error {
	return RunTaskOnNodes(ctx, ctx.Cluster.Hosts, task, parallel)
}

// RunTaskOnLeader runs the given task on the leader host.
func RunTaskOnLeader(ctx *Context, task NodeTask) error {
	leader, err := ctx.Cluster.Leader()
	if err != nil {
		return err
	}
	hosts := []*config.HostConfig{
		leader,
	}

	return RunTaskOnNodes(ctx, hosts, task, false)
}

// RunTaskOnFollowers runs the given task on the follower hosts.
func RunTaskOnFollowers(ctx *Context, task NodeTask, parallel bool) error {
	return RunTaskOnNodes(ctx, ctx.Cluster.Followers(), task, parallel)
}
