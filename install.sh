#!/bin/sh
#
# This is a simple installer script for KubeOne 
#
#
# What OS is used
OS=$(uname)
# find out what's the latest version
VERSION=$(curl -w '%{url_effective}' -I -L -s -S https://github.com/kubermatic/kubeone/releases/latest -o /dev/null | sed -e 's|.*/v||')
# download URL for the latest version
URL="https://github.com/kubermatic/kubeone/releases/download/v${VERSION}/kubeone_${VERSION}_${OS}_amd64.zip"
echo $OS
echo $VERSION
echo $URL

# 'kubeone' will be installed into this dir:
DEST=/usr/local/bin

# Download the latest version for the OS and save it as zip
curl -LO $URL

if [ $? -eq 0 ]
then 
    echo "Copying kubeone binary into $DEST"
    # unpack:
    unzip kubeone_${VERSION}_${OS}_amd64.zip -d kubeone_${VERSION}_${OS}amd64

    if [ $? -eq 0 ]
    then
        sudo mv kubeone_${VERSION}_${OS}amd64/kubeone $DEST
        rm kubeone_${VERSION}_${OS}_amd64.zip
        rm -rf kubeone_${VERSION}_${OS}amd64
        echo "kubeone has been installed into $DEST/kubeone"
        exit 0
    fi
else
    echo "Failed to determine your platform.\n Try downloading from https://github.com/kubermatic/kubeone/releases"
fi

exit 1
