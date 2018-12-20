package ssh

import (
	"sync"
	"time"

	"github.com/kubermatic/kubeone/pkg/config"
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
func (c *Connector) Connect(node config.HostConfig) (Connection, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	var err error

	conn, exists := c.connections[node.PublicAddress]
	if !exists || conn.Closed() {
		opts := Opts{
			Username:    node.SSHUsername,
			Port:        node.SSHPort,
			Hostname:    node.PublicAddress,
			KeyFile:     node.SSHPrivateKeyFile,
			AgentSocket: node.SSHAgentSocket,
			Timeout:     10 * time.Second,
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
}
