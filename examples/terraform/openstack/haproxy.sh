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

PKG_MANAGER="noop"

if [ "$(command -v yum)" ]; then
  PKG_MANAGER=yum
  sudo setenforce Permissive
  sudo sed -i /etc/sysconfig/selinux -e s/SELINUX=enforcing/SELINUX=disabled/
fi

[ "$(command -v apt-get)" ] && PKG_MANAGER=apt-get

if [ "$PKG_MANAGER" != "noop" ]; then
  sudo ${PKG_MANAGER} install haproxy -y
fi

sudo systemctl daemon-reload
sudo systemctl enable --now haproxy.service
