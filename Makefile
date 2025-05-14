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

# follow up on: https://tech.davis-hansson.com/p/make/

# do not rely on /bin/sh, it can be a symlink to anything (sh/bash/dash/busybox/etc).
SHELL := bash

# ensures each Make task is ran as one single shell session, rather than one new shell per line.
.ONESHELL:

# if a Make rule fails, it’s target file is deleted.
.DELETE_ON_ERROR:

# pass strict shell flags, to fail early
.SHELLFLAGS := -eu -o pipefail -c

# be loud about missing make variables
MAKEFLAGS += --warn-undefined-variables

# disable magic rules
MAKEFLAGS += --no-builtin-rules

export GOPATH?=$(shell go env GOPATH)
export CGO_ENABLED=0
export GOPROXY?=https://proxy.golang.org
export GO111MODULE=on
export GOFLAGS?=-mod=readonly -trimpath
export DEFAULT_STABLE=$(shell curl -SsL https://dl.k8s.io/release/stable-1.33.txt)

BUILD_DATE=$(shell if hash gdate 2>/dev/null; then gdate --rfc-3339=seconds | sed 's/ /T/'; else date --rfc-3339=seconds | sed 's/ /T/'; fi)
GITCOMMIT=$(shell git log -1 --pretty=format:"%H")
GITTAG=$(shell git describe --tags --always)
GOLDFLAGS?=-s -w -extldflags=-zrelro -extldflags=-znow \
	-X k8c.io/kubeone/pkg/cmd.defaultKubeVersion=$(DEFAULT_STABLE) \
	-X k8c.io/kubeone/pkg/cmd.version=$(GITTAG) \
	-X k8c.io/kubeone/pkg/cmd.commit=$(GITCOMMIT) \
	-X k8c.io/kubeone/pkg/cmd.date=$(BUILD_DATE)

GORELEASER_FLAGS ?= --clean

.PHONY: all
all: install

.PHONY: install
install: buildenv
	go install -ldflags='$(GOLDFLAGS)' -v .

.PHONY: build
build: dist/kubeone

.PHONY: clean
clean:
	rm -f dist/kubeone

.PHONY: vendor
vendor: buildenv
	go mod vendor

dist/kubeone: buildenv download-gocache
	go build -ldflags='$(GOLDFLAGS)' -v -o $@ .

dist/kubeone-debug: buildenv
	export GOFLAGS=-mod=readonly; \
	go build -gcflags='all=-N -l' -v -o $@ .

download-gocache:
	@./hack/ci/download-gocache.sh
	@# Prevent this from getting executed multiple times
	@touch download-gocache

.PHONY: update-codegen
update-codegen: GOFLAGS = -mod=readonly
update-codegen: vendor
	./hack/update-codegen.sh
	rm -rf ./vendor

.PHONY: test
test: download-gocache
	go test ./pkg/... ./test/...

.PHONY: e2e-test
e2e-test: download-gocache install
	./hack/run-ci-e2e-test.sh

.PHONY: buildenv
buildenv:
	@go version

.PHONY: lint
lint:
	golangci-lint run --timeout=5m -v ./pkg/... ./test/...

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

.PHONE: verify-apidocs
verify-apidocs: GOFLAGS = -mod=readonly
verify-apidocs: vendor
	./hack/verify-apidocs.sh

.PHONY: shfmt
shfmt:
	shfmt -w -sr -i 2 hack

.PHONY: prowfmt
prowfmt:
	yq --inplace eval .prow.yaml

.PHONY: tffmt
tffmt:
	terraform fmt -write=true -recursive .

fmt: shfmt prowfmt tffmt

gogenerate:
	go generate ./pkg/...
	go generate ./test/...

.PHONY: goreleaser
goreleaser:
	goreleaser release $(GORELEASER_FLAGS)
