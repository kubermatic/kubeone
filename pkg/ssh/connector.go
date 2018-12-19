package ssh

import (
	"sync"
	"time"

	"github.com/kubermatic/kubeone/pkg/config"
)

// Connector holds a map of Connections
type Connector struct {
	lock        sync.RWMutex
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
	var err error

	c.lock.RLock()
	conn, exists := c.connections[node.PublicAddress]
	c.lock.RUnlock()
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

		c.lock.Lock()
		defer c.lock.Unlock()
		c.connections[node.PublicAddress] = conn
	}

	return conn, nil
}

// CloseAll closes all connections
func (c *Connector) CloseAll() {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, conn := range c.connections {
		conn.Close()
	}
}
