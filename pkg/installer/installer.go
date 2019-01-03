package installer

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/installation"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/installer/version/kube112"
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
	var err error

	ctx := i.createContext(options)

	v, err := semver.NewVersion(i.cluster.Versions.Kubernetes)
	if err != nil {
		return fmt.Errorf("can't parse kubernetes version: %v", err)
	}

	majorMinor := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	switch majorMinor {
	case "1.12":
		err = kube112.Install(ctx)
	default:
		err = installation.Install(ctx)
	}

	return err
}

// Reset resets cluster:
// * destroys all the worker machines
// * kubeadm reset masters
func (i *Installer) Reset(options *Options) error {
	var err error

	ctx := i.createContext(options)

	v := semver.MustParse(i.cluster.Versions.Kubernetes)
	majorMinor := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	switch majorMinor {
	case "1.12":
		err = kube112.Reset(ctx)
	default:
		err = installation.Install(ctx)
	}

	return err
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
