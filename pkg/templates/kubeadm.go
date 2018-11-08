package templates

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver"
	yaml "gopkg.in/yaml.v2"

	"github.com/kubermatic/kubeone/pkg/manifest"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/v1alpha1"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/v1alpha2"
)

func KubeadmConfig(manifest *manifest.Manifest) (string, error) {
	masterNodes := manifest.Hosts
	if len(masterNodes) == 0 {
		return "", errors.New("manifest does not contain at least one master node")
	}

	v := semver.MustParse(manifest.Versions.Kubernetes)
	majorMinor := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	var (
		cfg interface{}
		err error
	)

	switch majorMinor {
	case "1.10":
		cfg, err = v1alpha1.NewConfig(manifest)
	case "1.11":
		cfg, err = v1alpha2.NewConfig(manifest)
	default:
		err = fmt.Errorf("unsupported Kubernetes version %s", majorMinor)
	}

	if err != nil {
		return "", err
	}

	encoded, err := yaml.Marshal(cfg)

	return string(encoded), err
}
