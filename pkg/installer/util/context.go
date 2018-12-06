package util

import (
	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

type Context struct {
	Cluster        *config.Cluster
	Logger         logrus.FieldLogger
	Connector      *ssh.Connector
	Configuration  *Configuration
	WorkDir        string
	JoinCommand    string
	Verbose        bool
	BackupFile     string
	DestroyWorkers bool
}

// Clone returns a shallow copy of the context.
func (c *Context) Clone() *Context {
	return &Context{
		Cluster:        c.Cluster,
		Logger:         c.Logger,
		Connector:      c.Connector,
		Configuration:  c.Configuration,
		WorkDir:        c.WorkDir,
		JoinCommand:    c.JoinCommand,
		Verbose:        c.Verbose,
		BackupFile:     c.BackupFile,
		DestroyWorkers: c.DestroyWorkers,
	}
}
