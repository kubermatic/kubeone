/*
Copyright 2024 The KubeOne Authors.

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
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/runner"
	"k8c.io/kubeone/pkg/state"
)

var systemFiles = map[string][]string{
	"kubelet": {
		"/lib/systemd/system/kubelet.service",
		"/lib/systemd/system/kubelet.service.d/*",
		"/var/lib/kubelet/config.yaml",
	},
	"cni": {
		"/etc/cni/net.d/*",
	},
}

func fixFilePermissions(s *state.State) error {
	s.Logger.Info("Fixing permissions of the kubernetes system files")

	return s.RunTaskOnAllNodes(func(ctx *state.State, _ *kubeoneapi.HostConfig, _ executor.Interface) error {
		for _, pathList := range systemFiles {
			for _, path := range pathList {
				args := runner.TemplateVariables{"PATH": path}
				if ctx.Verbose {
					args["VERBOSE"] = "--verbose"
				}

				if _, _, err := ctx.Runner.Run("sudo chmod 600 {{ with .VERBOSE }}{{ . }} {{ end }}{{ .PATH }}", args); err != nil {
					return err
				}
			}
		}

		return nil
	}, state.RunParallel)
}
