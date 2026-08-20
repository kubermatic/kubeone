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

# This script is mostly used in CI
# It installs dependencies and starts the tests

set -o errexit
set -o nounset
set -o monitor
set -o pipefail

PROVIDER=${PROVIDER:-"NONE"}
RUNNING_IN_CI=${JOB_NAME:-""}
BUILD_ID=${BUILD_ID:-"${USER}-${RANDOM}"}
BUILD_DIR=$(mktemp -d "${BUILD_ID}-XXX" -p /tmp)
TEST_TIMEOUT=${TEST_TIMEOUT:-"120m"}
SSH_PRIVATE_KEY_FILE=${SSH_PRIVATE_KEY_FILE:-"${BUILD_DIR}/ssh_key_kubeone_e2e"}
PATH=$PATH:$(go env GOPATH)/bin
SSH_PUBLIC_KEY_FILE="${SSH_PRIVATE_KEY_FILE}.pub"
CREDENTIALS_FILE_PATH=""
TERRAFORM_VERSION=${TERRAFORM_VERSION:-"1.13.3"}
SONOBUOY_VERSION=${SONOBUOY_VERSION:-"0.57.3"}
PROTOKOL_VERSION=${PROTOKOL_VERSION:-"0.7.5"}
TOOLS_CACHE_DIR=${TOOLS_CACHE_DIR:-"${HOME}/.cache/kubeone-e2e/tools"}
OS_NAME=$(uname -s | tr '[:upper:]' '[:lower:]')

case $(uname -m) in
x86_64) ARCH_NAME=amd64 ;;
aarch64 | arm64) ARCH_NAME=arm64 ;;
*) ARCH_NAME=$(uname -m) ;;
esac

# tool name -> version, used to keep ensure_tool() cache dirs version-scoped
declare -A TOOL_VERSIONS=(
  [terraform]="${TERRAFORM_VERSION}"
  [sonobuoy]="${SONOBUOY_VERSION}"
  [protokol]="${PROTOKOL_VERSION}"
)

# tool name -> release download URL, used by ensure_tool()
declare -A TOOL_DOWNLOAD_URLS=(
  [terraform]="https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${OS_NAME}_${ARCH_NAME}.zip"
  [sonobuoy]="https://github.com/vmware-tanzu/sonobuoy/releases/download/v${SONOBUOY_VERSION}/sonobuoy_${SONOBUOY_VERSION}_${OS_NAME}_${ARCH_NAME}.tar.gz"
  [protokol]="https://codeberg.org/xrstf/protokol/releases/download/v${PROTOKOL_VERSION}/protokol_${PROTOKOL_VERSION}_${OS_NAME}_${ARCH_NAME}.tar.gz"
)

export PATH
export TF_VAR_cluster_name=k1-${BUILD_ID}
export TF_VAR_ssh_public_key_file=${SSH_PUBLIC_KEY_FILE}

function cleanup() {
  test -d "$BUILD_DIR" && rm -rf "$BUILD_DIR"
  ssh-agent -k || true
}
trap cleanup EXIT

function fatal() {
  echo "$1"
  exit 1
}

# ensure_tool downloads $1 into a versioned cache dir and prepends it to PATH unless already on PATH
function ensure_tool() {
  local tool=$1

  if command -v "${tool}" > /dev/null 2>&1; then
    return
  fi

  local url=${TOOL_DOWNLOAD_URLS[${tool}]:-}
  if [ -z "${url}" ]; then
    fatal "no download URL configured for ${tool}"
  fi

  local archive_name
  archive_name=$(basename "${url}")
  local install_dir="${TOOLS_CACHE_DIR}/${tool}/${TOOL_VERSIONS[${tool}]}"

  if [ ! -x "${install_dir}/${tool}" ]; then
    echo "Downloading ${tool} from ${url}"
    mkdir -p "${install_dir}"

    local archive_path="${BUILD_DIR}/${archive_name}"
    curl --fail --location --output "${archive_path}" "${url}"

    # extract to a scratch dir first, some archives nest the binary in a subdirectory
    local extract_dir="${BUILD_DIR}/${tool}-extracted"
    mkdir -p "${extract_dir}"
    case "${archive_name}" in
    *.zip) unzip -o -q "${archive_path}" -d "${extract_dir}" ;;
    *.tar.gz) tar -xzf "${archive_path}" -C "${extract_dir}" ;;
    *) fatal "unsupported archive format for ${tool}: ${archive_name}" ;;
    esac

    local extracted_bin
    extracted_bin=$(find "${extract_dir}" -type f -name "${tool}" | head -n 1)
    if [ -z "${extracted_bin}" ]; then
      fatal "could not find ${tool} binary inside ${archive_name}"
    fi

    mv "${extracted_bin}" "${install_dir}/${tool}"
    rm -rf "${extract_dir}"
    chmod +x "${install_dir}/${tool}"
  fi

  PATH="${install_dir}:${PATH}"
  export PATH
}

