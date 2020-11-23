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
	"encoding/json"

	"github.com/Masterminds/semver"
)

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
{{ dockerCfg .DOCKER_INSECURE_REGISTRY }}
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

type dockerConfig struct {
	ExecOpts           []string          `json:"exec-opts,omitempty"`
	StorageDriver      string            `json:"storage-driver,omitempty"`
	LogDriver          string            `json:"log-driver,omitempty"`
	LogOpts            map[string]string `json:"log-opts,omitempty"`
	InsecureRegistries []string          `json:"insecure-registries,omitempty"`
}

func dockerCfg(insecureRegistry string) (string, error) {
	cfg := dockerConfig{
		ExecOpts:      []string{"native.cgroupdriver=systemd"},
		StorageDriver: "overlay2",
		LogDriver:     "json-file",
		LogOpts: map[string]string{
			"max-size": "100m",
		},
	}
	if insecureRegistry != "" {
		cfg.InsecureRegistries = []string{insecureRegistry}
	}

	b, err := json.MarshalIndent(cfg, "", "	")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func aptDockerFunc(v string) (string, error) {
	sver, err := semver.NewVersion(v)
	if err != nil {
		return "", err
	}

	lessThen117, _ := semver.NewConstraint("< 1.17")

	if lessThen117.Check(sver) {
		return "docker-ce=5:18.09.9~3-0~ubuntu-bionic docker-ce-cli=5:18.09.9~3-0~ubuntu-bionic", nil
	}

	// return default
	return "docker-ce=5:19.03.12~3-0~ubuntu-bionic docker-ce-cli=5:19.03.12~3-0~ubuntu-bionic", nil
}

func yumDockerFunc(v string) (string, error) {
	sver, err := semver.NewVersion(v)
	if err != nil {
		return "", err
	}

	lessThen117, _ := semver.NewConstraint("< 1.17")

	if lessThen117.Check(sver) {
		return "docker-ce-18.09.9-3.el7 docker-ce-cli-18.09.9-3.el7", nil
	}

	// return default
	return "docker-ce-19.03.12-3.el7 docker-ce-cli-19.03.12-3.el7", nil
}

func amznYumDockerFunc(v string) (string, error) {
	sver, err := semver.NewVersion(v)
	if err != nil {
		return "", err
	}

	lessThen117, _ := semver.NewConstraint("< 1.17")

	if lessThen117.Check(sver) {
		return "docker-18.09.9ce-2.amzn2", nil
	}

	// return default
	return "docker-19.03.13ce-1.amzn2", nil
}
