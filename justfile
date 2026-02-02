set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

export CGO_ENABLED := "0"
export GOPROXY := "https://proxy.golang.org"
export GOFLAGS := "-mod=readonly -trimpath"

DEFAULT_STABLE := `curl -SsL https://dl.k8s.io/release/stable-1.35.txt`

# Install the build binary to $GOBIN
[default]
[group('build')]
install: _buildenv
    go install -v .

# Run golangci-lint on the codebase
[group('test')]
lint:
    golangci-lint run --timeout=5m -v ./pkg/... ./test/...

# Build binary without copying to $GOBIN
[group('build')]
build:

# Run unit tests
[group('test')]
test:

# Run end-to-end tests, involving external dependencies on cloud providers
[group('test')]
e2e-test:

# Remove different artifacts produces by the build processes (vendor, binaries, etc.)
clean:

# Dump dependencies to ./vendor
vendor:
    go mod vendor

# Rerun different code generators
[group('build')]
update-codegen: gogenerate

# Run go generate in all packages
[group('build')]
gogenerate:
	go generate ./pkg/...
	go generate ./test/...

# Format shell scripts
[group('build')]
shfmt:
	shfmt -w -sr -i 2 hack

# Format terraform files
[group('build')]
tffmt:
	terraform fmt -write=true -recursive .

# Format Go files
[group('build')]
gofmt:
    gofmt ./pkg/...

# Format imports in Go files
[group('build')]
gimps:
    go run go.xrstf.de/gimps@latest .

# Run all formatters
[group('build')]
fmt: shfmt tffmt gofmt gimps

# Verify that codegen is up to date
[group('test')]
verify-codegen: vendor

# Verify that API documentation is up to date
[group('test')]
verify-apidocs: vendor

# Verify that all boilerplate in files is up to date
[group('test')]
verify-boilerplate:

# Create a new release
release: _buildenv


### Helper private recipes ###

_buildenv:
    @go version
