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

	ctx := c.Clone()
	ctx.Logger = ctx.Logger.WithField("node", node.PublicAddress)

	// connect to the host (and do not close connection
	// because we want to re-use it for future tasks)
	conn, err = ctx.Connector.Connect(*node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", node.PublicAddress, err)
	}

	prefix := ""
	if prefixed {
		prefix = fmt.Sprintf("[%s] ", node.PublicAddress)
	}

	ctx.Runner = &Runner{
		Conn:    conn,
		Verbose: ctx.Verbose,
		OS:      node.OperatingSystem,
		Prefix:  prefix,
	}

	return task(ctx, node, conn)
}

// RunTaskOnNodes runs the given task on the given selection of hosts.
func (c *Context) RunTaskOnNodes(nodes []*config.HostConfig, task NodeTask, parallel bool) error {
	var err error

	wg := sync.WaitGroup{}
	hasErrors := false

	for _, node := range nodes {
		if parallel {
			wg.Add(1)
			go func(node *config.HostConfig) {
				err = c.runTask(node, task, parallel)
				if err != nil {
					hasErrors = true
				}
				wg.Done()
			}(node)
		} else {
			err = c.runTask(node, task, parallel)
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
