package tasks

import (
	"github.com/kubermatic/kubeone/pkg/manifest"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/sirupsen/logrus"
)

type Context struct {
	Manifest      *manifest.Manifest
	Logger        logrus.FieldLogger
	Connector     *ssh.Connector
	Configuration *Configuration
	WorkDir       string
	JoinCommand   string
}
