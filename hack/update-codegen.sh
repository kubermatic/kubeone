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

set -eu -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}
TERRAFORM_DOCS="go run github.com/terraform-docs/terraform-docs@v0.16.0"

source "${SCRIPT_ROOT}/hack/lib.sh"
source "${CODEGEN_PKG}/kube_codegen.sh"

echodate "Generating Terraform documentation..."
for input in examples/terraform/*/README.md.in; do
  dir=$(dirname "$input")
  target=$(basename "$input" .in)
  full_target="$dir/$target"
  echo "$full_target"
  cat "$input" > "$dir/$target"
  $TERRAFORM_DOCS --lockfile=false md "$dir" >> "$dir/$target"
done

echodate "Generating Kubernetes helpers..."
kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/apis/kubeone"

echodate "Generating Go code..."
make gogenerate
