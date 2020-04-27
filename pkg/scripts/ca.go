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
	copyPKIHomeScriptTemplate = `
mkdir -p {{ .WORK_DIR }}/pki/etcd
sudo cp /etc/kubernetes/pki/ca.crt {{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/ca.key {{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/sa.key {{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/sa.pub {{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/front-proxy-ca.crt {{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/front-proxy-ca.key {{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/etcd/ca.{crt,key} {{ .WORK_DIR }}/pki/etcd/
sudo chown -R "$(id -u):$(id -g)" {{ .WORK_DIR }}
`
)

func CopyPKIHome(workdir string) (string, error) {
	return Render(copyPKIHomeScriptTemplate, Data{
		"WORK_DIR": workdir,
	})
}
