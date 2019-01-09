# KubeOne

KubeOne makes installing and upgrading a HA Kubernetes cluster a breeze.
It integrates with your favorite infrastructure tools or can be used standalone.

From a bird eye view, KubeOne connects to your machines using SSH and then uses
`kubeadm` to bootstrap the cluster.

## Features

* Kubernetes 1.12 support
* easy Terraform integration
* provides rolling cluster upgrades (TODO)

## Usage

First, build the binary by running `make`. After that you need to provide
KubeOne with a configuration with details about the node IPs, Kubernetes
versions etc.

By default, the configuration happens by providing a single YAML file. Take a
look at the `config.yaml.dist` for more details and create a copy of it to make
adjustments.

Armed with your configuration, you can now run KubeOne:

    kubeone install myconfig.yaml

This will SSH into the machines, install the required dependencies and then
perform the necessary steps to bring up an etcd ring and a HA Kubernetes
control plane.

## Workers definition

KubeOne relies on the [machine-controller
project](https://github.com/kubermatic/machine-controller/) to create worker nodes after the Kubernetes master nodes are bootstrapped.
The machine controller will be deployed as part of the cluster creation and is configured using the cloud provider credentials from the shell environment.
(`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` in case of AWS)

Worker nodes are managed by the [machine-controller
project](https://github.com/kubermatic/machine-controller/) (which kubeone
automatically deploys).

The config section `workers: []` specifies `MachineDeployment` objects that will be created when KubeOne runs.

You can find example machine deployments in the [machine-controller examples](https://github.com/kubermatic/machine-controller/blob/master/examples/aws-machinedeployment.yaml).

## Terraform Integration

KubeOne supports discovery of `hosts`, the Kubernetes API LB and can read worker configuration from Terraform's output.
Take a look at the files in the `terraform/aws` directory and especially the `output.tf` file to learn more about what data KubeOne expects to read from Terraform.

To use Terraform's output, use the `--tfjson` CLI flag:

    terraform apply
    terraform output -json > tf.json

    kubeone install --tfjson tf.json myconfig.yaml


### Worker configuration from Terraform

While the `hosts: []` configuration will be completely taken from Terraform's output, the `workers: []` configuration will be merged by populating empty config fields.

Output value names and types correspond to `cloudProviderSpec` fields for the chosen
provider, for more information please consult the `RawConfig` structure in corresponding provider package in the [machine-controller provider](https://github.com/kubermatic/machine-controller/tree/master/pkg/cloudprovider/provider) folder.

The `terraform/aws` example and the `config.yaml.dist` file contain multiple worker configs that use this mechanism to add `availabilityZone`, `ami`, etc. fields.

## Debugging

To see exactly what's happening during installation, pass the `--verbose` flag
to KubeOne:

    kubeone --verbose install myconfig.yaml

This will stream the output of all shell commands to your shell.

## License

Apache License
