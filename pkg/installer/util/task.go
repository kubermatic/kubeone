package util

import (
	"fmt"
	"sync"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// NodeTask is a task that is specifically tailored to run on a single node.
type NodeTask func(ctx *Context, node *config.HostConfig, conn ssh.Connection) error

func (c *Context) runTask(node *config.HostConfig, task NodeTask, prefixed bool) error {
	var (
		err  error
		conn ssh.Connection
	)

	// connect to the host (and do not close connection
	// because we want to re-use it for future tasks)
	conn, err = c.Connector.Connect(*node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	prefix := ""
	if prefixed {
		prefix = fmt.Sprintf("[%s] ", node.PublicAddress)
	}

	c.Runner = &Runner{
		Conn:    conn,
		Verbose: c.Verbose,
		OS:      node.OperatingSystem,
		Prefix:  prefix,
	}

	return task(c, node, conn)
}

// RunTaskOnNodes runs the given task on the given selection of hosts.
func (c *Context) RunTaskOnNodes(nodes []*config.HostConfig, task NodeTask, parallel bool) error {
	var err error

	wg := sync.WaitGroup{}
	hasErrors := false

	for _, node := range nodes {
		ctx := c.Clone()
		ctx.Logger = ctx.Logger.WithField("node", node.PublicAddress)

		if parallel {
			wg.Add(1)
			go func(ctx *Context, node *config.HostConfig) {
				err = ctx.runTask(node, task, parallel)
				if err != nil {
					ctx.Logger.Error(err)
					hasErrors = true
				}
				wg.Done()
			}(ctx, node)
		} else {
			err = ctx.runTask(node, task, parallel)
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
func (c *Context) RunTaskOnAllNodes(task NodeTask, parallel bool) error {
	return c.RunTaskOnNodes(c.Cluster.Hosts, task, parallel)
}

// RunTaskOnLeader runs the given task on the leader host.
func (c *Context) RunTaskOnLeader(task NodeTask) error {
	leader, err := c.Cluster.Leader()
	if err != nil {
		return err
	}

	hosts := []*config.HostConfig{
		leader,
	}

	return c.RunTaskOnNodes(hosts, task, false)
}

// RunTaskOnFollowers runs the given task on the follower hosts.
func (c *Context) RunTaskOnFollowers(task NodeTask, parallel bool) error {
	return c.RunTaskOnNodes(c.Cluster.Followers(), task, parallel)
}
