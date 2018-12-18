package ssh

import (
	"time"

	"github.com/kubermatic/kubeone/pkg/config"
)

// Connector holds a map of Connections
type Connector struct {
	connections map[string]Connection
}

// NewConnector constructor
func NewConnector() *Connector {
	return &Connector{
		connections: make(map[string]Connection),
	}
}

const (
	defaultConTimeout = 10 * time.Second
)

// Connect to the node
func (c *Connector) Connect(node config.HostConfig) (Connection, error) {
	var err error

	conn, exists := c.connections[node.PublicAddress]
	if !exists || conn.Closed() {
		opts := Opts{
			Username:    node.SSHUsername,
			Port:        node.SSHPort,
			Hostname:    node.PublicAddress,
			KeyFile:     node.SSHPrivateKeyFile,
			AgentSocket: node.SSHAgentSocket,
			Timeout:     defaultConTimeout,
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
	for _, conn := range c.connections {
		conn.Close()
	}
}
