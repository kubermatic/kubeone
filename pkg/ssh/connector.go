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

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
)

// Connector holds a map of Connections
type Connector struct {
	lock        sync.Mutex
	connections map[string]Connection
}

// NewConnector constructor
func NewConnector() *Connector {
	return &Connector{
		connections: make(map[string]Connection),
	}
}

// Connect to the node
func (c *Connector) Connect(node kubeoneapi.HostConfig) (Connection, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	var err error

	conn, found := c.connections[node.PublicAddress]
	if !found {
		opts := Opts{
			Username:    node.SSHUsername,
			Port:        node.SSHPort,
			Hostname:    node.PublicAddress,
			KeyFile:     node.SSHPrivateKeyFile,
			AgentSocket: node.SSHAgentSocket,
			Timeout:     10 * time.Second,
			Bastion:     node.Bastion,
			BastionPort: node.BastionPort,
			BastionUser: node.BastionUser,
		}

		conn, err = NewConnection(opts)
		if err != nil {
			return nil, err
		}

		c.connections[node.PublicAddress] = conn
	}

	return conn, nil
}

// CloseAll closes all connections
func (c *Connector) CloseAll() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, conn := range c.connections {
		conn.Close()
	}
	c.connections = make(map[string]Connection)
}
