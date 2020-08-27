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

export GOCACHE="$(mktemp -d)"
export CGO_ENABLED=0
export GO111MODULE=on
export GOFLAGS=-mod=readonly -trimpath

make build
go test -run thisTestDoesNotExist ./pkg/... ./test/...
go test -run thisTestDoesNotExist -tags e2e ./pkg/... ./test/...

GIT_HEAD_HASH="$(git rev-parse HEAD | tr -d '\n')"
GO_VERSION="$(go version | awk '{ print $3 }' | sed 's/go//g')"
ARCHIVE_FILE="/tmp/${GIT_HEAD_HASH}.tar"

# No compression because that needs quite a bit of CPU
tar -C "${GOCACHE}" -cf "${ARCHIVE_FILE}" .

curl --fail -T "${ARCHIVE_FILE}" -H 'Content-Type: application/octet-stream' "${GOCACHE_MINIO_ADDRESS}/kubeone-${GIT_HEAD_HASH}-${GO_VERSION}.tar"
