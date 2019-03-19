/*
Copyright 2019 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubeadm

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/templates"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/v1beta1"
	"github.com/kubermatic/kubeone/pkg/util"
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
