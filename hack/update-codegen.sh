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

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

# The code generation script takes the following arguments:
# * generators (we use only deepcopy, conversion and defaulter)
# * output path for clientset (we don't generate clienset, therefore it's empty)
# * the internal types dir
# * the external types dir
# * group and versions to generate code for
${CODEGEN_PKG}/generate-internal-groups.sh "deepcopy,conversion,defaulter" \
  "" github.com/kubermatic/kubeone/pkg/apis github.com/kubermatic/kubeone/pkg/apis \
  kubeone:v1alpha1 \
  --output-base "$(dirname ${BASH_SOURCE})/../../../.." \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt
