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

set -xe

mkdir /tmp/gobetween
cd /tmp/gobetween
curl -L -o gobetween_0.7.0_linux_amd64.tar.gz \
    https://github.com/yyyar/gobetween/releases/download/0.7.0/gobetween_0.7.0_linux_amd64.tar.gz
tar xvf gobetween_0.7.0_linux_amd64.tar.gz
sudo mv gobetween /usr/local/sbin/gobetween
sudo chown root:root /usr/local/sbin/gobetween

cat <<EOF | sudo tee /etc/systemd/system/gobetween.service
[Unit]
Description=Gobetween - modern LB for cloud era
Documentation=https://github.com/yyyar/gobetween/wiki
After=network.target remote-fs.target nss-lookup.target

[Service]
Type=simple
PIDFile=/run/gobetween.pid
ExecStart=/usr/local/sbin/gobetween -c /etc/gobetween.toml
ExecReload=/bin/kill -s HUP $MAINPID
ExecStop=/bin/kill -s QUIT $MAINPID
PrivateTmp=true
User=nobody
Group=nogroup

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable gobetween.service
sudo systemctl start gobetween.service
