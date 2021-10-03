#!/usr/bin/env bash

# Copyright 2021 The KubeOne Authors.
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

### Runs as a postsubmit and refreshes the gocache by downloading the
### previous version, compiling everything and then tar'ing up the
### Go cache again.

set -euo pipefail

# Required for signal propagation to work so
# the cleanup trap gets executed when the script
# receives a SIGINT
set -o monitor

cd $(dirname $0)/../..
source hack/lib.sh

if [ -z "${GOCACHE_MINIO_ADDRESS:-}" ]; then
  echodate "Fatal: env var GOCACHE_MINIO_ADDRESS unset"
  exit 1
fi

# The gocache needs a matching go version to work, so append that to the name
GO_VERSION="$(go version | awk '{ print $3 }' | sed 's/go//g')"
GOARCH="$(go env GOARCH)"

GOCACHE_DIR="$(mktemp -d)"
export GOCACHE="${GOCACHE_DIR}"
export GIT_HEAD_HASH="$(git rev-parse HEAD | tr -d '\n')"
export CGO_ENABLED=0

# PULL_BASE_REF is the name of the current branch in case of a post-submit
# or the name of the base branch in case of a PR.
GIT_BRANCH="${PULL_BASE_REF:-}"

# normalize branch name to prevent accidental directories being created
GIT_BRANCH="$(echo "$GIT_BRANCH" | sed 's#/#-#g')"

echodate "Creating cache for revision ${GIT_BRANCH}/${GIT_HEAD_HASH} / Go ${GO_VERSION}/${GOARCH} ..."

echodate "Building binaries"

(
  # prevent the Makefile from downloading the old Gocache. This ensures that
  # our cache does not grow over time, as packages are added and removed,
  # but makes creating the cache a tiny bit slower
  touch download-gocache

  make build
)

echodate "Building tests"

(
  go test -run thisTestDoesNotExist ./pkg/... ./test/...
  go test -run thisTestDoesNotExist -tags e2e ./pkg/... ./test/...
)

TEST_NAME="Creating gocache archive"
echodate "Creating gocache archive"

ARCHIVE_FILE="/tmp/${GIT_HEAD_HASH}.tar"
# No compression because that needs quite a bit of CPU
retry 2 tar -C "$GOCACHE" -cf "$ARCHIVE_FILE" .

TEST_NAME="Uploading gocache archive"
echodate "Uploading gocache archive"

# Passing the Headers as space-separated literals doesn't seem to work
# in conjunction with the retry func, so we just put them in a file instead
echo 'Content-Type: application/octet-stream' > /tmp/headers
retry 2 curl --fail -T "${ARCHIVE_FILE}" -H @/tmp/headers "${GOCACHE_MINIO_ADDRESS}/kubeone/${GIT_BRANCH}/${GIT_HEAD_HASH}-${GO_VERSION}-${GOARCH}.tar"

echodate "Upload complete."
