#!/usr/bin/env sh

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

# This is a simple installer script for KubeOne.

set -eu

# What OS is used
OS="$(uname | tr '[:upper:]' '[:lower:]')"
# Find out what's the latest version
VERSION="$(curl -s -H "Accept: application/vnd.github.v3+json" https://api.github.com/repos/kubermatic/kubeone/releases | sed -n 's/^\s*"tag_name": "v\([^"]*\)",$/\1/p' | sort -V | tail -1)"
# Download URL for the latest version
URL="https://github.com/kubermatic/kubeone/releases/download/v${VERSION}/kubeone_${VERSION}_${OS}_amd64.zip"

# 'kubeone' will be installed into this dir
DEST=/usr/local/bin

if [ ! -x "$(command -v unzip)" ]; then
  echo "Your system is missing 'unzip'. Please install it and try again."
  exit 2
fi

# Download the latest version for the OS and save it as zip
if curl -LO "$URL"
then
  echo "Copying kubeone binary into $DEST"
  
  if unzip "kubeone_${VERSION}_${OS}_amd64.zip" -d "kubeone_${VERSION}_${OS}_amd64"
  then
    sudo mv "kubeone_${VERSION}_${OS}_amd64/kubeone" "$DEST"
    rm "kubeone_${VERSION}_${OS}_amd64.zip"
    echo "Kubermatic KubeOne has been installed into $DEST/kubeone"
    echo "Terraform example configs, addons, and helper scripts have been downloaded into the ./kubeone_${VERSION}_${OS}_amd64 directory"
    exit 0
  fi
else
  printf "Failed to determine your platform.\n Try downloading from https://github.com/kubermatic/kubeone/releases"
fi

exit 1
