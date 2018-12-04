export GOPATH?=$(shell go env GOPATH)
export CGO_ENABLED=0
BUILD_IMAGE?=golang:1.11.2
TFJSON?=
KUBEONE_CONFIG_FILE?=config.yaml.dist
all: install

install:
	go install -v ./cmd/kubeone


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

e2e_cluster:
	./hack/run_cluster_e2e.sh

.PHONY: build install e2e_cluster
