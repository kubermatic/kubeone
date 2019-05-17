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
export TFJSON?=
export KUBERNETES_VERSION=1.14.1
BUILD_DATE=$(shell if hash gdate 2>/dev/null; then gdate --rfc-3339=seconds | sed 's/ /T/'; else date --rfc-3339=seconds | sed 's/ /T/'; fi)
BUILD_IMAGE?=golang:1.12.5
GITCOMMIT=$(shell git log -1 --pretty=format:"%H")
GITTAG=$(shell git describe --tags --always)
GOLDFLAGS?=-s -w -X github.com/kubermatic/kubeone/pkg/cmd.version=$(GITTAG) -X github.com/kubermatic/kubeone/pkg/cmd.commit=$(GITCOMMIT) -X github.com/kubermatic/kubeone/pkg/cmd.date=$(BUILD_DATE)

PROVIDER=$(notdir $(wildcard ./terraform/*))
CREATE_TARGETS=$(addsuffix -env,$(PROVIDER))
DESTROY_TARGETS=$(addsuffix -env-cleanup,$(PROVIDER))

all: install

.PHONY: install
install:
	go install -ldflags='$(GOLDFLAGS)' -v .

.PHONY: build
kubeone: build
build: dist/kubeone

.PHONY: lint
lint:
	@golangci-lint --version
	golangci-lint run

.PHONY: test
test:
	CGO_ENABLED=1 go test -race ./...

.PHONY: dep
dep:
	dep ensure -v

.PHONY: licence-check
licence-check:
	wwhrd check

.PHONY: check-dependencies
check-dependencies:
	dep check

docker-make-install:
	docker run -it --rm \
		-v $(PWD):/go/src/github.com/kubermatic/kubeone \
		-v $(GOPATH)/pkg:/go/pkg \
		-w /go/src/github.com/kubermatic/kubeone \
		$(BUILD_IMAGE) \
		make install

.PHONY: e2e_test
e2e_test: build lint test dep
	./hack/run_ci_e2e_test.sh

dist/kubeone: $(shell find . -name '*.go')
	go build -ldflags='$(GOLDFLAGS)' -v -o $@ .

$(CREATE_TARGETS): kubeone
	$(eval PROVIDERNAME := $(@:-env=))
	cd terraform/$(PROVIDERNAME) && terraform apply --auto-approve
	terraform output -state=terraform/$(PROVIDERNAME)/terraform.tfstate -json > tf.json
	$(eval USER := $(shell cat tf.json |jq -r '.kubeone_hosts.value.control_plane[0].ssh_user'|sed 's/null/root/g'))
	for host in $$(cat tf.json |jq -r '.kubeone_hosts.value.control_plane[0].public_address|.[]'); do \
		until ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no $(USER)@$$host exit; do sleep 1; done; \
	done
	./dist/kubeone config print --full --provider $(PROVIDERNAME) > ./dist/fresh_config.yaml
	./dist/kubeone install ./dist/fresh_config.yaml  --tfjson tf.json

$(DESTROY_TARGETS): kubeone
	$(eval PROVIDERNAME := $(@:-env-cleanup=))
	./dist/kubeone config print --full --provider $(PROVIDERNAME) > ./dist/fresh_config.yaml
	./dist/kubeone reset ./dist/fresh_config.yaml  --tfjson tf.json
	cd terraform/$(PROVIDERNAME) && terraform destroy --auto-approve
