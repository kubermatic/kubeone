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

DOCKER_REPO="${DOCKER_REPO:-quay.io/kubermatic}"
ARCHITECTURES=${ARCHITECTURES:-amd64 arm64}
NOMOCK=${NOMOCK:-false}

PRIMARY_TAG="$(git rev-parse HEAD | tr -d '\n')"
TAGS=${TAGS:-}

gocaches="$(mktemp -d)"
for ARCH in ${ARCHITECTURES}; do
  cacheDir="$gocaches/$ARCH"
  mkdir -p "$cacheDir"
 
  # try to get a gocache for this arch; this can "fail" but still exit with 0
  echodate "Attempting to fetch gocache for $ARCH..."
  TARGET_DIRECTORY="$cacheDir" GOARCH="$ARCH" ./hack/ci/download-gocache.sh
done

echodate "Building ${DOCKER_REPO}/kubeone:${PRIMARY_TAG}"

# build multi-arch images
buildah manifest create "${DOCKER_REPO}/kubeone:${PRIMARY_TAG}"
for ARCH in ${ARCHITECTURES}; do
  echodate "Building a KubeOne image for $ARCH..."

  # Building via buildah does not use the gocache, but that's okay, because we
  # wouldn't want to cache arm64 stuff anyway, as it would just blow up the
  # cache size and force every e2e test to download gigabytes worth of unneeded
  # arm64 stuff. We might need to change this once we run e2e tests on arm64.
  buildah bud \
    --tag="${DOCKER_REPO}/kubeone-${ARCH}:${PRIMARY_TAG}" \
    --build-arg="GOPROXY=${GOPROXY:-}" \
    --build-arg="GOCACHE=/gocache" \
    --arch="$ARCH" \
    --override-arch="$ARCH" \
    --format=docker \
    --file Dockerfile \
    --volume "$gocaches/$ARCH:/gocache" \
    .
  buildah manifest add "${DOCKER_REPO}/kubeone:${PRIMARY_TAG}" "${DOCKER_REPO}/kubeone-${ARCH}:${PRIMARY_TAG}"
done

if [ "$NOMOCK" = true ]; then
  echodate "Pushing ${DOCKER_REPO}/kubeone:${PRIMARY_TAG}..."
  buildah manifest push --all "${DOCKER_REPO}/kubeone:${PRIMARY_TAG}" "docker://${DOCKER_REPO}/kubeone:${PRIMARY_TAG}"

  for TAG in ${TAGS}; do
    echodate "Pushing ${DOCKER_REPO}/kubeone:${TAG}..."
    buildah tag "${DOCKER_REPO}/kubeone:${PRIMARY_TAG}" "${DOCKER_REPO}/kubeone:${TAG}"
    buildah manifest push --all "${DOCKER_REPO}/kubeone:${TAG}" "docker://${DOCKER_REPO}/kubeone:${TAG}"
  done
fi

echodate "Done. :-)"
