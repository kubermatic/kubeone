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

const (
	kubeadmJoinScriptTemplate = `
if [[ -f /etc/kubernetes/admin.conf ]]; then exit 0; fi

sudo kubeadm join {{ .VERBOSE }} \
	--config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`

	kubeadmCertScriptTemplate = `
if [[ -d ./{{ .WORK_DIR }}/pki ]]; then
   sudo rsync -av ./{{ .WORK_DIR }}/pki/ /etc/kubernetes/pki/
   sudo chown -R root:root /etc/kubernetes
   rm -rf ./{{ .WORK_DIR }}/pki
fi
sudo kubeadm {{ .VERBOSE }} init phase certs all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`

	kubeadmInitScriptTemplate = `
if [[ -f /etc/kubernetes/admin.conf ]]; then
	sudo kubeadm {{ .VERBOSE }} token create {{ .TOKEN }} --ttl {{ .TOKEN_DURATION }}
	exit 0;
fi
sudo kubeadm {{ .VERBOSE }} init --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`

	kubeadmResetScriptTemplate = `
sudo kubeadm {{ .VERBOSE }} reset --force || true
sudo rm -f /etc/kubernetes/cloud-config
sudo rm -rf /var/lib/etcd/
rm -rf "{{ .WORK_DIR }}"
`

	kubeadmUpgradeLeaderScriptTemplate = `
sudo {{ .KUBEADM_UPGRADE }} --config=./{{ .WORK_DIR }}/cfg/master_0.yaml`
)

func KubeadmJoin(workdir string, nodeID int, verboseFlag string) (string, error) {
	return Render(kubeadmJoinScriptTemplate, Data{
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
