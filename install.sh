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

KUBEONE_DIR="kubeone-tmp"
INSTALL_DIR="/usr/local/bin"

# --- 1. Argument Handling and Dependency Check ---

if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    echo "This script downloads and installs the latest KubeOne CLI."
    echo "Usage: curl -sfL https://get.kubeone.io | sh"
    echo "To install a specific version: KUBEONE_VERSION=v1.2.3 curl -sfL https://get.kubeone.io | sh"
    exit 0
fi

# Check for necessary tools
if ! command -v curl >/dev/null 2>&1; then
    echo "Error: curl is required to download KubeOne." >&2
    exit 1
fi
if ! command -v unzip >/dev/null 2>&1; then
    echo "Error: unzip is required to extract KubeOne." >&2
    exit 1
fi

# --- 2. Determine OS and Architecture (MODIFIED FOR ARM64) ---

# Determine OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    linux)
        OS="linux"
        ;;
    darwin)
        OS="darwin"
        ;;
    *)
        echo "Error: unsupported operating system ${OS}" >&2
        exit 1
        ;;
esac

# Determine Architecture - ** ARM64/AARCH64 SUPPORT ADDED HERE **
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Error: unsupported architecture $ARCH" >&2
        exit 1
        ;;
esac

# --- 3. Determine Version ---

if [ -z "$KUBEONE_VERSION" ]; then
    echo "Determining latest KubeOne version..."
    # Uses the redirect URL of the /latest release to find the tag
    VERSION="$(curl -s -H "Accept: application/vnd.github.v3+json" https://api.github.com/repos/kubermatic/kubeone/releases | sed -n -E 's/^[[:space:]]*"tag_name": "v(([0-9])*\.([0-9])*\.([0-9])*)",$/\1/p' | sort -V | tail -1)"
else
    # Remove 'v' prefix if user specified it
    VERSION="${KUBEONE_VERSION#v}"
fi

if [ -z "$VERSION" ]; then
    echo "Error: Could not determine KubeOne version." >&2
    exit 1
fi

echo "KubeOne Version: v${VERSION} (${OS}/${ARCH})"

# --- 4. Download and Install ---

FILENAME="kubeone_${VERSION}_${OS}_${ARCH}.zip"
DOWNLOAD_URL="https://github.com/kubermatic/kubeone/releases/download/v${VERSION}/${FILENAME}"
TEMP_ARCHIVE="/tmp/${FILENAME}"

echo "Downloading from: ${DOWNLOAD_URL}"
curl -sfL "$DOWNLOAD_URL" -o "$TEMP_ARCHIVE"

if [ $? -ne 0 ]; then
    echo "Error: Download failed. Check if v${VERSION} release exists for ${OS}/${ARCH}." >&2
    exit 1
fi

# Create a temporary directory and unzip
mkdir -p "$KUBEONE_DIR"
unzip -q "$TEMP_ARCHIVE" -d "$KUBEONE_DIR"

echo "Installing KubeOne to ${INSTALL_DIR}..."
# The zip archive typically contains the binary directly in the root of the folder.
sudo mv "${KUBEONE_DIR}/kubeone" "$INSTALL_DIR/kubeone"
sudo chmod +x "$INSTALL_DIR/kubeone"

# --- 5. Cleanup and Finish ---

echo "Cleaning up..."
rm -rf "$KUBEONE_DIR"
rm "$TEMP_ARCHIVE"

echo ""
echo "Successfully installed KubeOne v${VERSION}!"
echo "Run 'kubeone version' to verify."

# Check if INSTALL_DIR is in PATH
if echo "$PATH" | grep -q "$INSTALL_DIR"; then
    : # Do nothing, it's in PATH
else
    echo "Warning: ${INSTALL_DIR} is not in your PATH. You may need to add it manually to run 'kubeone'." >&2
fi