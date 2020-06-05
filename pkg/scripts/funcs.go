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

import "github.com/Masterminds/semver"

const (
	libraryTemplate = `
{{ define "detect-host-cpu-architecture" }}
HOST_ARCH=""
case $(uname -m) in
x86_64)
	HOST_ARCH="amd64"
	;;
aarch64)
	HOST_ARCH="arm64"
	;;
*)
	echo "unsupported CPU architecture, exiting"
	exit 1
	;;
esac
{{ end }}

{{ define "docker-daemon-config" }}
sudo mkdir -p /etc/docker
cat <<EOF | sudo tee /etc/docker/daemon.json
{
	"exec-opts": ["native.cgroupdriver=systemd"],
	"storage-driver": "overlay2",
	"log-driver": "json-file",
	"log-opts": {
		"max-size": "100m"
	}
}
EOF
{{ end }}

{{ define "sysctl-k8s" }}
sudo mkdir -p /etc/sysctl.d
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
kernel.panic_on_oops = 1
kernel.panic = 10
net.ipv4.ip_forward = 1
vm.overcommit_memory = 1
fs.inotify.max_user_watches = 1048576
EOF
sudo sysctl --system
{{ end }}

{{ define "journald-config" }}
sudo mkdir -p /etc/systemd/journald.conf.d
cat <<EOF | sudo tee /etc/systemd/journald.conf.d/max_disk_use.conf
[Journal]
SystemMaxUse=5G
EOF
sudo systemctl force-reload systemd-journald
{{ end }}
`
)

func aptDockerFunc(v string) (string, error) {
	sver, err := semver.NewVersion(v)
	if err != nil {
		return "", err
	}

	lessThen117, _ := semver.NewConstraint("< 1.17")

	if lessThen117.Check(sver) {
		return "docker-ce=18.03.1~ce~3-0~ubuntu", nil
	}

	// return default
	return "docker-ce=5:19.03.9~3-0~ubuntu-$(lsb_release -cs) docker-ce-cli=5:19.03.9~3-0~ubuntu-$(lsb_release -cs)", nil
}

func yumDockerFunc(v string) (string, error) {
	sver, err := semver.NewVersion(v)
	if err != nil {
		return "", err
	}

	lessThen117, _ := semver.NewConstraint("< 1.17")

	if lessThen117.Check(sver) {
		return "docker-ce-18.03.1.ce-1.el7.centos", nil
	}

	// return default
	return "docker-ce-19.03.9-3.el7 docker-ce-cli-19.03.9-3.el7", nil
}

func cniVersionFunc(kubernetesVersion string) (string, error) {
	s, err := semver.NewVersion(kubernetesVersion)
	if err != nil {
		return "", err
	}

	c, _ := semver.NewConstraint(">= 1.13.0, <= 1.13.4")

	// Validation ensures that the oldest cluster version is 1.13.0.
	// Versions 1.13.0-1.13.4 uses 0.6.0, so it's safe to return 0.6.0
	// if >= 1.13.0, <= 1.13.4 constraint check successes.
	if c.Check(s) {
		return "0.6.0", nil
	}

	// return default
	return "0.7.5", nil
}
