export GOPATH?=$(shell go env GOPATH)
export CGO_ENABLED=0
export TFJSON?=
export KUBEONE_CONFIG_FILE?=config.yaml.dist
export KUBERNETES_VERSION=1.12.3
BUILD_IMAGE?=golang:1.11.2

.PHONY: build install e2e_test dep

all: install

install:
	go install -v ./cmd/kubeone

build: dist/kubeone

lint:
	golangci-lint run

dep:
	dep ensure -v

docker-make-install:
	docker run -it --rm \
		-v $(PWD):/go/src/github.com/kubermatic/kubeone \
		-v $(GOPATH)/pkg:/go/pkg \
		-w /go/src/github.com/kubermatic/kubeone \
		$(BUILD_IMAGE) \
		make install

e2e_test:
	./hack/run_cluster_e2e.sh

dist/kubeone:
	go build -v -o $@ ./cmd/kubeone
