#!/bin/bash

# exit on any error
set -e

KUBEONE_ROOT=$(dirname "${BASH_SOURCE}")/..
BUILD_PATH=${KUBEONE_ROOT}/_build
KUBECONFIG_PATH=$HOME/.kube/config

MASTER_IP=""
MASTER_URL=""
KUBERNETES_VERSION=""
KUBERNETES_MAJOR_VERSION=""
KUBERNETES_MINOR_VERSION=""

# Make sure KUBEONE_CONFIG_FILE is properly set
if [[ -z ${KUBEONE_CONFIG_FILE} ]]; then
    echo "Please export KUBEONE_CONFIG_FILE in your env"
    exit 1
fi

# Make sure TFJSON is set
if [[ -z ${TFJSON} ]]; then
    echo "[WARNING] Please export TFJSON in your env if you use terraform for infrastructure deployment"
fi

create_kubeconfig() {
  echo "creating kubeconfig"
  mkdir -p ${BUILD_PATH}
  if [[ -z ${TFJSON} ]]; then
    kubeone kubeconfig $(KUBEONE_CONFIG_FILE) > ${KUBECONFIG_PATH}
  else
    kubeone kubeconfig --tfjson ${TFJSON} ${KUBEONE_CONFIG_FILE} > ${KUBECONFIG_PATH}
  fi
}

get_master_address() {
  MASTER_IP=$(grep -A0 'server:' ${KUBECONFIG_PATH} | grep -oE '((1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\.){3}(1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])')
  MASTER_URL=$(grep -A0 'server:' ${KUBECONFIG_PATH} | sed -n 's#.*\(https*://[^"]*\).*#\1#;p')
}

get_k8s_version() {
  echo "get k8s version"
  KUBERNETES_VERSION=$(grep 'kubernetes:' ${KUBEONE_CONFIG_FILE} | sed 's/[:[:alpha:]|(|[:space:]]//g'| sed "s/['\"]//g")
  semver=( ${KUBERNETES_VERSION//./ } )
  KUBERNETES_MAJOR_VERSION="${semver[0]}"
  KUBERNETES_MINOR_VERSION="${semver[1]}"
}

# Start e2e conformance tests
start_tests() {
  echo "start e2e tests"
  echo "KUBE_MASTER=${MASTER_IP}"
  echo "KUBE_MASTER_IP=${MASTER_IP}"
  echo "MASTER_URL=${MASTER_URL}"

  export KUBERNETES_PROVIDER=skeleton
  export KUBE_MASTER=${MASTER_IP}
  export KUBE_MASTER_IP=${MASTER_IP}
  export SKIP="Alpha|\[(Disruptive|Feature:[^\]]+|Flaky)\]"

  version=""

  # For < Kubernetes 1.12 use
  if [ "${KUBERNETES_MINOR_VERSION}" -lt 12 ];then
    export SKIP="Alpha|Kubectl|\[(Disruptive|Feature:[^\]]+|Flaky)\]"
  fi

  echo "get kubetest"
  go get -u k8s.io/test-infra/kubetest

  (
    cd ${BUILD_PATH}

    # check if kubernetes in the same version already exists
    if [ -f "./kubernetes/version" ];then
      version=$(head -n 1 ./kubernetes/version)
      if [ "v${KUBERNETES_VERSION}" != "${version}" ]; then
        rm -rf ./kubernetes
        kubetest -v --extract=v${KUBERNETES_VERSION}
        rm kubernetes.tar.gz
      fi
    else
      kubetest -v --extract=v${KUBERNETES_VERSION}
      rm kubernetes.tar.gz
    fi

    cd ./kubernetes
    kubetest --provider=skeleton \
         --test \
         --ginkgo-parallel \
         --test_args="--ginkgo.focus=\[Conformance\] --ginkgo.skip=${SKIP} --kubeconfig=${KUBECONFIG_PATH} --host=${MASTER_URL}" \
         --kubeconfig=${KUBECONFIG_PATH}
  )
}

create_kubeconfig
get_k8s_version
get_master_address
start_tests
