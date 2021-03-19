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
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/MakeNowJust/heredoc/v2"
)

var (
	libraryTemplate = heredoc.Doc(`
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
		{{ dockerCfg .INSECURE_REGISTRY }}
		EOF
		{{ end }}

		{{ define "sysctl-k8s" }}
		cat <<EOF | sudo tee /etc/modules-load.d/containerd.conf
		overlay
		br_netfilter
		EOF
		sudo modprobe overlay
		sudo modprobe br_netfilter
		sudo mkdir -p /etc/sysctl.d
		cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
		fs.inotify.max_user_watches         = 1048576
		kernel.panic                        = 10
		kernel.panic_on_oops                = 1
		net.bridge.bridge-nf-call-ip6tables = 1
		net.bridge.bridge-nf-call-iptables  = 1
		net.ipv4.ip_forward                 = 1
		net.netfilter.nf_conntrack_max      = 1000000
		vm.overcommit_memory                = 1
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
	`)
)

const (
	defaultDockerVersion           = "19.03.14"
	defaultAmazonDockerVersion     = "19.03.13"
	defaultLegacyDockerVersion     = "18.09.9"
	defaultContainerdVersion       = "1.4.3"
	defaultAmazonContainerdVersion = "1.4.1"
	defaultAmazonCrictlVersion     = "1.13.0"
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

type containerdConfig struct {
	Version int                    `toml:"version"`
	Metrics *containerdMetrics     `toml:"metrics"`
	Plugins map[string]interface{} `toml:"plugins"`
}

type containerdMetrics struct {
	Address string `toml:"address"`
}

type containerdCRIPlugin struct {
	Containerd *containerdCRISettings `toml:"containerd"`
	Registry   *containerdCRIRegistry `toml:"registry"`
}

type containerdCRISettings struct {
	Runtimes map[string]containerdCRIRuntime `toml:"runtimes"`
}

type containerdCRIRuntime struct {
	RuntimeType string      `toml:"runtime_type"`
	Options     interface{} `toml:"options"`
}

type containerdCRIRuncOptions struct {
	SystemdCgroup bool
}

type containerdCRIRegistry struct {
	Mirrors map[string]containerdMirror `toml:"mirrors"`
}

type containerdMirror struct {
	Endpoint []string `toml:"endpoint"`
}

func containerdCfg(insecureRegistry string) (string, error) {
	criPlugin := containerdCRIPlugin{
		Containerd: &containerdCRISettings{
			Runtimes: map[string]containerdCRIRuntime{
				"runc": {
					RuntimeType: "io.containerd.runc.v2",
					Options: containerdCRIRuncOptions{
						SystemdCgroup: true,
					},
				},
			},
		},
		Registry: &containerdCRIRegistry{
			Mirrors: map[string]containerdMirror{
				"docker.io": {
					Endpoint: []string{"https://registry-1.docker.io"},
				},
			},
		},
	}

	if insecureRegistry != "" {
		criPlugin.Registry.Mirrors[insecureRegistry] = containerdMirror{
			Endpoint: []string{fmt.Sprintf("http://%s", insecureRegistry)},
		}
	}

	cfg := containerdConfig{
		Version: 2,
		Metrics: &containerdMetrics{
			// metrics available at http://127.0.0.1:1338/v1/metrics
			Address: "127.0.0.1:1338",
		},

		Plugins: map[string]interface{}{
			"io.containerd.grpc.v1.cri": criPlugin,
		},
	}

	var buf strings.Builder
	enc := toml.NewEncoder(&buf)
	enc.Indent = ""
	err := enc.Encode(cfg)

	return buf.String(), err
}
