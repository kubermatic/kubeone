package installer

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/installer/version/v1_10"
	"github.com/kubermatic/kubeone/pkg/installer/version/v1_11"
	"github.com/kubermatic/kubeone/pkg/manifest"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

type installer struct {
	manifest *manifest.Manifest
	logger   *logrus.Logger
}

func NewInstaller(manifest *manifest.Manifest, logger *logrus.Logger) *installer {
	return &installer{
		manifest: manifest,
		logger:   logger,
	}
}

func (i *installer) Run(verbose bool) (*Result, error) {
	var err error

	ctx := &util.Context{
		Manifest:      i.manifest,
		Connector:     ssh.NewConnector(),
		Configuration: util.NewConfiguration(),
		WorkDir:       "kubermatic-installer",
		Verbose:       verbose,
		Logger:        i.logger,
	}

	v := semver.MustParse(i.manifest.Versions.Kubernetes)
	majorMinor := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	switch majorMinor {
	case "1.10":
		err = v1_10.Install(ctx)
	case "1.11":
		err = v1_11.Install(ctx)
	default:
		err = fmt.Errorf("unsupported Kubernetes version %s", majorMinor)
	}

	return nil, err
}
