#!/usr/bin/env bash

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

# This script is mostly used in CI
# It installs dependencies and starts the tests

set -o errexit
set -o nounset
set -o monitor
set -o pipefail

RUNNING_IN_CI=${JOB_NAME:-""}
BUILD_ID=${BUILD_ID:-"${USER}-local"}
PROVIDER=${PROVIDER:-"PROVIDER-MISSING"}
CONTAINER_RUNTIME=${CONTAINER_RUNTIME:-}
KUBETESTS_ROOT=$(realpath "${KUBETESTS_ROOT:-"/opt/kube-test"}")
KUBEONE_TEST_RUN=${KUBEONE_TEST_RUN:-}
TEST_CONFIG_API_VERSION=${TEST_CONFIG_API_VERSION:-"v1beta2"}
TEST_CLUSTER_TARGET_VERSION=${TEST_CLUSTER_TARGET_VERSION:-}
TEST_CLUSTER_INITIAL_VERSION=${TEST_CLUSTER_INITIAL_VERSION:-}
TEST_OS_CONTROL_PLANE=${TEST_OS_CONTROL_PLANE:-}
TEST_OS_WORKERS=${TEST_OS_WORKERS:-}
TEST_TIMEOUT=${TEST_TIMEOUT:-"60m"}
PATH=$PATH:$(go env GOPATH)/bin
TERRAFORM_DIR=$PWD/examples/terraform
SSH_PRIVATE_KEY_FILE="${HOME}/.ssh/id_rsa_kubeone_e2e"

export TF_IN_AUTOMATION=true
export TF_CLI_ARGS="-no-color"
export TF_VAR_cluster_name=k1-${BUILD_ID}
export TF_VAR_subnets_cidr=27
export SSH_PUBLIC_KEY_FILE="${SSH_PRIVATE_KEY_FILE}.pub"
export TF_VAR_ssh_public_key_file=${SSH_PUBLIC_KEY_FILE}
CREDENTIALS_FILE_PATH=""

function cleanup() {
  set +e
  for try in {1..3}; do
    cd "${TERRAFORM_DIR}/${PROVIDER}"
    echo "Cleaning up terraform state, attempt ${try}"
    # Upstream interpolation bug, but we dont care about the output
    # at destroy time anyways: https://github.com/hashicorp/terraform/issues/17691
    terraform init --backend-config=key="${BUILD_ID}"
    if terraform destroy -auto-approve; then break; fi
    echo "Sleeping for $try seconds"
    sleep "${try}"s
  done
}
trap cleanup EXIT

function fail() {
  echo "$1"
  exit 1
}

function link_s3_backend() {
  # set up terraform remote backend configuration
  for dir in "${TERRAFORM_DIR}"/*; do
    ln -s "${PWD}"/test/e2e/testdata/s3_backend.tf "${dir}"/s3_backend.tf
  done
}

function setup_ci_environment_vars() {
  # If the following variable is set then this script is running in CI
  # and the assumption is that the image contains kubernetes binaries
  #
  # note:
  # kubetest assumes that the last part of that path contains "kubernetes", if not then it complains,
  # additionally the version must be in a very specific format.
  case ${PROVIDER} in
  "aws")
    export AWS_ACCESS_KEY_ID=${AWS_E2E_TESTS_KEY_ID}
    export AWS_SECRET_ACCESS_KEY=${AWS_E2E_TESTS_SECRET}
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
    ;;
  "openstack")
    export OS_AUTH_URL=${OS_AUTH_URL}
    export OS_DOMAIN_NAME=${OS_DOMAIN}
    export OS_REGION_NAME=${OS_REGION}
    export OS_TENANT_NAME=${OS_TENANT_NAME}
    export OS_USERNAME=${OS_USERNAME}
    export OS_PASSWORD=${OS_PASSWORD}
    export TF_VAR_external_network_name="ext-net"
    export TF_VAR_subnet_cidr="10.0.42.0/24"
    echo "${OS_K1_CREDENTIALS}" > /tmp/credentials.yaml
    CREDENTIALS_FILE_PATH=/tmp/credentials.yaml
    ;;
  *)
    fail "unknown provider ${PROVIDER}"
    ;;
  esac

  if [ -d "${KUBETESTS_ROOT}" ]; then
    kubeone_build_dir="_build"
    for kubetest_dir in "${KUBETESTS_ROOT}"/*; do
      basekubetest_name=$(basename "${kubetest_dir}")
      kubetest_dst_dir="${kubeone_build_dir}/${basekubetest_name}"
      mkdir -p "${kubetest_dst_dir}"
      ln -fs "${kubetest_dir}"/* "${kubetest_dst_dir}"
    done
  else
    fail "kubetests directory ${KUBETESTS_ROOT} in not found"
  fi
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

function runE2E() {
  local run_filter=$1
  set -x

  # split run_filter by / in case if filter contain nested subtests.
  # nested subtests are valid for `go test -run`, but will not be reflected by `go test -list`.
  local list_filter
  IFS='/' read -ra list_filter <<< "${run_filter}"

  numberOfTestsToRun=$(go test ./test/e2e -list "${list_filter[0]}" | wc -l)
  numberOfTestsToRun=$(("$numberOfTestsToRun" - 1))

  if [[ "$numberOfTestsToRun" == "0" ]]; then
    fail "run_filter '${run_filter}' selects no tests to run"
  fi

  go test \
    ./test/e2e \
    -v \
    -timeout="${TEST_TIMEOUT}" \
    -run="${run_filter}" \
    -args \
    -credentials="${CREDENTIALS_FILE_PATH}" \
    -identifier="${BUILD_ID}" \
    -provider="${PROVIDER}" \
    -container-runtime="${CONTAINER_RUNTIME}" \
    -os-control-plane="${TEST_OS_CONTROL_PLANE}" \
    -os-workers="${TEST_OS_WORKERS}" \
    -target-version="${TEST_CLUSTER_TARGET_VERSION}" \
    -initial-version="${TEST_CLUSTER_INITIAL_VERSION}" \
    -config-api-version="${TEST_CONFIG_API_VERSION}"
}

if [ -n "${RUNNING_IN_CI}" ]; then
  setup_ci_environment_vars
  link_s3_backend
  generate_ssh_key "${SSH_PRIVATE_KEY_FILE}"
  ssh_agent "${SSH_PRIVATE_KEY_FILE}"
fi

runE2E "${KUBEONE_TEST_RUN}"
