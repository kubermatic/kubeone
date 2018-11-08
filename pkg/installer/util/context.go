package util

import (
	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/manifest"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

type Context struct {
	Manifest      *manifest.Manifest
	Logger        logrus.FieldLogger
	Connector     *ssh.Connector
	Configuration *Configuration
	WorkDir       string
	JoinCommand   string
	Verbose       bool
}

// Clone returns a shallow copy of the context.
func (c *Context) Clone() *Context {
	return &Context{
		Manifest:      c.Manifest,
		Logger:        c.Logger,
		Connector:     c.Connector,
		Configuration: c.Configuration,
		WorkDir:       c.WorkDir,
		JoinCommand:   c.JoinCommand,
		Verbose:       c.Verbose,
	}
}
