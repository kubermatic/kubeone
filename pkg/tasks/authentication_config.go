/*
Copyright 2026 The KubeOne Authors.

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

package tasks

import (
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/configupload"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/state"
)

const authenticationConfigFilename = "cfg/authentication-config.yaml"

// LoadAuthenticationConfiguration reads the AuthenticationConfiguration manifest referenced by the
// KubeOneCluster and validates it.
func LoadAuthenticationConfiguration(s *state.State) (string, error) {
	b, err := configupload.ReadFilePath(s.Cluster.Features.AuthenticationConfiguration.ConfigFilePath, s.ManifestFilePath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// updateAuthenticationConfig uploads only the AuthenticationConfiguration to the control plane nodes. The
// kube-apiserver watches the file and reloads it automatically, so no restart is needed.
func updateAuthenticationConfig(s *state.State) error {
	config, err := LoadAuthenticationConfiguration(s)
	if err != nil {
		return err
	}

	cfg := configupload.NewConfiguration()
	cfg.AddFile(authenticationConfigFilename, config)

	return s.RunTaskOnControlPlane(func(s *state.State, _ *kubeoneapi.HostConfig, conn executor.Interface) error {
		s.Logger.Infoln("Updating authentication configuration...")

		if err := cfg.UploadTo(conn, s.WorkDir); err != nil {
			return err
		}

		cmd, err := scripts.SaveAuthenticationConfig(s.WorkDir)
		if err != nil {
			return err
		}

		_, _, err = s.Runner.RunRaw(cmd)

		return fail.SSH(err, "saving authentication configuration")
	}, state.RunParallel)
}
