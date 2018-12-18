#!/usr/bin/env bash

# This script is mostly used in CI
# It installs dependencies and starts the tests

set -euxo pipefail

IS_DEBUG=$(echo "$-" | grep "x")
RUNNING_IN_CI=${JOB_NAME:-""}

# Install dependencies
if ! [ -x "$(command -v unzip)" ]; then
 apt update && apt install -y unzip
fi
if ! [ -x "$(command -v terraform)" ]; then
(
  echo "Installing terraform"
  cd /tmp
  curl https://releases.hashicorp.com/terraform/0.11.10/terraform_0.11.10_linux_amd64.zip -Lo terraform.zip
  unzip -n terraform.zip terraform
  chmod +x terraform
  mv terraform /usr/local/bin
  rm terraform.zip
 )
fi
if ! [ -x "$(command -v kubetest)" ]; then
  echo "Installing kubetest"
  go get k8s.io/test-infra/kubetest
  PATH=$PATH:$(go env GOPATH)/bin
fi
# If the following variable is set then this script is running in CI
# and the assumption is that the image contains kubernetes binaries
#
# note:
# kubetest assumes that the last part of that path contains "kubernetes", if not then it complains.
# additionally if the version doesn't match the reg exp it will be changed for example, from 1.12 to 1.12.0
if [ -n "${RUNNING_IN_CI}" ]; then
 # terraform expects to find AWS credentials in the following env variables
 if [ -n ${IS_DEBUG} ]; then set +x; fi
 export AWS_ACCESS_KEY_ID=$AWS_E2E_TESTS_KEY_ID
 export AWS_SECRET_ACCESS_KEY=$AWS_E2E_TESTS_SECRET
 if [ -n ${IS_DEBUG} ]; then set -x; fi
 KUBE_TEST_DIR="/opt/kube-test"
 if [ -d "${KUBE_TEST_DIR}" ]; then
 KUBEONE_BUILD_DIR="$(go env GOPATH)/src/github.com/kubermatic/kubeone/_build"
 mkdir -p ${KUBEONE_BUILD_DIR}
 for dir in ${KUBE_TEST_DIR}/*
  do
   VERSION_REG_EXP="^(\d+\.\d+\.\d+[\w.\-+]*)$"
   KUBE_VERSION=$(basename $dir)
   if ! [[ ${KUBE_VERSION} =~ ${VERSION_REG_EXP} ]]; then
    KUBE_VERSION="${KUBE_VERSION}.0"
   fi
   KUBE_TEST_DST_DIR="${KUBEONE_BUILD_DIR}/kubernetes-v${KUBE_VERSION}/kubernetes"
   mkdir -p "${KUBE_TEST_DST_DIR}"
   ln -s $dir/* "${KUBE_TEST_DST_DIR}"
 done
 else
  echo "The directory ${KUBE_TEST_DIR} does not exist, we need to download additional binaries for the tests. This might make the test to run longer."
 fi
 else
  echo "The script is not running in CI thus we need to download additional binaries for the tests. This might make the test to run longer."
fi

# Generate SSH key pair
if [ ! -f "$HOME/.ssh/id_rsa_kubeone_e2e" ]; then
 echo "Generating SSH key pair"
 ssh-keygen -f $HOME/.ssh/id_rsa_kubeone_e2e -N ''
 SSH_PUBLIC_KEY_FILE="$HOME/.ssh/id_rsa_kubeone_e2e.pub"
 SSH_PRIVATE_KEY_FILE="$HOME/.ssh/id_rsa_kubeone_e2e"
 export SSH_PUBLIC_KEY_FILE
 chmod 400 ${SSH_PRIVATE_KEY_FILE}
 eval `ssh-agent`
 ssh-add ${SSH_PRIVATE_KEY_FILE}
fi

# Build binaries
echo "Building kubeone ..."
make install

# Start the tests
echo "Running E2E tests ..."
go test -race -tags=e2e -v -timeout 30m  ./test/e2e/...
