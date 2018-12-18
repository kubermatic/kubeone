package kubeadm

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/templates"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/v1alpha3"
)

//TODO(GvW): Slice of supported sem-versions

// Config returns appropriate version of kubeadm config as YAML
func Config(cluster *config.Cluster, instance *config.HostConfig) (string, error) {
	masterNodes := cluster.Hosts
	if len(masterNodes) == 0 {
		return "", errors.New("cluster does not contain at least one master node")
	}

	v := semver.MustParse(cluster.Versions.Kubernetes)
	majorMinor := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	var (
		clusterCfg, initCfg interface{}
		err                 error
	)

	switch majorMinor {
	case "1.12":
		initCfg, clusterCfg, err = v1alpha3.NewConfig(cluster, instance)
	default:
		err = fmt.Errorf("unsupported Kubernetes version %s", majorMinor)
	}

	if err != nil {
		return "", err
	}

	return templates.KubernetesToYAML([]interface{}{initCfg, clusterCfg})
}
