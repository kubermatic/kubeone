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
full_versions["1.15"]="v1.15.6"
full_versions["1.16"]="v1.16.3"
full_versions["1.17"]="v1.17.0"

root_dir=${KUBETESTS_ROOT:-"/opt/kube-test"}

for version in "${!full_versions[@]}"; do
    full_version="${full_versions[${version}]}"
    directory="${root_dir}/kubernetes-${version}"
    if [[ ! -d "${directory}" ]]; then
        mkdir -p "${directory}"
        cd "${directory}"
        kubetest --extract="${full_version}"
        cd -

        find "${directory}" -name "*.tar.gz" -type f -delete
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
