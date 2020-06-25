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
	drainNodeScriptTemplate = `
sudo KUBECONFIG=/etc/kubernetes/admin.conf \
    kubectl drain {{ .NODE_NAME }} --ignore-daemonsets --delete-local-data
`

	uncordonNodeScriptTemplate = `
sudo KUBECONFIG=/etc/kubernetes/admin.conf \
    kubectl uncordon {{ .NODE_NAME }}
`
)

func DrainNode(nodeName string) (string, error) {
	return Render(drainNodeScriptTemplate, Data{
		"NODE_NAME": nodeName,
	})
}

func UncordonNode(nodeName string) (string, error) {
	return Render(uncordonNodeScriptTemplate, Data{
		"NODE_NAME": nodeName,
	})
}
