package kubeadm

import (
	"errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/v1beta1"
)

// Config returns appropriate version of kubeadm config as YAML
func Config(ctx *util.Context, instance *config.HostConfig) (string, error) {
	cluster := ctx.Cluster
	masterNodes := cluster.Hosts
	if len(masterNodes) == 0 {
		return "", errors.New("cluster does not contain at least one master node")
	}

	configs, err := v1beta1.NewConfig(ctx, instance)
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
