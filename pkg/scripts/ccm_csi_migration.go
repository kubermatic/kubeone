/*
Copyright 2021 The KubeOne Authors.

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

package scripts

import (
	"github.com/MakeNowJust/heredoc/v2"

	"k8c.io/kubeone/pkg/fail"
)

var (
	ccmMigrationRegenerateControlPlaneConfigs = heredoc.Doc(`
		sudo kubeadm {{ .VERBOSE }} init phase control-plane apiserver \
			--config={{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml

		sudo kubeadm {{ .VERBOSE }} init phase control-plane controller-manager \
			--config={{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
		
		sudo kubeadm {{ .VERBOSE }} init phase kubelet-start \
			--config={{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
	`)
)

func CCMMigrationRegenerateControlPlaneConfigs(workdir string, nodeID int, verboseFlag string) (string, error) {
	result, err := Render(ccmMigrationRegenerateControlPlaneConfigs, Data{
		"WORK_DIR": workdir,
		"NODE_ID":  nodeID,
		"VERBOSE":  verboseFlag,
	})

	return result, fail.Runtime(err, "rendering ccmMigrationRegenerateControlPlaneConfigs script")
}
