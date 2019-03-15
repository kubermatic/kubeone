# Project Structure

The Project Structure document explains how the project is structured and what's the responsibility of each package.

## CLI

The KubeOne CLI is built using [Cobra](https://github.com/spf13/cobra). All commands and the relevant logic are implemented in the [`pkg/cmd`](https://github.com/kubermatic/kubeone/tree/master/pkg/cmd) package. The application entrypoint is [`main.go`](https://github.com/kubermatic/kubeone/blob/master/main.go) located at the root level of the repository.

```
├── pkg
│   ├── cmd
│   │   ├── install.go      # Install Kubernetes
│   │   ├── kubeconfig.go   # Download Kubeconfig
│   │   ├── reset.go        # Reset cluster and destroy worker nodes
│   │   ├── root.go         # Root command
│   │   ├── shared.go       # Shared logic used by all commands
│   │   ├── upgrade.go      # Upgrade cluster
│   │   └── version.go      # KubeOne and machine-controller version
├── main.go                 # Application entrypoint
```

## Documentation

The project documentation is located in the [`docs`](https://github.com/kubermatic/kubeone/tree/master/docs) directory. Proposals for upcoming features are located in the [`docs/proposals`](https://github.com/kubermatic/kubeone/tree/master/docs/proposals) directory. We're currently using GitHub for hosting and previewing documentation, but we're researching other options as well.

```
├── docs                            # KubeOne documentation
│   ├── proposals                   # Proposals for upcoming features
│   │   └── 0002-upgrades.md
│   ├── environment_variables.md
│   ├── project_structure.md
│   ├── quickstart-aws.md
│   └── quickstart-digitalocean.md
├── CONTRIBUTING.md                 # KubeOne Contributor Guide
├── README.md
```

## Terrafrom scripts

The example Terraform scripts are located in the [`examples/terraform`](https://github.com/kubermatic/kubeone/tree/master/examples/terraform) directory. In the `terraform` directory you can find a subdirectory for each cloud provider.

```
├── examples
│   └── terraform           # Terraform example scripts
│       ├── aws             # Scripts for Amazon Web Services (AWS)
│       ├── digitalocean    # Scripts for DigitalOcean
│       ├── hetzner         # Scripts for Hetzner
│       └── openstack       # Scripts for OpenStack
```

## Packages

The [`pkg`](https://github.com/kubermatic/kubeone/tree/master/pkg) package has all KubeOne's functionality. Below you can find a list of `pkg` packages along with their responsibilities.

```
├── pkg
│   ├── apis                # External APIs used by KubeOne
│   │   └── kubeadm             # kubeadm v1beta1 API
│   │       └── v1beta1
│   ├── archive             # Create .tar.gz archive
│   ├── certificate         # Generate certificates needed for the machine-controller webhook
│   ├── cmd                 # KubeOne CLI
│   ├── config              # KubeOne Configuration API
│   ├── features            # Activate optional cluster features (e.g. dynamic audit, PodSecurityPolicy)
│   ├── installer           # Install Kubernetes
│   │   └── installation        # Scripts used to install Kubernetes
│   ├── ssh                 # Connect to nodes over SSH
│   ├── templates           # Templates for Kubernetes resources
│   │   ├── canal               # Canal CNI resources
│   │   ├── kubeadm             # Parses kubeadm v1beta1 config
│   │   │   └── v1beta1
│   │   └── machinecontroller   # machine-controller resources
│   ├── terraform           # Sources KubeOne config from the Terraform output
│   ├── upgrader            # Upgrade Kubernetes
│   │   └── upgrade             # Scripts used to upgrade Kubernetes
│   └── util                # Common-used functions
```

## End-To-End Tests

End-To-End tests are located in the [`test/e2e`](https://github.com/kubermatic/kubeone/tree/master/test/e2e) package. To run E2E test in the CI pipeline, a project maintainer needs to create a [ProwJob](https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md) for that test suite.

```
├── test
│   └── e2e
│       ├── testdata                # Static manifests used in tests
│       ├── conformance_test.go     # Conformance tests
│       ├── helper.go
│       ├── kubeone.go
│       ├── kubetest.go
│       ├── main_test.go
│       ├── provisioner.go
│       └── upgrade_test.go         # Upgrades tests
```
