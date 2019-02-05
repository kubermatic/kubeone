package installer

import (
	"github.com/pkg/errors"

	"github.com/Masterminds/semver"
	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/installation"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// Options groups the various possible options for running
// the Kubernetes installation.
type Options struct {
	Verbose        bool
	BackupFile     string
	DestroyWorkers bool
}

// Installer is entrypoint for installation process
type Installer struct {
	cluster *config.Cluster
	logger  *logrus.Logger
}

// NewInstaller returns a new installer, responsible for dispatching
// between the different supported Kubernetes versions and running the
func NewInstaller(cluster *config.Cluster, logger *logrus.Logger) *Installer {
	return &Installer{
		cluster: cluster,
		logger:  logger,
	}
}

// Install run the installation process
func (i *Installer) Install(options *Options) error {
	ctx := i.createContext(options)

	v, err := semver.NewVersion(i.cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}
	if v.Minor() < 13 {
		return errors.New("kubernetes versions lower than 1.13 are not supported")
	}

	return installation.Install(ctx)
}

// Reset resets cluster:
// * destroys all the worker machines
// * kubeadm reset masters
func (i *Installer) Reset(options *Options) error {
	ctx := i.createContext(options)

	v := semver.MustParse(i.cluster.Versions.Kubernetes)
	if v.Minor() < 13 {
		return errors.New("kubernetes versions lower than 1.13 are not supported")
	}

	return installation.Reset(ctx)
}

// createContext creates a basic, non-host bound context with
// all relevant information, but *no* Runner yet. The various
// task helper functions will take care of setting up Runner
// structs for each task individually.
func (i *Installer) createContext(options *Options) *util.Context {
	return &util.Context{
		Cluster:        i.cluster,
		Connector:      ssh.NewConnector(),
		Configuration:  util.NewConfiguration(),
		WorkDir:        "kubeone",
		Logger:         i.logger,
		Verbose:        options.Verbose,
		BackupFile:     options.BackupFile,
		DestroyWorkers: options.DestroyWorkers,
	}
}
