#!/usr/bin/env bash

# Copyright 2020 The KubeOne Authors.
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

# This is a wrapper script around GoReleaser used inside CI to automatically
# build KubeOne releases. The scripts should be run only in the Kubermatic CI
# environment, as it requires GitHub credentials.

set -euo pipefail

export GITHUB_TOKEN=$(cat /etc/github/oauth | tr -d '\n')

cd $(dirname $0)/..

git remote add origin git@github.com:kubermatic/kubeone.git

goreleaser release
