export GO111MODULE=on

all: install

install:
	go install -v ./cmd/...
