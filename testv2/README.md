# Generated E2E tests usage

The matrix of possible combinations of generated tests is located at [tests.yml](tests.yml).
The API of this matrix is:

```go
type Infrastructure struct {
	Name      string `json:"name"`
	AlwaysRun bool   `json:"alwaysRun"`
	Optional  bool   `json:"optional"`
}

type KubeoneTest struct {
	Scenario        string           `json:"scenario"`
	InitVersion     string           `json:"initVersion"`
	UpgradedVersion string           `json:"upgradedVersion"`
	Infrastructures []Infrastructure `json:"infrastructures"`
}
```

Each scenario represents a scenario from e2e package (an implementation of
`e2e.Scenario` interface).

Example:

```yaml
- scenario: install_containerd
  initVersion: v1.21.14
  infrastructures:
    - name: aws_defaults
      alwaysRun: true
    - name: openstack_default
```

This can be "decoded" as an instruction to generate "install_containerd"
scenarion, for kubernetes version v1.21.14 and run it on default aws and openstack
infrastructures.

## Scenario

Represents a set of actions to run + kubeone configuraton groupped together.
Currently we have 3 basic scenarios to run:

* `scenarioInstall`

    This will install the cluster and run basic tests along with some smaller
    subset of sonobuoy e2e tests.

* `scenarioUpgrade`

    This will use `scenarioInstall` to init the cluster, the will run the
    upgrade proceedure, following by basic tests along with some smaller subset
    of sonobuoy e2e tests.

* `scenarioConformance`

    This will use `scenarioInstall` to init the cluster, the will run some basic
    tests along with some full blown sonobuoy conformance tests.

## Infras

Infra references the terraform config to use and it's variables. Multiplied
together with Scenarions they form a matrix of diffeernt cloud providers /
version / configuration options

## Regenerating tests

```shell
make gogenerate
```

Will rerun `go generate` in respected directories.

## Adding new cases

Infras and Scenarios are being defined in [tests_definitions.go](e2e/tests_definitions.go).

Those definitions are used to express the veriability of available configs /
things to test.

Once definitions with needed configs are written down, they can be referenced in
[tests.yml](tests.yml) to actually produce the generated code, prow and Go tests
that will be in sync with each other.

## Generated code

The generator will generate and overwrite [tests_test.go](e2e/tests_test.go)
plus prow.yaml config with corresponding calls to generated test functions.

## Running generated tests

There is a shell [go-test-e2e.sh](go-test-e2e.sh) scrint to run small setup
proceedures (like generating SSH keys and extracting auth variables from
envirionmen). It's being used in generated prow cases. It's possible to launch
it manually.

```shell
export HCLOUD_TOKEN=xxx
./testv2/go-test-e2e.sh TestHetznerDefaultInstallContainerdV1_22_11
```
