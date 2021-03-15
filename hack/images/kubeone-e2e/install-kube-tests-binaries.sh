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

set -euox pipefail

declare -A full_versions
full_versions["1.18"]="v1.18.15"
full_versions["1.19"]="v1.19.7"
full_versions["1.20"]="v1.20.2"

root_dir=${KUBETESTS_ROOT:-"/opt/kube-test"}
tmp_root=${TMP_ROOT:-"/tmp/get-kube"}

for version in "${!full_versions[@]}"; do
  full_version="${full_versions[${version}]}"
  directory="${root_dir}/kubernetes-${version}"
  tmp_dir="${tmp_root}/kubernetes-${version}"
  if [[ ! -d "${directory}" ]]; then
    mkdir -p "${tmp_dir}"
    mkdir -p "${directory}"

    curl -L https://gcsweb.k8s.io/gcs/kubernetes-release/release/"${full_version}"/kubernetes.tar.gz -o "${tmp_dir}"/kubernetes.tar.gz
    tar -zxvf "${tmp_dir}"/kubernetes.tar.gz -C "${tmp_dir}"
    mv "${tmp_dir}"/* "${directory}"/

    cd ${directory}/kubernetes
    KUBERNETES_SERVER_ARCH=amd64 KUBE_VERSION="${full_version}" KUBERNETES_DOWNLOAD_TESTS=true KUBERNETES_SKIP_CONFIRM=true ./cluster/get-kube-binaries.sh
    cd -

    find "${directory}" -name "*.tar.gz" -type f -delete
    rm -rf "${directory}"/kubernetes/platforms/linux/arm
    rm -rf "${directory}"/kubernetes/platforms/linux/arm64
    rm -rf "${directory}"/kubernetes/platforms/linux/ppc64le
    rm -rf "${directory}"/kubernetes/platforms/linux/s390x
    rm "${directory}"/kubernetes/platforms/linux/amd64/gendocs
    rm "${directory}"/kubernetes/platforms/linux/amd64/genkubedocs
    rm "${directory}"/kubernetes/platforms/linux/amd64/genman
    rm "${directory}"/kubernetes/platforms/linux/amd64/genswaggertypedocs
    rm "${directory}"/kubernetes/platforms/linux/amd64/genyaml
    rm "${directory}"/kubernetes/platforms/linux/amd64/kubemark
    rm "${directory}"/kubernetes/platforms/linux/amd64/linkcheck
    if [ "$(command -v upx)" ]; then
      upx "${directory}"/kubernetes/platforms/linux/amd64/*
    fi
  fi
done

rm -rf /tmp/get-kube*
