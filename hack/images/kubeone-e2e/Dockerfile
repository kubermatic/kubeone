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

# building image

FROM golang:1.17.5 as builder

RUN apt-get update && apt-get install -y \
    unzip

ENV TERRAFORM_VERSION "1.1.0"
RUN curl -fL https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip | funzip >/usr/local/bin/terraform
RUN chmod +x /usr/local/bin/terraform

ENV KUBETEST_COMMIT "eb1a87c9083a31f9a9dda7822cb674f9f8d99439"
ENV KUBETESTS_ROOT "/opt/kube-test"
ENV GO111MODULE=on

# Install the kubetest binary
RUN go get -v k8s.io/test-infra/kubetest@${KUBETEST_COMMIT} && \
    go clean -modcache -cache

COPY install-kube-tests-binaries.sh /opt/install-kube-tests-binaries.sh
RUN /opt/install-kube-tests-binaries.sh

# resulting image


FROM golang:1.17.5

ARG version

LABEL "io.kubeone"="Kubermatic GmbH"
LABEL version=${version}
LABEL description="Set of kubernetes binaries to conduct kubeone E2E tests"
LABEL maintainer="https://github.com/kubermatic/kubeone/blob/master/OWNERS"

ENV KUBETESTS_ROOT "/opt/kube-test"
ENV GO111MODULE on
ENV USER root

COPY --from=builder /usr/local/bin/terraform /usr/local/bin/terraform
COPY --from=builder /go/bin/kubetest /usr/local/bin/kubetest
COPY --from=builder ${KUBETESTS_ROOT} ${KUBETESTS_ROOT}
