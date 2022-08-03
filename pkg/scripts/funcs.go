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

		{{ define "sysctl-k8s" }}
		cat <<EOF | sudo tee /etc/modules-load.d/containerd.conf
		overlay
		br_netfilter
		ip_tables
		EOF
		sudo modprobe overlay
		sudo modprobe br_netfilter
		sudo modprobe ip_tables
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
		{{- if .CILIUM }}
		cat <<EOF | sudo tee /etc/sysctl.d/99-zzz-override_cilium.conf
		# Disable rp_filter on ALL interfaces since it may cause mangled packets to be dropped
		# https://github.com/cilium/cilium/blob/v1.11.1/pkg/datapath/loader/base.go#L244
		net.ipv4.conf.all.rp_filter = 0
		EOF
		{{ end }}
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
	defaultDockerVersion           = "'19.03.*'"
	latestDockerVersion            = "'20.10.*'"
	defaultContainerdVersion       = "'1.5.*'"
	defaultAmazonContainerdVersion = "'1.4.*'"
)
