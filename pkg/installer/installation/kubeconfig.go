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

package installation

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/kubeconfig"
	"github.com/kubermatic/kubeone/pkg/runner"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

func copyKubeconfig(s *state.State) error {
	return s.RunTaskOnAllNodes(func(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
		s.Logger.Infoln("Copying Kubeconfig to home directory…")

		_, _, err := s.Runner.Run(`
mkdir -p $HOME/.kube/
sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
`, runner.TemplateVariables{})
		if err != nil {
			return err
		}

		return nil
	}, true)
}

func saveKubeconfig(s *state.State) error {
	kc, err := kubeconfig.Download(s.Cluster)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s-kubeconfig", s.Cluster.Name)
	err = ioutil.WriteFile(fileName, kc, 0644)
	return errors.Wrap(err, "error saving kubeconfig file to the local machine")
}
