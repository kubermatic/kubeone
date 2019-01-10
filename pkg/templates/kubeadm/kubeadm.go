package kubeadm

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/v1alpha3"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/v1beta1"

	"k8s.io/apimachinery/pkg/runtime"
)

// Config returns appropriate version of kubeadm config as YAML
func Config(ctx *util.Context, instance *config.HostConfig) (string, error) {
	cluster := ctx.Cluster
	masterNodes := cluster.Hosts
	if len(masterNodes) == 0 {
		return "", errors.New("cluster does not contain at least one master node")
	}

	v := semver.MustParse(cluster.Versions.Kubernetes)
	majorMinor := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	var configs []runtime.Object
	var err error
	switch majorMinor {
	case "1.12":
		configs, err = v1alpha3.NewConfig(cluster, instance)
	case "1.13":
		configs, err = v1beta1.NewConfig(ctx, instance)
	default:
		err = fmt.Errorf("unsupported Kubernetes version %s", majorMinor)
	}

	if err != nil {
		return "", err
	}

	//TODO: Change KubernetesToYAML to accept runtime.Object instead of empty interface
	var kubernetesToYAMLInput []interface{}
	for _, config := range configs {
		kubernetesToYAMLInput = append(kubernetesToYAMLInput, interface{}(config))
	}
	return templates.KubernetesToYAML(kubernetesToYAMLInput)
}
