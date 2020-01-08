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

FROM golang:1.13 AS build
COPY go.mod go.sum /src/
RUN cd /src && go mod download
COPY Makefile /src/
COPY build/kubeoneink8s/ /src/build/kubeoneink8s/
COPY main.go /src/
COPY pkg/ /src/pkg/
RUN cd /src && make dist/kubeone dist/kubeoneink8s && cp dist/* /usr/local/bin/

FROM gcr.io/distroless/static:nonroot
WORKDIR /tmp
USER nobody
ENTRYPOINT ["/usr/local/bin/kubeone"]
COPY --from=build /usr/local/bin/* /usr/local/bin/
