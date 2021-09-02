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

import "github.com/MakeNowJust/heredoc/v2"

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

		sudo kubeadm {{ .VERBOSE }} init --config={{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
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

	kubeadmUpgradeLeaderScriptTemplate = heredoc.Doc(`
		sudo {{ .KUBEADM_UPGRADE }} --config={{ .WORK_DIR }}/cfg/master_0.yaml
	`)

	kubeadmPauseImageVersionScriptTemplate = heredoc.Doc(`
		sudo kubeadm config images list --kubernetes-version={{ .KUBERNETES_VERSION }} |
			grep "k8s.gcr.io/pause" |
			cut -d ":" -f2
	`)
)

func KubeadmJoin(workdir string, nodeID int, verboseFlag string) (string, error) {
	return Render(kubeadmJoinScriptTemplate, Data{
		"WORK_DIR": workdir,
		"NODE_ID":  nodeID,
		"VERBOSE":  verboseFlag,
	})
}

func KubeadmJoinWorker(workdir string, nodeID int, verboseFlag string) (string, error) {
	return Render(kubeadmWorkerJoinScriptTemplate, Data{
		"WORK_DIR": workdir,
		"NODE_ID":  nodeID,
		"VERBOSE":  verboseFlag,
	})
}

func KubeadmCert(workdir string, nodeID int, verboseFlag string) (string, error) {
	return Render(kubeadmCertScriptTemplate, Data{
		"WORK_DIR": workdir,
		"NODE_ID":  nodeID,
		"VERBOSE":  verboseFlag,
	})
}

func KubeadmInit(workdir string, nodeID int, verboseFlag, token, tokenTTL string) (string, error) {
	return Render(kubeadmInitScriptTemplate, Data{
		"WORK_DIR":       workdir,
		"NODE_ID":        nodeID,
		"VERBOSE":        verboseFlag,
		"TOKEN":          token,
		"TOKEN_DURATION": tokenTTL,
	})
}

func KubeadmReset(verboseFlag, workdir string) (string, error) {
	return Render(kubeadmResetScriptTemplate, Data{
		"VERBOSE":  verboseFlag,
		"WORK_DIR": workdir,
	})
}

func KubeadmUpgradeLeader(kubeadmCmd, workdir string) (string, error) {
	return Render(kubeadmUpgradeLeaderScriptTemplate, map[string]interface{}{
		"KUBEADM_UPGRADE": kubeadmCmd,
		"WORK_DIR":        workdir,
	})
}

func KubeadmPauseImageVersion(kubernetesVersion string) (string, error) {
	return Render(kubeadmPauseImageVersionScriptTemplate, map[string]interface{}{
		"KUBERNETES_VERSION": kubernetesVersion,
	})
}
