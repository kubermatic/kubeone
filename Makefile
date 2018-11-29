export GO111MODULE=on
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
		-v $(PWD):/app \
		-v $(GOPATH)/pkg/mod:/go/pkg/mod \
		-w /app \
		$(BUILD_IMAGE) \
		make install
