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
	"fmt"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates"
	"k8c.io/kubeone/pkg/templates/kubeadm/v1beta1"
)

type kubeadmv1beta1 struct {
	version string
}

func (*kubeadmv1beta1) Config(s *state.State, instance kubeoneapi.HostConfig) (string, error) {
	config, err := v1beta1.NewConfig(s, instance)
	if err != nil {
		return "", err
	}

	return templates.KubernetesToYAML(config)
}

func (*kubeadmv1beta1) ConfigWorker(s *state.State, instance kubeoneapi.HostConfig) (string, error) {
	config, err := v1beta1.NewConfigWorker(s, instance)
	if err != nil {
		return "", err
	}

	return templates.KubernetesToYAML(config)
}
func (k *kubeadmv1beta1) UpgradeLeaderCommand() string {
	return fmt.Sprintf("kubeadm upgrade apply -y %s", k.version)
}

func (*kubeadmv1beta1) UpgradeFollowerCommand() string {
	return "kubeadm upgrade node experimental-control-plane"
}

func (*kubeadmv1beta1) UpgradeStaticWorkerCommand() string {
	return kubeadmUpgradeNodeCommand
}
