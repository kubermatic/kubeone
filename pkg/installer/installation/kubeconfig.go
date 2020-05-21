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
	"github.com/kubermatic/kubeone/pkg/scripts"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

func copyKubeconfig(s *state.State) error {
	return s.RunTaskOnAllNodes(func(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
		s.Logger.Infoln("Copying Kubeconfig to home directory…")
		cmd, err := scripts.KubernetesAdminConfig()
		if err != nil {
			return err
		}

		_, _, err = s.Runner.RunRaw(cmd)
		return err
	}, true)
}

func saveKubeconfig(s *state.State) error {
	s.Logger.Info("Downloading kubeconfig…")

	kc, err := kubeconfig.Download(s)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s-kubeconfig", s.Cluster.Name)
	err = ioutil.WriteFile(fileName, kc, 0600)
	return errors.Wrap(err, "error saving kubeconfig file to the local machine")
}
