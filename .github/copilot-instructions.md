# AI Agent Guide for KubeOne
## Big Picture
- KubeOne is a CLI orchestrator that provisions and reconciles Kubernetes HA clusters via Cobra commands rooted in [pkg/cmd/root.go](../pkg/cmd/root.go) and executed from [main.go](../main.go).
- Desired state comes from YAML manifests plus optional Terraform output parsed into internal APIs inside [pkg/apis/kubeone/config](../pkg/apis/kubeone/config).
- Operations converge live infrastructure by comparing cluster telemetry in [pkg/state](../pkg/state) against desired config and scheduling tasks defined in [pkg/tasks](../pkg/tasks).
- Remote actions happen over SSH or local execution, so any new functionality must respect the existing executor abstractions and avoid assuming direct Kubernetes API access early in workflows.
## Core Architecture
- Global CLI options assemble a state.State in [pkg/cmd/shared.go](../pkg/cmd/shared.go) then pass it to command-specific runners such as apply, reset, migrate found under [pkg/cmd](../pkg/cmd).
- state.State holds cluster config, live probes, executor handles, and convenience helpers for node fan-out defined across [pkg/state/context.go](../pkg/state/context.go) and [pkg/state/task.go](../pkg/state/task.go).
- Task pipelines in [pkg/tasks/tasks.go](../pkg/tasks/tasks.go) compose ordered Task structs with predicates; they rely on scripts from [pkg/scripts](../pkg/scripts) rendered per operating system.
- Addons and templated resources are rendered through [pkg/addons](../pkg/addons) and [pkg/templates](../pkg/templates), combining embedded manifests from [addons](../addons) with dynamic credentials and image resolution logic in [pkg/templates/images/images.go](../pkg/templates/images/images.go).
- Credentials and image registries are sourced via [pkg/credentials](../pkg/credentials) and [pkg/templates/images](../pkg/templates/images) to hydrate machine-controller, operating-system-manager, and CSI deployments.
## Configuration & Data Sources
- Manifest parsing accepts v1beta2 API objects and applies defaults plus validation through [pkg/apis/kubeone/config/config.go](../pkg/apis/kubeone/config/config.go); migration utilities live next door for future schema upgrades.
- Terraform output loaders in [pkg/apis/kubeone/config/config.go](../pkg/apis/kubeone/config/config.go) and supporting [pkg/terraform/v1beta3](../pkg/terraform/v1beta3) structs understand both file, stdin, and directory invocation of terraform output -json.
- Embedded Terraform examples in [examples/terraform](../examples/terraform) mirror supported providers; [examples/embed.go](../examples/embed.go) shows how they are published into release archives.
- Cluster features and component overrides rely on helpers in [pkg/features](../pkg/features) and [pkg/apis/kubeone/types.go](../pkg/apis/kubeone/types.go) for validation, so extend those locations when introducing new toggles.
## Execution Flow
- The apply command in [pkg/cmd/apply.go](../pkg/cmd/apply.go) probes hosts, determines install versus upgrade, then executes task bundles like WithFullInstall defined in [pkg/tasks/tasks.go](../pkg/tasks/tasks.go); reuse predicates instead of ad-hoc conditionals.
- state.State fan-out helpers in [pkg/state/task.go](../pkg/state/task.go) manage parallel versus sequential execution and mutate shared state; always clone state when running per-node logic to avoid data races.
- SSH connectors in [pkg/ssh/connector.go](../pkg/ssh/connector.go) cache sessions keyed by host ID; new code should close connections via Executor.Close only when replacing a host to preserve reuse across tasks.
- Image selection and registry overrides come from Images resolver built in [pkg/state/context.go](../pkg/state/context.go); prefer that getter rather than hard-coding container tags.
## Workflows
- Developer builds run make install for a local binary or make build for dist artifacts; make test covers unit tests within ./pkg and ./test while make lint runs golangci-lint.
- End-to-end suites are generated under [test/e2e](../test/e2e) via go generate (make gogenerate) using the matrix defined in [test/tests.yml](../test/tests.yml); run test/go-test-e2e.sh ScenarioName with provider credentials exported.
- Code generation helpers such as [hack/update-codegen.sh](../hack/update-codegen.sh) and [hack/update-apidocs.sh](../hack/update-apidocs.sh) expect GOFLAGS=-mod=readonly; allow go env to manage caches via download-gocache before heavy targets.
- Release packaging leans on [Makefile](../Makefile) targets goreleaser and dist/kubeone plus the embedded addon assets, so update embedded resources before cutting builds.
## Conventions & Pitfalls
- Error handling routes through [pkg/fail](../pkg/fail) helper constructors to preserve exit codes surfaced in [pkg/cmd/root.go](../pkg/cmd/root.go); avoid returning fmt.Errorf directly from task logic.
- Logging uses logrus loggers stored on state.State, set via cobra flag verbosity; prefer s.Logger.WithField for per-node context instead of creating new loggers.
- When touching addons or templated manifests, ensure CredentialsHash and webhook cert helpers in [pkg/addons/applier.go](../pkg/addons/applier.go) stay consistent otherwise checksum triggers unnecessary reapply cycles.
- Machine deployments and OSM integration hinges on flags in [pkg/tasks/tasks.go](../pkg/tasks/tasks.go) and feature detection helpers; coordinate updates across tasks, features, and templates to maintain CCM or CSI migration invariants.
