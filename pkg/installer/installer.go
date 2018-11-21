package installer

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/installer/version/kube110"
	"github.com/kubermatic/kubeone/pkg/installer/version/kube111"
	"github.com/kubermatic/kubeone/pkg/installer/version/kube112"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// Options groups the various possible options for running
// the Kubernetes installation.
type Options struct {
	Verbose    bool
	BackupFile string
}

type installer struct {
	cluster *config.Cluster
	logger  *logrus.Logger
}

// NewInstaller returns a new installer, responsible for dispatching
// between the different supported Kubernetes versions and running the
// installation procedure.
func NewInstaller(cluster *config.Cluster, logger *logrus.Logger) *installer {
	return &installer{
		cluster: cluster,
		logger:  logger,
	}
}

func (i *installer) Install(options *Options) (*Result, error) {
	var err error

	ctx := i.createContext(options)

	v := semver.MustParse(i.cluster.Versions.Kubernetes)
	majorMinor := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	switch majorMinor {
	case "1.10":
		err = kube110.Install(ctx)
	case "1.11":
		err = kube111.Install(ctx)
	case "1.12":
		err = kube112.Install(ctx)
	default:
		err = fmt.Errorf("unsupported Kubernetes version %s", majorMinor)
	}

	return nil, err
}

func (i *installer) Reset(verbose bool) (*Result, error) {
	var err error

	ctx := i.createContext(&Options{
		Verbose: verbose,
	})

	v := semver.MustParse(i.cluster.Versions.Kubernetes)
	majorMinor := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	switch majorMinor {
	case "1.10":
		err = kube110.Reset(ctx)
	case "1.11":
		err = kube111.Reset(ctx)
	case "1.12":
		err = kube112.Reset(ctx)
	default:
		err = fmt.Errorf("unsupported Kubernetes version %s", majorMinor)
	}

	return nil, err
}

func (i *installer) createContext(options *Options) *util.Context {
	return &util.Context{
		Cluster:       i.cluster,
		Connector:     ssh.NewConnector(),
		Configuration: util.NewConfiguration(),
		WorkDir:       "kubeone",
		Logger:        i.logger,
		Verbose:       options.Verbose,
		BackupFile:    options.BackupFile,
	}
}
