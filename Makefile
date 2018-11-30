export GOPATH?=$(shell go env GOPATH)
export CGO_ENABLED=0
BUILD_IMAGE?=golang:1.11.2

all: install

.PHONY: install
install:
	go install -v ./cmd/kubeone

.PHONY: build
build: dist/kubeone

dist/kubeone:
	go build -v -o $@ ./cmd/kubeone

docker-make-install:
	docker run -it --rm \
		-v $(PWD):/go/src/github.com/kubermatic/kubeone \
		-v $(GOPATH)/pkg:/go/pkg \
		-w /go/src/github.com/kubermatic/kubeone \
		$(BUILD_IMAGE) \
		make install
