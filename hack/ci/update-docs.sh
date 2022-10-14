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

### Updates the docs repository by copying over the API references.

set -euo pipefail

cd $(dirname $0)/../..
source hack/lib.sh

TARGET_DIR=docs_sync
GIT_REVISION=$(git rev-parse --short HEAD)
GIT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
KUBEONE_VERSION=$(basename ${GIT_BRANCH})

echodate "Updating KubeOne API references for ${KUBEONE_VERSION} (${GIT_REVISION})"

# configure Git
git config --global user.email "dev@kubermatic.com"
git config --global user.name "Prow CI Robot"
git config --global core.sshCommand 'ssh -o CheckHostIP=no -i /ssh/id_rsa'
ensure_github_host_pubkey

# create a fresh clone
git clone git@github.com:kubermatic/docs.git $TARGET_DIR
cd $TARGET_DIR

find ../docs/api_reference -name '*.en.md' -print0 | while IFS= read -r -d '' docsPath; do 
  # Convert ../docs/api_reference/v1beta2.en.md -> v1beta2
  apiVersion=$(basename ${docsPath} | awk -F. '{print $1}')
  echodate "Copying ${apiVersion} docs..."

  mkdir -p content/kubeone/"${KUBEONE_VERSION}"/references/kubeone-cluster-"${apiVersion}"
  cp ../docs/api_reference/"${apiVersion}".en.md \
    content/kubeone/"${KUBEONE_VERSION}"/references/kubeone-cluster-"${apiVersion}"/_index.en.md
done

# update repo
git add .

if ! git diff --cached --stat --exit-code; then
  echodate "Pushing changes to the docs repo..."
  git commit -m "Syncing with kubermatic/kubeone@${GIT_REVISION}"
  git push
fi

echodate "Done. :-)"
