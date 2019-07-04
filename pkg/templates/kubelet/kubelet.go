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

package kubelet

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/runner"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeletconfig "k8s.io/kubelet/config/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kyaml "sigs.k8s.io/yaml"
)

const (
	kubeletConfigMapName = "kubelet-config-%d.%d"

	deployKubeletConfig = `
sudo kubeadm alpha kubelet config download
sudo systemctl restart kubelet
`
)

// GetConfig fetch kubelet ConfigMap and unmarshal it to strtucutre
func GetConfig(s *state.State) (*kubeletconfig.KubeletConfiguration, error) {
	kver, err := semver.NewVersion(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse kubernetes version")
	}

	configmapName := fmt.Sprintf(kubeletConfigMapName, kver.Major(), kver.Minor())
	objkey := client.ObjectKey{
		Name:      configmapName,
		Namespace: metav1.NamespaceSystem,
	}
	configMap := corev1.ConfigMap{}
	ctx := context.Background()

	if errGet := s.DynamicClient.Get(ctx, objkey, &configMap); errGet != nil {
		return nil, errors.Wrap(errGet, "failed to get kubelet configmap")
	}

	kubletConfigString, ok := configMap.Data["kubelet"]
	if !ok {
		return nil, errors.New("no kubelet config data is found")
	}

	konfig := &kubeletconfig.KubeletConfiguration{}
	err = kyaml.UnmarshalStrict([]byte(kubletConfigString), konfig)
	return konfig, errors.Wrap(err, "failed to unmarshal KubeletConfiguration")
}

// SaveConfig upload KubeletConfiguration to ConfigMap, update kubelet config file
// and restart systemd kubelet unit
func SaveConfig(s *state.State, konfig *kubeletconfig.KubeletConfiguration) error {
	kver, err := semver.NewVersion(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "failed to parse kubernetes version")
	}

	kubletConfigBuf, err := kyaml.Marshal(konfig)
	if err != nil {
		return errors.Wrap(err, "failed to marshal KubeletConfiguration")
	}

	configmapName := fmt.Sprintf(kubeletConfigMapName, kver.Major(), kver.Minor())
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: metav1.NamespaceSystem,
		},
	}
	kubletConfig := string(kubletConfigBuf)
	ctx := context.Background()

	_, err = controllerutil.CreateOrUpdate(ctx,
		s.DynamicClient,
		&configMap,
		func(_ runtime.Object) error {
			if configMap.ObjectMeta.CreationTimestamp.IsZero() {
				// this configmap MUST be present
				return errors.Errorf("%q config map not found", configmapName)
			}

			// replace config
			configMap.Data["kubelet"] = kubletConfig

			// let it update ConfigMap
			return nil
		})

	return errors.Wrap(err, "failed to update kubelet configmap")
}

// DeployConfig download kubelet config on the node and restart kubelet
func DeployConfig(s *state.State, host *kubeone.HostConfig, _ ssh.Connection) error {
	s.Logger.WithField("node", host.PublicAddress).Info("deploying kubelet config…")

	_, _, runErr := s.Runner.Run(deployKubeletConfig, runner.TemplateVariables{})
	return errors.Wrap(runErr, "failed to deploy")
}
