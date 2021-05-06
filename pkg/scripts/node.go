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
	hostnameScript = heredoc.Doc(`
		fqdn=$(hostname -f)
		[ "$fqdn" = localhost ] && fqdn=$(hostname)
		echo "$fqdn"
	`)

	drainNodeScriptTemplate = heredoc.Doc(`
		sudo KUBECONFIG=/etc/kubernetes/admin.conf \
		kubectl drain {{ .NODE_NAME }} --ignore-daemonsets --delete-local-data
	`)

	restartKubeAPIServerCrictlTemplate = heredoc.Doc(`
		apiserver_id=$(sudo crictl ps --name=kube-apiserver -q)
		[ -z "$apiserver_id" ] && exit 1
	{{ if .ENSURE }}
		sudo crictl rm "$apiserver_id"
		sleep 30
	{{ else }}
		sudo crictl logs "$apiserver_id" > /tmp/kube-apiserver.log 2>&1
		if sudo grep -q "etcdserver: no leader" /tmp/kube-apiserver.log; then
			sudo crictl stop "$apiserver_id"
			sudo crictl rm "$apiserver_id"
			sleep 10
		fi
	{{ end }}
	`)
)

func DrainNode(nodeName string) (string, error) {
	return Render(drainNodeScriptTemplate, Data{
		"NODE_NAME": nodeName,
	})
}

func Hostname() string {
	return hostnameScript
}

func RestartKubeAPIServerCrictl(ensure bool) (string, error) {
	return Render(restartKubeAPIServerCrictlTemplate, Data{
		"ENSURE": ensure,
	})
}
