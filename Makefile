export GO111MODULE=on
export GOPATH?=$(shell go env GOPATH)
BUILD_IMAGE?=golang:1.11.2

all: install

.PHONY: install
install:
	go install -v ./cmd/kubeone

docker-make-install:
	docker run -it --rm \
		-v $(PWD):/app \
		-v $(GOPATH)/pkg/mod:/go/pkg/mod \
		-w /app \
		$(BUILD_IMAGE) \
		make install
