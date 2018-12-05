#!/bin/bash

# exit on any error
set -e

KUBEONE_ROOT=$(dirname "${BASH_SOURCE}")/..
BUILD_PATH=${KUBEONE_ROOT}/_build
KUBECONFIG=${HOME}/.kube/config

# Make sure KUBEONE_CONFIG_FILE is properly set
if [[ -z ${KUBEONE_CONFIG_FILE} ]]; then
    echo "Please export KUBEONE_CONFIG_FILE in your env"
    exit 1
fi

# Make sure TFJSON is set
if [[ -z ${TFJSON} ]]; then
    echo "[WARNING] Please export TFJSON in your env if you use terraform for infrastructure deployment"
fi

KUBERNETES_VERSION=${KUBERNETES_VERSION:-$(grep 'kubernetes:' ${KUBEONE_CONFIG_FILE} | sed 's/[:[:alpha:]|(|[:space:]]//g'| sed "s/['\"]//g")}
semver=( ${KUBERNETES_VERSION//./ } )
KUBERNETES_MAJOR_VERSION="${semver[0]}"
KUBERNETES_MINOR_VERSION="${semver[1]}"

create_kubeconfig() {
  echo "creating kubeconfig"
  mkdir -p ${HOME}/.kube
  if [[ -z ${TFJSON} ]]; then
    kubeone kubeconfig $(KUBEONE_CONFIG_FILE) > ${KUBECONFIG}
  else
    kubeone kubeconfig --tfjson ${TFJSON} ${KUBEONE_CONFIG_FILE} > ${KUBECONFIG}
  fi
}

# Start e2e conformance tests
start_tests() {
  echo "start e2e tests"

  export KUBERNETES_CONFORMANCE_TEST=y
  export SKIP="Alpha|\[(Disruptive|Feature:[^\]]+|Flaky)\]"

  version=""

  # For < Kubernetes 1.12 use
  if [ "${KUBERNETES_MINOR_VERSION}" -lt 12 ];then
    export SKIP="Alpha|Kubectl|\[(Disruptive|Feature:[^\]]+|Flaky)\]"
  fi

  echo "get kubetest"
  go get -u k8s.io/test-infra/kubetest

  mkdir -p ${BUILD_PATH}
  (
    cd ${BUILD_PATH}

    # check if kubernetes in the same version already exists
    if [ -f "./kubernetes/version" ];then
      version=$(head -n 1 ./kubernetes/version)
      if [ "v${KUBERNETES_VERSION}" != "${version}" ]; then
        rm -rf ./kubernetes
        kubetest --extract=v${KUBERNETES_VERSION}
        rm kubernetes.tar.gz
      fi
    else
      kubetest --extract=v${KUBERNETES_VERSION}
      rm kubernetes.tar.gz
    fi

    cd ./kubernetes
    kubetest --provider=skeleton \
         --test \
         --ginkgo-parallel \
         --test_args="--ginkgo.focus=\[Conformance\] --ginkgo.skip=${SKIP} "
  )
}

create_kubeconfig
start_tests
