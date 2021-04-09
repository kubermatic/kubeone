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

export GOPATH?=$(shell go env GOPATH)
export CGO_ENABLED=0
export GOPROXY?=https://proxy.golang.org
export GO111MODULE=on
export GOFLAGS?=-mod=readonly -trimpath

BUILD_DATE=$(shell if hash gdate 2>/dev/null; then gdate --rfc-3339=seconds | sed 's/ /T/'; else date --rfc-3339=seconds | sed 's/ /T/'; fi)
GITCOMMIT=$(shell git log -1 --pretty=format:"%H")
GITTAG=$(shell git describe --tags --always)
GOLDFLAGS?=-s -w -extldflags=-zrelro -extldflags=-znow \
	-X k8c.io/kubeone/pkg/cmd.version=$(GITTAG) \
	-X k8c.io/kubeone/pkg/cmd.commit=$(GITCOMMIT) \
	-X k8c.io/kubeone/pkg/cmd.date=$(BUILD_DATE)

.PHONY: all
all: install

.PHONY: install
install: buildenv
	go install -ldflags='$(GOLDFLAGS)' -v .

.PHONY: build
build: dist/kubeone

.PHONY: vendor
vendor: buildenv
	go mod vendor

dist/kubeone: buildenv
	go build -ldflags='$(GOLDFLAGS)' -v -o $@ .

dist/kubeone-debug: buildenv
	export GOFLAGS=-mod=readonly; \
	go build -gcflags='all=-N -l' -v -o $@ .

.PHONY: generate-internal-groups
generate-internal-groups: GOFLAGS = -mod=readonly
generate-internal-groups: vendor
	./hack/update-codegen.sh

.PHONY: test
test:
	go test ./pkg/... ./test/...

.PHONY: e2e-test
e2e-test: install
	./hack/run-ci-e2e-test.sh

.PHONY: buildenv
buildenv:
	@go version

.PHONY: lint
lint:
	@golangci-lint --version
	golangci-lint run -v ./pkg/... ./test/...

.PHONY: verify-licence
verify-licence: GOFLAGS = -mod=readonly
verify-licence: vendor
	wwhrd check

.PHONY: verify-codegen
verify-codegen: GOFLAGS = -mod=readonly
verify-codegen: vendor
	./hack/verify-codegen.sh

.PHONY: verify-boilerplate
verify-boilerplate:
	./hack/verify-boilerplate.sh

.PHONY: shfmt
shfmt:
	shfmt -w -sr -i 2 hack

.PHONY: prowfmt
prowfmt:
	yq --inplace eval .prow.yaml

fmt: shfmt prowfmt
