export GOPATH?=$(shell go env GOPATH)
export CGO_ENABLED=0
export TFJSON?=
export KUBEONE_CONFIG_FILE?=config.yaml.dist
export KUBERNETES_VERSION=1.13.3
BUILD_DATE=$(shell if hash gdate 2>/dev/null; then gdate --rfc-3339=seconds | sed 's/ /T/'; else date --rfc-3339=seconds | sed 's/ /T/'; fi)
BUILD_IMAGE?=golang:1.11.5
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
	dep status
	dep check

docker-make-install:
	docker run -it --rm \
		-v $(PWD):/go/src/github.com/kubermatic/kubeone \
		-v $(GOPATH)/pkg:/go/pkg \
		-w /go/src/github.com/kubermatic/kubeone \
		$(BUILD_IMAGE) \
		make install

.PHONY: e2e_test
e2e_test:
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
	./dist/kubeone install config.yaml.dist  --tfjson tf.json

$(DESTROY_TARGETS): kubeone
	$(eval PROVIDERNAME := $(@:-env-cleanup=))
	./dist/kubeone reset config.yaml.dist  --tfjson tf.json
	cd terraform/$(PROVIDERNAME) && terraform destroy --auto-approve
