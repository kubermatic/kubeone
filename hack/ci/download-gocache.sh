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

### This script is used when building Go binaries to download a prewarmed
### cache for `$GOCACHE`. This is an uncompressed tar archive in a local
### S3-compatible storage, which is just downloaded and extracted. This
### significantly speeds up CI jobs, as much fewer packages have to be
### compiled.
###
### A dedicated postsubmit job that runs after every change to the master
### branch is keeping the cache up-to-date.

set -euo pipefail

# Required for signal propagation to work so
# the cleanup trap gets executed when the script
# receives a SIGINT
set -o monitor

# The gocache needs a matching go version to work, so append that to the name
GO_VERSION="$(go version | awk '{ print $3 }' | sed 's/go//g')"
GOARCH="$(go env GOARCH)"

# Make sure we never error, this is always best-effort only
exit_gracefully() {
  if [ $? -ne 0 ]; then
    echodate "Encountered error when trying to download gocache"
  fi
  exit 0
}
trap exit_gracefully EXIT

source $(dirname $0)/../lib.sh

if [ -z "${GOCACHE_MINIO_ADDRESS:-}" ]; then
  echodate "env var GOCACHE_MINIO_ADDRESS unset, can not download gocache"
  exit 0
fi

GOCACHE="$(go env GOCACHE)"
# Make sure it actually exists
mkdir -p "${GOCACHE}"

# PULL_BASE_REF is the name of the current branch in case of a post-submit
# or the name of the base branch in case of a PR.
GIT_BRANCH="${PULL_BASE_REF:-}"
CACHE_VERSION="${PULL_BASE_SHA:-}"

# Periodics just use their head ref
if [[ -z "${CACHE_VERSION}" ]]; then
  CACHE_VERSION="$(git rev-parse HEAD)"
  GIT_BRANCH="master"
fi

if [ -z "${PULL_NUMBER:-}" ]; then
  # Special case: This is called in a Postsubmit. Go one revision back,
  # as there can't be a cache for the current revision
  CACHE_VERSION="$(git rev-parse ${CACHE_VERSION}~1)"
fi

# normalize branch name to prevent accidental directories being created
GIT_BRANCH="$(echo "$GIT_BRANCH" | sed 's#/#-#g')"

ARCHIVE_NAME="${CACHE_VERSION}-${GO_VERSION}-${GOARCH}.tar"
URL="${GOCACHE_MINIO_ADDRESS}/kubeone/${GIT_BRANCH}/${ARCHIVE_NAME}"

# Do not go through the retry loop when there is nothing, but do try the
# first parent if no cache was found. This is helpful for retests happening
# quickly after something got merged to master and no gocache for the most
# recent commit exists yet. In this case, taking the previous commit's
# cache is better than nothing.
if ! curl --head --silent --fail "${URL}" > /dev/null; then
  echodate "Remote has no gocache ${ARCHIVE_NAME}, trying previous commit as a fallback..."

  CACHE_VERSION="$(git rev-parse ${CACHE_VERSION}~1)"
  ARCHIVE_NAME="${CACHE_VERSION}-${GO_VERSION}-${GOARCH}.tar"
  URL="${GOCACHE_MINIO_ADDRESS}/kubeone/${GIT_BRANCH}/${ARCHIVE_NAME}"

  if ! curl --head --silent --fail "${URL}" > /dev/null; then
    echodate "Remote has no gocache ${ARCHIVE_NAME}, giving up."
    exit 0
  fi
fi

echodate "Downloading and extracting gocache"
TEST_NAME="Download and extract gocache"
# Passing the Headers as space-separated literals doesn't seem to work
# in conjunction with the retry func, so we just put them in a file instead
echo 'Content-Type: application/octet-stream' > /tmp/headers
retry 5 curl --fail -H @/tmp/headers "${URL}" | tar -C $GOCACHE -xf -

echodate "Successfully fetched gocache into $GOCACHE"
