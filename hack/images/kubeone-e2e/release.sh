#!/usr/bin/env bash

set -euox pipefail

TAG=v0.1.0

docker build --pull -t kubermatic/kubeone-e2e:${TAG} .
docker push kubermatic/kubeone-e2e:${TAG}
