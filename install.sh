#!/bin/sh

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

#
# This is a simple installer script for KubeOne #
#

# What OS is used
OS=$(uname)
# find out what's the latest version
VERSION=$(curl -w '%{url_effective}' -I -L -s -S https://github.com/kubermatic/kubeone/releases/latest -o /dev/null | sed -e 's|.*/v||')
# download URL for the latest version
URL="https://github.com/kubermatic/kubeone/releases/download/v${VERSION}/kubeone_${VERSION}_${OS}_amd64.zip"

# 'kubeone' will be installed into this dir:
DEST=/usr/local/bin

# Download the latest version for the OS and save it as zip

if curl -LO "$URL"
then 
    echo "Copying kubeone binary into $DEST"
    # unpack:
    

    if unzip "kubeone_${VERSION}_${OS}_amd64.zip" -d "kubeone_${VERSION}_${OS}_amd64"
    then
        sudo mv "kubeone_${VERSION}_${OS}_amd64/kubeone" "$DEST"
        rm "kubeone_${VERSION}_${OS}_amd64.zip"
        rm -rf "kubeone_${VERSION}_${OS}_amd64"
        echo "kubeone has been installed into $DEST/kubeone"
        exit 0
    fi
else
    printf "Failed to determine your platform.\n Try downloading from https://github.com/kubermatic/kubeone/releases"
fi

exit 1