function generate_ssh_key() {
  local private_ssh_key_file=$1

  if [ ! -f "${private_ssh_key_file}" ]; then
    echo "Generating SSH key pair"
    ssh-keygen -f "${private_ssh_key_file}" -N ''
    chmod 400 "${private_ssh_key_file}"
  fi
}

function ssh_agent() {
  local private_ssh_key_file=$1

  ssh-agent -k || true
  eval "$(ssh-agent)"
  ssh-add "${private_ssh_key_file}"
}

function setup_ci_environment_vars() {
  # If the following variable is set then this script is running in CI
  # and the assumption is that the image contains kubernetes binaries
  case ${PROVIDER} in
  "aws")
    export AWS_ACCESS_KEY_ID=${AWS_E2E_TESTS_KEY_ID}
    export AWS_SECRET_ACCESS_KEY=${AWS_E2E_TESTS_SECRET}
    ;;
  "azure")
    export ARM_CLIENT_ID=${AZURE_E2E_TESTS_CLIENT_ID}
    export ARM_CLIENT_SECRET=${AZURE_E2E_TESTS_CLIENT_SECRET}
    export ARM_SUBSCRIPTION_ID=${AZURE_E2E_TESTS_SUBSCRIPTION_ID}
    export ARM_TENANT_ID=${AZURE_E2E_TESTS_TENANT_ID}
    export TF_VAR_rhsm_username=${RHEL_SUBSCRIPTION_MANAGER_USER:-""}
    export TF_VAR_rhsm_password=${RHEL_SUBSCRIPTION_MANAGER_PASSWORD:-""}
    export TF_VAR_rhsm_offline_token=${REDHAT_SUBSCRIPTIONS_OFFLINE_TOKEN:-""}
    CREDENTIALS_FILE_PATH="${BUILD_DIR}/credentials.yaml"

    cat > "${CREDENTIALS_FILE_PATH}" << EOL
cloudConfig: |
  {
    "aadClientId": "${ARM_CLIENT_ID}",
    "aadClientSecret": "${ARM_CLIENT_SECRET}",
    "subscriptionId": "${ARM_SUBSCRIPTION_ID}",
    "tenantId": "${ARM_TENANT_ID}",
    "resourceGroup": "${TF_VAR_cluster_name}-rg",
    "location": "westeurope",
    "subnetName": "${TF_VAR_cluster_name}-subnet",
    "routeTableName": "",
    "loadBalancerSku": "Standard",
    "securityGroupName": "${TF_VAR_cluster_name}-sg",
    "vnetName": "${TF_VAR_cluster_name}-vpc",
    "primaryAvailabilitySetName": "${TF_VAR_cluster_name}-avset",
    "useInstanceMetadata": true,
    "useManagedIdentityExtension": false,
    "userAssignedIdentityID": "",
    "vmType": "standard"
  }
