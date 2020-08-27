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

set -euo pipefail

cd $(dirname $0)/..

if [ -z "${GOCACHE_MINIO_ADDRESS:-}" ]; then
  echo "Fatal: env var GOCACHE_MINIO_ADDRESS unset"
  exit 1
fi

GO_VERSION="$(go version | awk '{ print $3 }' | sed 's/go//g')"
GOCACHE="$(go env GOCACHE)"
CACHE_VERSION="${PULL_BASE_SHA:-}"

# Prow periodics just use their head ref
if [ -z "${CACHE_VERSION}" ]; then
  CACHE_VERSION="$(git rev-parse HEAD)"
fi

if [ -z "${PULL_NUMBER:-}" ]; then
  # This is a postubmit job: go one revision back,
  # as there can't be a cache for the current revision
  CACHE_VERSION="$(git rev-parse ${CACHE_VERSION}~1)"
fi

ARCHIVE_NAME="kubeone-${CACHE_VERSION}-${GO_VERSION}.tar"
URL="${GOCACHE_MINIO_ADDRESS}/${ARCHIVE_NAME}"

if curl --head -s "${URL}" | grep -q 404; then
  echo "Note: remote has no gocache ${ARCHIVE_NAME}"
  exit 0
fi

echo "Download gocache"
mkdir -p "${GOCACHE}"
curl --fail -H 'Content-Type: application/octet-stream' -H @/tmp/headers "${URL}" | tar -C "${GOCACHE}" -xf -
