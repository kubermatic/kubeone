#!/usr/bin/env bash

# Copyright 2019 The KubeOne Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script is mostly used in CI
# It installs dependencies and starts the tests

set -euf -o pipefail

GOBETWEEN_VERSION=0.7.0

noop() { : "didn't detected package manager, noop"; }

PKG_MANAGER="noop"

[ "$(command -v yum)" ] && PKG_MANAGER=yum
[ "$(command -v apt-get)" ] && PKG_MANAGER=apt-get

sudo ${PKG_MANAGER} install tar -y

mkdir -p /tmp/gobetween
cd /tmp/gobetween
curl -L -o gobetween_${GOBETWEEN_VERSION}_linux_amd64.tar.gz \
    https://github.com/yyyar/gobetween/releases/download/${GOBETWEEN_VERSION}/gobetween_${GOBETWEEN_VERSION}_linux_amd64.tar.gz
tar xvf gobetween_${GOBETWEEN_VERSION}_linux_amd64.tar.gz
sudo mkdir -p /opt/bin
sudo mv gobetween /opt/bin/gobetween
sudo chown root:root /opt/bin/gobetween

cat <<EOF | sudo tee /etc/systemd/system/gobetween.service
[Unit]
Description=Gobetween - modern LB for cloud era
Documentation=https://github.com/yyyar/gobetween/wiki
After=network.target remote-fs.target nss-lookup.target

[Service]
Type=simple
ExecStart=/opt/bin/gobetween -c /etc/gobetween.toml
PrivateTmp=true
User=nobody

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable gobetween.service
sudo systemctl start gobetween.service
