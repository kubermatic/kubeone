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

package ssh

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"

	"k8s.io/client-go/util/homedir"
)

// Connector holds a map of Connections
type Connector struct {
	lock        sync.Mutex
	connections map[int]executor.Interface
	ctx         context.Context
}

// NewConnector constructor
func NewConnector(ctx context.Context) *Connector {
	return &Connector{
		connections: make(map[int]executor.Interface),
		ctx:         ctx,
	}
}

// Tunnel returns established SSH tunnel
func (c *Connector) Tunnel(host kubeoneapi.HostConfig) (executor.Tunneler, error) {
	conn, err := c.Open(host)
	if err != nil {
		return nil, err
	}

	tunn, ok := conn.(executor.Tunneler)
	if !ok {
		err = fail.RuntimeError{
			Op:  "tunneler interface",
			Err: errors.New("unable to assert"),
		}
	}

	return tunn, err
}

// Open to the node
func (c *Connector) Open(host kubeoneapi.HostConfig) (executor.Interface, error) {
	var err error

	c.lock.Lock()
	defer c.lock.Unlock()

	conn, found := c.connections[host.ID]
	if !found {
		opts := sshOpts(host)
		opts.Context = c.ctx
		conn, err = NewConnection(c, opts)
		if err != nil {
			return nil, err
		}

		c.connections[host.ID] = conn
	}

	return conn, nil
}

func (c *Connector) forgetConnection(conn *connection) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for k := range c.connections {
		if c.connections[k] == conn {
			delete(c.connections, k)
		}
	}
}

func sshOpts(host kubeoneapi.HostConfig) Opts {
	privateKeyFile := host.SSHPrivateKeyFile
	// Expand ~/ as path to the home directory
	if strings.HasPrefix(privateKeyFile, "~/") {
		privateKeyFile = filepath.Join(homedir.HomeDir(), privateKeyFile[2:])
	}

	return Opts{
		Username:    host.SSHUsername,
		Port:        host.SSHPort,
		Hostname:    host.PublicAddress,
		KeyFile:     privateKeyFile,
		AgentSocket: host.SSHAgentSocket,
		Timeout:     10 * time.Second,
		Bastion:     host.Bastion,
		BastionPort: host.BastionPort,
		BastionUser: host.BastionUser,
	}
}
