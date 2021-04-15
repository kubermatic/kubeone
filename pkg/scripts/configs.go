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

	"k8c.io/kubeone/pkg/certificate/cabundle"
)

var (
	cloudConfigScriptTemplate = heredoc.Doc(`
		sudo mkdir -p /etc/systemd/system/kubelet.service.d/ /etc/kubernetes
		sudo mv {{ .WORK_DIR }}/cfg/cloud-config /etc/kubernetes/cloud-config
		sudo chown root:root /etc/kubernetes/cloud-config
		sudo chmod 600 /etc/kubernetes/cloud-config
	`)

	auditPolicyScriptTemplate = heredoc.Doc(`
		if [[ -f "{{ .WORK_DIR }}/cfg/audit-policy.yaml" ]]; then
			sudo mkdir -p /etc/kubernetes/audit
			sudo mv {{ .WORK_DIR }}/cfg/audit-policy.yaml /etc/kubernetes/audit/policy.yaml
			sudo chown root:root /etc/kubernetes/audit/policy.yaml
		fi
	`)

	podNodeSelectorConfigTemplate = heredoc.Doc(`
		if [[ -f "{{ .WORK_DIR }}/cfg/podnodeselector.yaml" ]]; then
			sudo mkdir -p /etc/kubernetes/admission
			sudo mv {{ .WORK_DIR }}/cfg/podnodeselector.yaml /etc/kubernetes/admission/podnodeselector.yaml
			sudo mv {{ .WORK_DIR }}/cfg/admission-config.yaml /etc/kubernetes/admission/admission-config.yaml
			sudo chown root:root /etc/kubernetes/admission/podnodeselector.yaml
			sudo chown root:root /etc/kubernetes/admission/admission-config.yaml
		fi
	`)

	caBundleTemplate = heredoc.Doc(`
		sudo mkdir -p {{ .CA_CERTS_DIR }}
		sudo mv {{ .WORK_DIR }}/ca-certs/{{ .CA_BUNDLE_FILENAME }} {{ .CA_CERTS_DIR }}
		sudo chown -R root:root {{ .CA_CERTS_DIR }}
	`)
)

func SaveCloudConfig(workdir string) (string, error) {
	return Render(cloudConfigScriptTemplate, Data{
		"WORK_DIR": workdir,
	})
}

func SaveAuditPolicyConfig(workdir string) (string, error) {
	return Render(auditPolicyScriptTemplate, Data{
		"WORK_DIR": workdir,
	})
}

func SavePodNodeSelectorConfig(workdir string) (string, error) {
	return Render(podNodeSelectorConfigTemplate, Data{
		"WORK_DIR": workdir,
	})
}

func SaveCABundle(workdir string) (string, error) {
	return Render(caBundleTemplate, Data{
		"CA_BUNDLE_FILENAME": cabundle.FileName,
		"CA_CERTS_DIR":       cabundle.CertsDir,
		"WORK_DIR":           workdir,
	})
}
