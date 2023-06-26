#!/usr/bin/env bash

# Copyright 2022 The KubeOne Authors.
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

### This script is used to build the KubeOne Docker image.
### The KubeOne image is currently experimental and NOT supposed to be used
### in production.

set -euo pipefail

# Required for signal propagation to work so
# the cleanup trap gets executed when the script
# receives a SIGINT
set -o monitor

source $(dirname $0)/../lib.sh

IMAGE="${IMAGE:-quay.io/kubermatic/kubeone}"
ARCHITECTURES=${ARCHITECTURES:-amd64 arm64}
NOMOCK=${NOMOCK:-false}

PRIMARY_TAG="$(git rev-parse HEAD | tr -d '\n')"
TAGS=${TAGS:-}

gocaches="./gocaches"
for ARCH in ${ARCHITECTURES}; do
  cacheDir="$gocaches/$ARCH"
  mkdir -p "$cacheDir"
 
  # try to get a gocache for this arch; this can "fail" but still exit with 0
  echodate "Attempting to fetch gocache for ${ARCH}..."
  TARGET_DIRECTORY="$cacheDir" GOARCH="${ARCH}" ./hack/ci/download-gocache.sh
done

echodate "Building ${IMAGE}:${PRIMARY_TAG}"

# build multi-arch images
docker buildx rm k8c-k1-release || true
docker buildx create --use --name=k8c-k1-release

for ARCH in ${ARCHITECTURES}; do
  echodate "Building a KubeOne image for ${ARCH}..."

  docker buildx build \
    --load \
    --progress=plain \
    --platform="linux/${ARCH}" \
    --build-arg="GOPROXY=${GOPROXY:-}" \
    --build-arg="GOCACHE=/go/src/k8c.io/kubeone/gocaches/${ARCH}" \
    --file="Dockerfile" \
    --tag "${IMAGE}:${PRIMARY_TAG}-${ARCH}" .
done

if [ "$NOMOCK" = true ]; then
  for ARCH in ${ARCHITECTURES}; do
    echodate "Pushing ${IMAGE}:${PRIMARY_TAG}-${ARCH}..."
    docker push "${IMAGE}:${PRIMARY_TAG}-${ARCH}"
  done

  docker manifest create --amend "${IMAGE}:${PRIMARY_TAG}" $(echo "${ARCHITECTURES}" | sed -e "s~[^ ]*~${IMAGE}:${PRIMARY_TAG}\-&~g")
  for ARCH in ${ARCHITECTURES}; do docker manifest annotate --arch "${ARCH}" "${IMAGE}:${PRIMARY_TAG}" "${IMAGE}:${PRIMARY_TAG}-${ARCH}"; done
  docker manifest push --purge "${IMAGE}:${PRIMARY_TAG}"

  for TAG in ${TAGS}; do
    docker manifest create --amend "${IMAGE}:${TAG}" $(echo "${ARCHITECTURES}" | sed -e "s~[^ ]*~${IMAGE}:${PRIMARY_TAG}\-&~g")
    for ARCH in ${ARCHITECTURES}; do docker manifest annotate --arch "${ARCH}" "${IMAGE}:${TAG}" "${IMAGE}:${PRIMARY_TAG}-${ARCH}"; done
    docker manifest push --purge "${IMAGE}:${TAG}"
  done
fi

echodate "Done. :-)"
