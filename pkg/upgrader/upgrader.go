package upgrader

import (
	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/upgrader/upgrade"
)

// Options groups the various possible options for running KubeOne upgrade
type Options struct {
	Verbose      bool
	ForceUpgrade bool
}

// Upgrader is entrypoint for the upgrade process
type Upgrader struct {
	cluster *config.Cluster
	logger  *logrus.Logger
}

// NewUpgrader returns a new upgrader, responsible for running the upgrade process
func NewUpgrader(cluster *config.Cluster, logger *logrus.Logger) *Upgrader {
	return &Upgrader{
		cluster: cluster,
		logger:  logger,
	}
}

// Upgrade run the upgrade process
func (u *Upgrader) Upgrade(options *Options) error {
	return upgrade.Upgrade(u.createContext(options))
}

// createContext creates a basic, non-host bound context with all relevant information, but no Runner yet.
// The various task helper functions will take care of setting up Runner structs for each task individually
func (u *Upgrader) createContext(options *Options) *util.Context {
	return &util.Context{
		Cluster:       u.cluster,
		Connector:     ssh.NewConnector(),
		Configuration: util.NewConfiguration(),
		WorkDir:       "kubeone",
		Logger:        u.logger,
		Verbose:       options.Verbose,
		ForceUpgrade:  options.ForceUpgrade,
	}
}
