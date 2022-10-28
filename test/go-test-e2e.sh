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
TEST_TIMEOUT=${TEST_TIMEOUT:-"60m"}
SSH_PRIVATE_KEY_FILE=${SSH_PRIVATE_KEY_FILE:-"${BUILD_DIR}/ssh_key_kubeone_e2e"}
PATH=$PATH:$(go env GOPATH)/bin
SSH_PUBLIC_KEY_FILE="${SSH_PRIVATE_KEY_FILE}.pub"
CREDENTIALS_FILE_PATH=""

export PATH
export TF_VAR_cluster_name=k1-${BUILD_ID}
export TF_VAR_ssh_public_key_file=${SSH_PUBLIC_KEY_FILE}

function cleanup() {
  test -d "$BUILD_DIR" && rm -rf "$BUILD_DIR"
  ssh-agent -k || true
}
trap cleanup EXIT

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
    "securityGroupName": "${TF_VAR_cluster_name}-sg",
    "vnetName": "${TF_VAR_cluster_name}-vpc",
    "primaryAvailabilitySetName": "${TF_VAR_cluster_name}-avset-workers",
    "useInstanceMetadata": true,
    "useManagedIdentityExtension": false,
    "userAssignedIdentityID": ""
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
  default-datastore="alpha1"
  resourcepool-path=""
  folder = "kubeone-e2e"

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
  default-datastore="alpha1"
  resourcepool-path=""
  folder = "kubeone-e2e"
EOL
    ;;
  *)
    echo "unknown provider ${PROVIDER}"
    exit 1
    ;;
  esac
}

generate_ssh_key "${SSH_PRIVATE_KEY_FILE}"
ssh_agent "${SSH_PRIVATE_KEY_FILE}"

if [ -n "${RUNNING_IN_CI}" ]; then
  setup_ci_environment_vars
fi

go_test_args=("$@")

if [ -n "${CREDENTIALS_FILE_PATH}" ]; then
  go_test_args+=("-credentials" "${CREDENTIALS_FILE_PATH}")
fi

cd test/e2e

go test -c . -tags e2e

# to handle OS signals directly, we launch e2e tests using dedicated binary
exec ./e2e.test \
  -test.timeout "$TEST_TIMEOUT" \
  -test.run \
  "${go_test_args[@]}"
