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
	hostnameScript = `
fqdn=$(hostname -f)
[ "$fqdn" = localhost ] && fqdn=$(hostname)
echo "$fqdn"
`

	verifyPrerequisitesScript = `
# Check is Docker installed
if ! type docker &>/dev/null; then exit 1; fi
# Check is Kubelet installed
if ! type kubelet &>/dev/null; then exit 1; fi
# Check is Kubeadm installed
if ! type kubeadm &>/dev/null; then exit 1; fi
# Check do Kubernetes directories and files exist
if [[ ! -d "/etc/kubernetes/manifests" ]]; then exit 1; fi
if [[ ! -d "/etc/kubernetes/pki" ]]; then exit 1; fi
if [[ ! -f "/etc/kubernetes/kubelet.conf" ]]; then exit 1; fi
# Check are kubelet running
if ! sudo systemctl is-active --quiet kubelet &>/dev/null; then exit 1; fi
`
)

func Hostname() string {
	return hostnameScript
}

func OSID() string {
	return "source /etc/os-release && echo -n $ID"
}

func VerifyPrerequisites() (string, error) {
	return Render(verifyPrerequisitesScript, nil)
}
