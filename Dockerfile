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

FROM docker.io/golang:1.19.3 as builder

ARG GOPROXY=
ENV GOPROXY=$GOPROXY

ARG GOCACHE=
ENV GOCACHE=$GOCACHE

WORKDIR /go/src/k8c.io/kubeone
COPY . .
RUN make build

FROM docker.io/alpine:3.15
LABEL maintainer="support@kubermatic.com"

# openssh-client is required for the ssh binary and for ssh-agent
RUN apk add --no-cache openssh-client

COPY --from=builder /go/src/k8c.io/kubeone/dist/kubeone /usr/local/bin/kubeone
ENTRYPOINT ["/usr/local/bin/kubeone"]
