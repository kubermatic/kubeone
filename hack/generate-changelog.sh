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

### This script is used to generate the changelog. It uses the Kubernetes
### release-notes tool (https://github.com/kubernetes/release) under the hood.
###
### Prerequisites:
###   1) GitHub token set as GITHUB_TOKEN environment variable
###   2) the release-notes tool installed
###
### Currently, there are no release-notes binaries, so you have to build it
### manually. This can be done by cloning https://github.com/kubernetes/release
### and running make generate-tools (note: this will include other
### Kubernetes tools as well, e.g. krel, cip-mm...).
###
### Usage:
###  The script can be used in the following way:
###    CHANGELOG_START_REV="v1.5.2" \
###    CHANGELOG_END_SHA="6c8a662a94ecf78ea98f3ad8cc899465445e7d86" \
###    CHANGELOG_BRANCH="release/v1.5" \
###    ./hack/generate-changelog.sh
###
###  The changelog will be saved to the /tmp directory with the random
###  generated filename. You can a custom path/filename by using the
###  CHANGELOG_OUTPUT variable.
###
###  CHANGELOG_START_REV and CHANGELOG_END_SHA are required variables.
###  There are additional environment variables that can control the changelog
###  generation, but those are optional. You can check the script contents for
###  more details.

set -euo pipefail

# Required for signal propagation to work so
# the cleanup trap gets executed when the script
# receives a SIGINT
set -o monitor

source $(dirname $0)/lib.sh

if [ -z "${GITHUB_TOKEN:-}" ]; then
  echodate "env var GITHUB_TOKEN unset, cannot generate changelog"
  exit 0
fi

if [ -z "${CHANGELOG_START_REV:-}" ]; then
  echodate "env var CHANGELOG_START_REV unset, cannot generate changelog"
  exit 0
fi

if [ -z "${CHANGELOG_END_SHA:-}" ]; then
  echodate "env var CHANGELOG_END_SHA unset, cannot generate changelog"
  exit 0
fi

org=${CHANGELOG_ORG:-"kubermatic"}
repo=${CHANGELOG_REPO:-"kubeone"}
output=${CHANGELOG_OUTPUT:-""}
tpl=${CHANGELOG_TPL:-"$(dirname "$0")/changelog.tpl"}

git_branch="$(git rev-parse --abbrev-ref HEAD)"
branch=${CHANGELOG_BRANCH:-"$git_branch"}

release-notes \
    --org="$org" \
    --repo="$repo" \
    --start-rev="$CHANGELOG_START_REV" \
    --end-sha="$CHANGELOG_END_SHA" \
    --branch="$branch" \
    --go-template="go-template:$tpl" \
    --output="$output" \
    --required-author "" \
    --markdown-links
