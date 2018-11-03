package tasks

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
}
