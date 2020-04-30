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
	"sync"
	"time"

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
)

// Connector holds a map of Connections
type Connector struct {
	lock        sync.Mutex
	connections map[int]Connection
}

// NewConnector constructor
func NewConnector() *Connector {
	return &Connector{
		connections: make(map[int]Connection),
	}
}

// Tunnel returns established SSH tunnel
func (c *Connector) Tunnel(host kubeoneapi.HostConfig) (Tunneler, error) {
	conn, err := c.Connect(host)
	if err != nil {
		return nil, err
	}

	tunn, ok := conn.(Tunneler)
	if !ok {
		err = errors.New("unable to assert Tunneler")
	}

	return tunn, err
}

// Connect to the node
func (c *Connector) Connect(host kubeoneapi.HostConfig) (Connection, error) {
	var err error

	c.lock.Lock()
	defer c.lock.Unlock()

	conn, found := c.connections[host.ID]
	if !found {
		conn, err = NewConnection(c, sshOpts(host))
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
	return Opts{
		Username:    host.SSHUsername,
		Port:        host.SSHPort,
		Hostname:    host.PublicAddress,
		KeyFile:     host.SSHPrivateKeyFile,
		AgentSocket: host.SSHAgentSocket,
		Timeout:     10 * time.Second,
		Bastion:     host.Bastion,
		BastionPort: host.BastionPort,
		BastionUser: host.BastionUser,
	}
}
