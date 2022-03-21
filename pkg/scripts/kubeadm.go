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

package scripts

import (
	"github.com/MakeNowJust/heredoc/v2"

	"k8c.io/kubeone/pkg/fail"
)

var (
	kubeadmJoinScriptTemplate = heredoc.Doc(`
		[[ -f /etc/kubernetes/admin.conf ]] && exit 0

		sudo kubeadm {{ .VERBOSE }} join \
			--config={{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
	`)

	kubeadmWorkerJoinScriptTemplate = heredoc.Doc(`
		[[ -f /etc/kubernetes/kubelet.conf ]] && exit 0

		sudo kubeadm {{ .VERBOSE }} join \
			--config={{ .WORK_DIR }}/cfg/worker_{{ .NODE_ID }}.yaml
	`)

	kubeadmCertScriptTemplate = heredoc.Doc(`
		sudo kubeadm {{ .VERBOSE }} init phase certs all \
			--config={{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
	`)

	kubeadmInitScriptTemplate = heredoc.Doc(`
		if [[ -f /etc/kubernetes/admin.conf ]]; then
			sudo kubeadm {{ .VERBOSE }} token create {{ .TOKEN }} --ttl {{ .TOKEN_DURATION }}
			exit 0;
		fi

		sudo kubeadm {{ .VERBOSE }} init {{ if .SKIP_PHASE }}--skip-phases={{ .SKIP_PHASE }} {{ end}}--config={{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
	`)

	kubeadmResetScriptTemplate = heredoc.Doc(`
		sudo kubeadm {{ .VERBOSE }} reset --force || true
		sudo rm -f /etc/kubernetes/cloud-config
		sudo rm -rf /etc/kubernetes/admission
		sudo rm -rf /etc/kubernetes/encryption-providers
		sudo rm -rf /var/lib/etcd/
		sudo rm -rf "{{ .WORK_DIR }}"
		sudo rm -rf /etc/kubeone
	`)

	kubeadmUpgradeScriptTemplate = heredoc.Doc(`
		sudo {{ .KUBEADM_UPGRADE }}{{ if .LEADER }} --config={{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml{{ end }}
	`)

	kubeadmPauseImageVersionScriptTemplate = heredoc.Doc(`
		sudo kubeadm config images list --kubernetes-version={{ .KUBERNETES_VERSION }} |
			grep "k8s.gcr.io/pause" |
			cut -d ":" -f2
	`)
)

func KubeadmJoin(workdir string, nodeID int, verboseFlag string) (string, error) {
	result, err := Render(kubeadmJoinScriptTemplate, Data{
		"WORK_DIR": workdir,
		"NODE_ID":  nodeID,
		"VERBOSE":  verboseFlag,
	})

	return result, fail.Runtime(err, "rendering kubeadmJoinScriptTemplate script")
}

func KubeadmJoinWorker(workdir string, nodeID int, verboseFlag string) (string, error) {
	result, err := Render(kubeadmWorkerJoinScriptTemplate, Data{
		"WORK_DIR": workdir,
		"NODE_ID":  nodeID,
		"VERBOSE":  verboseFlag,
	})

	return result, fail.Runtime(err, "rendering kubeadmWorkerJoinScriptTemplate script")
}

func KubeadmCert(workdir string, nodeID int, verboseFlag string) (string, error) {
	result, err := Render(kubeadmCertScriptTemplate, Data{
		"WORK_DIR": workdir,
		"NODE_ID":  nodeID,
		"VERBOSE":  verboseFlag,
	})

	return result, fail.Runtime(err, "rendering kubeadmCertScriptTemplate script")
}

func KubeadmInit(workdir string, nodeID int, verboseFlag, token, tokenTTL string, skipPhases string) (string, error) {
	result, err := Render(kubeadmInitScriptTemplate, Data{
		"WORK_DIR":       workdir,
		"NODE_ID":        nodeID,
		"VERBOSE":        verboseFlag,
		"TOKEN":          token,
		"TOKEN_DURATION": tokenTTL,
		"SKIP_PHASE":     skipPhases,
	})

	return result, fail.Runtime(err, "rendering kubeadmInitScriptTemplate script")
}

func KubeadmReset(verboseFlag, workdir string) (string, error) {
	result, err := Render(kubeadmResetScriptTemplate, Data{
		"VERBOSE":  verboseFlag,
		"WORK_DIR": workdir,
	})

	return result, fail.Runtime(err, "rendering kubeadmResetScriptTemplate script")
}

func KubeadmUpgrade(kubeadmCmd, workdir string, leader bool, nodeID int) (string, error) {
	result, err := Render(kubeadmUpgradeScriptTemplate, map[string]interface{}{
		"KUBEADM_UPGRADE": kubeadmCmd,
		"WORK_DIR":        workdir,
		"NODE_ID":         nodeID,
		"LEADER":          leader,
	})

	return result, fail.Runtime(err, "rendering kubeadmUpgradeScriptTemplate script")
}

func KubeadmPauseImageVersion(kubernetesVersion string) (string, error) {
	result, err := Render(kubeadmPauseImageVersionScriptTemplate, map[string]interface{}{
		"KUBERNETES_VERSION": kubernetesVersion,
	})

	return result, fail.Runtime(err, "rendering kubeadmPauseImageVersionScriptTemplate script")
}
