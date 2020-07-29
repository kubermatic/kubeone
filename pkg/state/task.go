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

package state

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/runner"
	"k8c.io/kubeone/pkg/ssh"
)

// NodeTask is a task that is specifically tailored to run on a single node.
type NodeTask func(ctx *State, node *kubeoneapi.HostConfig, conn ssh.Connection) error

func (s *State) runTask(node *kubeoneapi.HostConfig, task NodeTask) error {
	var (
		err  error
		conn ssh.Connection
	)

	// connect to the host (and do not close connection
	// because we want to re-use it for future tasks)
	conn, err = s.Connector.Connect(*node)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to %s", node.PublicAddress)
	}

	s.Runner = &runner.Runner{
		Conn:    conn,
		Verbose: s.Verbose,
		OS:      node.OperatingSystem,
		Prefix:  fmt.Sprintf("[%s] ", node.PublicAddress),
	}

	return task(s, node, conn)
}

// RunTaskOnNodes runs the given task on the given selection of hosts.
func (s *State) RunTaskOnNodes(nodes []kubeoneapi.HostConfig, task NodeTask, parallel RunModeEnum) error {
	var err error

	wg := sync.WaitGroup{}
	hasErrors := false

	for i := range nodes {
		ctx := s.Clone()
		ctx.Logger = ctx.Logger.WithField("node", nodes[i].PublicAddress)

		if parallel == RunParallel {
			wg.Add(1)
			go func(ctx *State, node *kubeoneapi.HostConfig) {
				err = ctx.runTask(node, task)
				if err != nil {
					ctx.Logger.Error(err)
					hasErrors = true
				}
				wg.Done()
			}(ctx, &nodes[i])
		} else {
			err = ctx.runTask(&nodes[i], task)
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

type RunModeEnum bool

const (
	RunSequentially RunModeEnum = false
	RunParallel     RunModeEnum = true
)

// RunTaskOnAllNodes runs the given task on all hosts.
func (s *State) RunTaskOnAllNodes(task NodeTask, parallel RunModeEnum) error {
	// It's not possible to concatenate host lists in this function.
	// Some of the tasks(determineOS, determineHostname) write to the state and sending a copy would break that.
	if err := s.RunTaskOnControlPlane(task, parallel); err != nil {
		return err
	}

	return s.RunTaskOnStaticWorkers(task, parallel)
}

// RunTaskOnLeader runs the given task on the leader host.
func (s *State) RunTaskOnLeader(task NodeTask) error {
	leader, err := s.Cluster.Leader()
	if err != nil {
		return err
	}

	hosts := []kubeoneapi.HostConfig{
		leader,
	}

	return s.RunTaskOnNodes(hosts, task, false)
}

// RunTaskOnFollowers runs the given task on the follower hosts.
func (s *State) RunTaskOnFollowers(task NodeTask, parallel RunModeEnum) error {
	return s.RunTaskOnNodes(s.Cluster.Followers(), task, parallel)
}

func (s *State) RunTaskOnControlPlane(task NodeTask, parallel RunModeEnum) error {
	return s.RunTaskOnNodes(s.Cluster.ControlPlane.Hosts, task, parallel)
}

func (s *State) RunTaskOnStaticWorkers(task NodeTask, parallel RunModeEnum) error {
	return s.RunTaskOnNodes(s.Cluster.StaticWorkers.Hosts, task, parallel)
}
