#!/usr/bin/env bash

set -euox pipefail

declare -A full_versions
full_versions["1.14"]="v1.14.8"
full_versions["1.15"]="v1.15.5"
full_versions["1.16"]="v1.16.2"

root_dir=${KUBE_TEST_DIR:-"/opt/kube-test"}

for version in "${!full_versions[@]}"; do
    directory="${root_dir}/kubernetes-${version}"
    if [[ ! -d "${directory}" ]]; then
        full_version="${full_versions[${version}]}"
        mkdir -p ${directory}
        cd ${directory}
        kubetest --extract=${full_version}
        cd -

        find ${directory} -name "*.tar.gz" -type f -delete
        rm -r ${directory}/kubernetes/cluster
        rm ${directory}/kubernetes/platforms/linux/amd64/gendocs
        rm ${directory}/kubernetes/platforms/linux/amd64/genkubedocs
        rm ${directory}/kubernetes/platforms/linux/amd64/genman
        rm ${directory}/kubernetes/platforms/linux/amd64/genswaggertypedocs
        rm ${directory}/kubernetes/platforms/linux/amd64/genyaml
        rm ${directory}/kubernetes/platforms/linux/amd64/kubemark
        rm ${directory}/kubernetes/platforms/linux/amd64/linkcheck
        if [ $(command -v upx) ]; then
            upx ${directory}/kubernetes/platforms/linux/amd64/*
        fi
    fi
done

rm -rf /tmp/get-kube*
