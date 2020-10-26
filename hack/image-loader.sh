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

# The image-loader script is used to pull all Docker images used by KubeOne,
# Kubeadm, and Kubernetes, and push them to the the specified private registry.

# WARNING: This script heavily depends on KubeOne and Kubernetes versions.
# You must use the script coming the KubeOne release you've downloaded.

set -euo pipefail

KUBERNETES_VERSION=${KUBERNETES_VERSION}

TARGET_REGISTRY=${TARGET_REGISTRY:-127.0.0.1:5000}
PULL_OPTIONAL_IMAGES=${PULL_OPTIONAL_IMAGES:-true}

# Wrapper around echo to include time
function echodate() {
  # do not use -Is to keep this compatible with macOS
  echo "[$(date +%Y-%m-%dT%H:%M:%S%:z)]" "$@"
}

function retag() {
  local image="$1"

  # Trim registry
  local local_image="$(echo ${image} | cut -d/ -f1 --complement)"

  # Split into name and tag
  local name="$(echo ${local_image} | cut -d: -f1)"
  local tag="$(echo ${local_image} | cut -d: -f2)"

  # Build target image name
  local target_image="${TARGET_REGISTRY}/${name}:${tag}"

  echodate "Retagging \"${image}\" => \"${target_image}\"..."

  # Check does image already exists in the target registry
  if curl -s --fail "http://${TARGET_REGISTRY}/v2/${name}/tags/list" | jq -e ".tags | index(\"${tag}\")" >/dev/null; then
    echodate "Skipping \"${image}\", already exists"
    return
  fi

  docker pull "${image}"
  docker tag "${image}" "${target_image}"
  docker push "${target_image}"

  echodate "Done retagging \"${image}\"."
}

# The script is only supported on Linux because it depends on Kubeadm.
# You can run this script in a Docker container.
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
  echodate "This script works only on Linux because it depends on Kubeadm."
  echodate "As a workaround, you can run the script in a Docker container."
  exit 1
fi

kubeadm=kubeadm
if ! [ -x "$(command -v $kubeadm)" ]; then
  url="https://dl.k8s.io/v${KUBERNETES_VERSION}/bin/linux/amd64/kubeadm"
  kubeadm=/tmp/kubeadm-v${KUBERNETES_VERSION}

  echodate "Downloading kubeadm v$KUBERNETES_VERSION..."
  wget -O "$kubeadm" "$url"
  chmod +x "$kubeadm"
  echodate "Done!"
fi

k8simages=$("$kubeadm" config images list --kubernetes-version="$KUBERNETES_VERSION")

k1images=(
  # Core images deployed by default
  # Canal
  "docker.io/calico/cni:v3.15.1"
  "docker.io/calico/node:v3.15.1"
  "docker.io/calico/kube-controllers:v3.15.1"
  "quay.io/coreos/flannel:v0.13.0"
  # machine-controller
  "docker.io/kubermatic/machine-controller:v1.19.0"
  # metrics-server
  "k8s.gcr.io/metrics-server:v0.3.6"
  # NodeLocalDNSCache
  "k8s.gcr.io/k8s-dns-node-cache:1.15.13"
)

optionalimages=(
  # Optional images - only deployed on user request
  # WeaveNet
  "docker.io/weaveworks/weave-kube:2.5.2"
  "docker.io/weaveworks/weave-npc:2.5.2"
  # DigitalOcean CCM
  "docker.io/digitalocean/digitalocean-cloud-controller-manager:v0.1.23"
  # Hetzner CCM
  "docker.io/hetznercloud/hcloud-cloud-controller-manager:v1.7.0"
  # OpenStack CCM
  "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.17.0"
  # Packet CCM
  "docker.io/packethost/packet-ccm:v1.0.0"
)

for IMAGE in $k8simages; do
  retag "${IMAGE}"
done

for IMAGE in "${k1images[@]}"; do
  retag "${IMAGE}"
done

if [ "$PULL_OPTIONAL_IMAGES" == "false" ]; then
  echodate "Skipping pulling optional images because PULL_OPTIONAL_IMAGES is set to false."
  exit 0
fi

for IMAGE in "${optionalimages[@]}"; do
  retag "${IMAGE}"
done
