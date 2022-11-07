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

FROM golang:1.19.3 as builder

RUN apt-get update && apt-get install -y \
    unzip

WORKDIR /download

ENV TERRAFORM_VERSION "1.3.2"
RUN curl -fL https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip | funzip >/usr/local/bin/terraform
RUN chmod +x /usr/local/bin/terraform

ENV SONOBUOY_VERSION "0.56.10"
RUN curl -fL https://github.com/vmware-tanzu/sonobuoy/releases/download/v${SONOBUOY_VERSION}/sonobuoy_${SONOBUOY_VERSION}_linux_amd64.tar.gz | tar vxz
RUN chmod +x sonobuoy

ENV KUBECTL_VERSION="v1.24.7"
RUN curl --fail -Lo kubectl https://dl.k8s.io/release/$KUBECTL_VERSION/bin/linux/amd64/kubectl
RUN chmod +x kubectl

# resulting image

FROM golang:1.19.3

ARG version

LABEL "io.kubeone"="Kubermatic GmbH"
LABEL version=${version}
LABEL description="Set of kubernetes binaries to conduct kubeone E2E tests"
LABEL maintainer="https://github.com/kubermatic/kubeone/blob/main/OWNERS"

ENV USER root

# install base tools
RUN apt-get update && \
    apt-get install -y \
      apt-transport-https \
      bash \
      bash-completion \
      ca-certificates \
      curl \
      git \
      htop \
      make \
      nano \
      openssh-client \
      rsync \
      software-properties-common \
      unzip \
      vim \
      zip

# make bash much more pleasant to use
COPY bashrc /root/.bashrc
COPY htoprc /root/.config/htop/htoprc

COPY --from=builder /usr/local/bin/terraform /usr/local/bin/terraform
COPY --from=builder /download/sonobuoy /usr/local/bin/sonobuoy
COPY --from=builder /download/kubectl /usr/local/bin/kubectl
