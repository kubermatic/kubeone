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

set -euo pipefail

if [[ ${OSTYPE} == *"darwin"* ]]; then
  jsondate=$(gdate --rfc-3339=seconds | sed 's/ /T/')
else
  jsondate=$(date --rfc-3339=seconds | sed 's/ /T/')
fi

basedir="./docs/api_reference"
mkdir -p "${basedir}"

genVersionedDoc() {
  local version=$1
  local docfile="${basedir}/${version}.en.md"

  cat << EOF > "${docfile}"
+++
title = "${version} API Reference"
date = ${jsondate}
weight = 11
+++
EOF

  echo -e "${version} API"
  cat << EOF >> "${docfile}"
## ${version}

EOF

  find ./pkg/apis/kubeone/"${version}" -name '*types.go' -print0 |
    xargs go run ./hack/apidoc-gen/main.go -section-link="#${version}" >> "${docfile}"
}

genVersionedDoc "v1beta1"
