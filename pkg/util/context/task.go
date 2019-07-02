/*
Copyright 2019 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package context

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/runner"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// NodeTask is a task that is specifically tailored to run on a single node.
type NodeTask func(ctx *Context, node *kubeoneapi.HostConfig, conn ssh.Connection) error

func (c *Context) runTask(node *kubeoneapi.HostConfig, task NodeTask, prefixed bool) error {
	var (
		err  error
		conn ssh.Connection
	)

	// connect to the host (and do not close connection
	// because we want to re-use it for future tasks)
	conn, err = c.Connector.Connect(*node)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to %s", node.PublicAddress)
	}

	prefix := ""
	if prefixed {
		prefix = fmt.Sprintf("[%s] ", node.PublicAddress)
	}

	c.Runner = &runner.Runner{
		Conn:    conn,
		Verbose: c.Verbose,
		OS:      node.OperatingSystem,
		Prefix:  prefix,
	}

	return task(c, node, conn)
}

// RunTaskOnNodes runs the given task on the given selection of hosts.
func (c *Context) RunTaskOnNodes(nodes []kubeoneapi.HostConfig, task NodeTask, parallel bool) error {
	var err error

	wg := sync.WaitGroup{}
	hasErrors := false

	for i := range nodes {
		ctx := c.Clone()
		ctx.Logger = ctx.Logger.WithField("node", nodes[i].PublicAddress)

		if parallel {
			wg.Add(1)
			go func(ctx *Context, node *kubeoneapi.HostConfig) {
				err = ctx.runTask(node, task, parallel)
				if err != nil {
					ctx.Logger.Error(err)
					hasErrors = true
				}
				wg.Done()
			}(ctx, &nodes[i])
		} else {
			err = ctx.runTask(&nodes[i], task, parallel)
			if err != nil {
				break
			}
		}
	}

	wg.Wait()

	if hasErrors {
		err = errors.New("at least one of the tasks has encountered an error")
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

	hosts := []kubeoneapi.HostConfig{
		leader,
	}

	return c.RunTaskOnNodes(hosts, task, false)
}

// RunTaskOnFollowers runs the given task on the follower hosts.
func (c *Context) RunTaskOnFollowers(task NodeTask, parallel bool) error {
	return c.RunTaskOnNodes(c.Cluster.Followers(), task, parallel)
}
