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

set -euo pipefail
# Required for signal propagation to work so the cleanup trap
# gets executed when the script receives a SIGINT
set -o monitor

RUNNING_IN_CI=${JOB_NAME:-""}
BUILD_ID=${BUILD_ID:-"${USER}-local"}
PROVIDER=${PROVIDER:-"aws"}
KUBETESTS_ROOT=$(realpath ${KUBETESTS_ROOT:-"/opt/kube-test"})
TEST_SET=${TEST_SET:-"conformance"}
TEST_CLUSTER_TARGET_VERSION=${TEST_CLUSTER_TARGET_VERSION:-""}
TEST_CLUSTER_INITIAL_VERSION=${TEST_CLUSTER_INITIAL_VERSION:-""}
TEST_OS_CONTROL_PLANE=${TEST_OS_CONTROL_PLANE:-""}
TEST_OS_WORKERS=${TEST_OS_WORKERS:-""}
export TF_VAR_cluster_name=k1-${BUILD_ID}

PATH=$PATH:$(go env GOPATH)/bin

TERRAFORM_DIR=$PWD/examples/terraform
function cleanup() {
  set +e
  for try in {1..20}; do
    cd ${TERRAFORM_DIR}/${PROVIDER}
    echo "Cleaning up terraform state, attempt ${try}"
    # Upstream interpolation bug, but we dont care about the output
    # at destroy time anyways: https://github.com/hashicorp/terraform/issues/17691
    terraform init --backend-config=key=${BUILD_ID}
    terraform destroy -auto-approve
    if [[ $? == 0 ]]; then break; fi
    echo "Sleeping for $try seconds"
    sleep ${try}s
  done
}
trap cleanup EXIT

function fail() {
  echo $1
  exit 1
}

# If the following variable is set then this script is running in CI
# and the assumption is that the image contains kubernetes binaries
#
# note:
# kubetest assumes that the last part of that path contains "kubernetes", if not then it complains,
# additionally the version must be in a very specific format.
if [ -n "${RUNNING_IN_CI}" ]; then
  # set up terraform remote backend configuration
  for dir in ${TERRAFORM_DIR}/*; do
    ln -s $PWD/test/e2e/testdata/s3_backend.tf $dir/s3_backend.tf
  done

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
  "packet")
    export PACKET_AUTH_TOKEN=${PACKET_API_KEY}
    export TF_VAR_project_id=${PACKET_PROJECT_ID}
    ;;
  "gce")
    export GOOGLE_CREDENTIALS=$(echo ${KUBEONE_GOOGLE_SERVICE_ACCOUNT} | base64 -d)
    ;;
  "openstack")
    export OS_AUTH_URL=${OS_AUTH_URL}
    export OS_DOMAIN_NAME=${OS_DOMAIN_NAME}
    export OS_REGION_NAME=${OS_REGION_NAME}
    export OS_TENANT_NAME=${OS_TENANT_NAME}
    export OS_USERNAME=${OS_USERNAME}
    export OS_PASSWORD=${OS_PASSWORD}
    echo ${k1_credentials} >/tmp/credentials.yaml
    ;;
  *)
    echo "unknown provider ${PROVIDER}"
    exit -1
    ;;
  esac

  if [ -d "${KUBETESTS_ROOT}" ]; then
    kubeone_build_dir="_build"
    for kubetest_dir in ${KUBETESTS_ROOT}/*; do
      basekubetest_name=$(basename ${kubetest_dir})
      kubetest_dst_dir="${kubeone_build_dir}/${basekubetest_name}"
      mkdir -p "${kubetest_dst_dir}"
      ln -s $kubetest_dir/* "${kubetest_dst_dir}"
    done
  else
    fail "kubetests directory ${KUBETESTS_ROOT} in not found"
  fi
fi

SSH_PRIVATE_KEY_FILE="${HOME}/.ssh/id_rsa_kubeone_e2e"
export SSH_PUBLIC_KEY_FILE="${SSH_PRIVATE_KEY_FILE}.pub"
export TF_VAR_ssh_public_key_file=${SSH_PUBLIC_KEY_FILE}

if [ ! -f "${SSH_PRIVATE_KEY_FILE}" ]; then
  echo "Generating SSH key pair"
  ssh-keygen -f ${SSH_PRIVATE_KEY_FILE} -N ''
  chmod 400 ${SSH_PRIVATE_KEY_FILE}
  eval $(ssh-agent)
  ssh-add ${SSH_PRIVATE_KEY_FILE}
fi

function runE2E() {
  local test_set=$1
  local timeout=$2
  set -x

  go test \
    -tags=e2e \
    -v \
    -timeout=${timeout} \
    -run=${test_set} \
    ./test/e2e/... \
    -identifier=${BUILD_ID} \
    -provider=${PROVIDER} \
    -os-control-plane=${TEST_OS_CONTROL_PLANE} \
    -os-workers=${TEST_OS_WORKERS} \
    -target-version=${TEST_CLUSTER_TARGET_VERSION} \
    -initial-version=${TEST_CLUSTER_INITIAL_VERSION}
}

# Start the tests
echo "Running E2E tests ..."
case ${TEST_SET} in
"conformance")
  runE2E "TestClusterConformance" "60m"
  ;;
"upgrades")
  runE2E "TestClusterUpgrade" "120m"
  ;;
*)
  echo "unknown TEST_SET: ${TEST_SET}"
  exit -1
  ;;
esac
