package kubeadm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Masterminds/semver"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/v1alpha1"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/v1alpha2"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/v1alpha3"
)

// Config returns appropriate version of kubeadm config as YAML
func Config(cluster *config.Cluster, instance int) (string, error) {
	masterNodes := cluster.Hosts
	if len(masterNodes) == 0 {
		return "", errors.New("cluster does not contain at least one master node")
	}

	v := semver.MustParse(cluster.Versions.Kubernetes)
	majorMinor := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	var (
		cfg interface{}
		err error
	)

	switch majorMinor {
	case "1.10":
		cfg, err = v1alpha1.NewConfig(cluster)
	case "1.11":
		cfg, err = v1alpha2.NewConfig(cluster, instance)
	case "1.12":
		cfg, err = v1alpha3.NewConfig(cluster, instance)
	default:
		err = fmt.Errorf("unsupported Kubernetes version %s", majorMinor)
	}

	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(cfg)

	return string(encoded), err
}