EOL
    ;;
  "digitalocean")
    export DIGITALOCEAN_TOKEN=${DO_E2E_TESTS_TOKEN}
    ;;
  "hetzner")
    export HCLOUD_TOKEN=${HZ_E2E_TOKEN}
    ;;
  "equinixmetal")
    export TF_VAR_project_id=${METAL_PROJECT_ID}
    ;;
  "gce")
    GOOGLE_CREDENTIALS=$(base64 -d <<< "${KUBEONE_GOOGLE_SERVICE_ACCOUNT}")
    export GOOGLE_CREDENTIALS
    export TF_VAR_project="kubeone-terraform-test"

    CREDENTIALS_FILE_PATH="${BUILD_DIR}/credentials.yaml"
    cat > "${CREDENTIALS_FILE_PATH}" << EOL
cloudConfig: |
  [global]
  regional = true
  multi-zone = true
  token-url = "nil"
EOL
    ;;
  "openstack")
    export OS_AUTH_URL=${OS_AUTH_URL}
    export OS_DOMAIN_NAME=${OS_DOMAIN}
    export OS_REGION_NAME=${OS_REGION}
    export OS_TENANT_NAME=${OS_TENANT_NAME}
    export OS_USERNAME=${OS_USERNAME}
    export OS_PASSWORD=${OS_PASSWORD}
    CREDENTIALS_FILE_PATH="${BUILD_DIR}/credentials.yaml"
    echo "${OS_K1_CREDENTIALS}" > "${CREDENTIALS_FILE_PATH}"
    ;;
  "vsphere")
    export VSPHERE_ALLOW_UNVERIFIED_SSL=true
    export VSPHERE_SERVER="${VSPHERE_E2E_ADDRESS/http*:\/\//}"
    export VSPHERE_USER=${VSPHERE_E2E_USERNAME}
    export VSPHERE_PASSWORD=${VSPHERE_E2E_PASSWORD}
    CREDENTIALS_FILE_PATH="${BUILD_DIR}/credentials.yaml"

    cat > "${CREDENTIALS_FILE_PATH}" << EOL
cloudConfig: |
  [Global]
  secret-name = "vsphere-ccm-credentials"
  secret-namespace = "kube-system"
  port = "443"
  insecure-flag = "1"

  [VirtualCenter "${VSPHERE_SERVER}"]

  [Workspace]
  server = "${VSPHERE_SERVER}"
  datacenter = "Hamburg"
  default-datastore="Datastore0-truenas"
  resourcepool-path=""
  folder = "KubeOne-E2E"

  [Disk]
  scsicontrollertype = pvscsi

  [Network]
  public-network = "Default Network"
csiConfig: |
  [Global]
  cluster-id = "k1-${BUILD_ID}"
  user = "${VSPHERE_USER}"
  password = "${VSPHERE_PASSWORD}"
  port = "443"
  insecure-flag = "1"

  [VirtualCenter "${VSPHERE_SERVER}"]

  [Workspace]
  server = "${VSPHERE_SERVER}"
  datacenter = "Hamburg"
  default-datastore="Datastore0-truenas"
  resourcepool-path=""
  folder = "KubeOne-E2E"
EOL
    ;;
  *)
    echo "unknown provider ${PROVIDER}"
    exit 1
    ;;
  esac
}

if ! command -v unzip > /dev/null 2>&1; then
  # assume debian like
  apt update && apt install unzip -y
fi

ensure_tool "terraform"
ensure_tool "sonobuoy"
ensure_tool "protokol"

generate_ssh_key "${SSH_PRIVATE_KEY_FILE}"
ssh_agent "${SSH_PRIVATE_KEY_FILE}"

if [ -n "${RUNNING_IN_CI}" ]; then
  setup_ci_environment_vars
fi

go_test_args=("$@")
TEST_NAME="$*"

if [ -n "${CREDENTIALS_FILE_PATH}" ]; then
  go_test_args+=("-credentials" "${CREDENTIALS_FILE_PATH}")
fi

cd test/e2e

go test -c . -tags e2e

./e2e.test -test.list "$TEST_NAME" | grep -q "$TEST_NAME" || fatal "NO TESTS MATCH $TEST_NAME"

# to handle OS signals directly, we launch e2e tests using dedicated binary
exec ./e2e.test \
  -test.timeout "$TEST_TIMEOUT" \
  -test.v \
  -test.run \
  "${go_test_args[@]}"
