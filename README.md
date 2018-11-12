# KubeOne

KubeOne makes installing and upgrading a HA Kubernetes cluster a breeze.
It integrates with your favorite infrastructure tools or can be used standalone.

From a bird eye view, KubeOne connects to your machines using SSH and then uses
`kubeadm` to bootstrap the cluster.

## Features

* Kubernetes 1.12 support
* easy Terraform integration
* provides rolling cluster upgrades

## Usage

First, build the binary by running `make`. After that you need to provide KubeOne
with a configuration with details about the node IPs, Kubernetes versions etc.

By default, the configuration happens by providing a single YAML file. Take a look
at the `config.yaml.dist` for more details and create a copy of it to make
adjustments.

Armed with your configuration, you can now run KubeOne:

    ./kubeone install myconfig.yaml

This will SSH into the machines, install the required dependencies and then perform
the neccessary steps to bring up an etcd ring and a HA Kubernetes control plane.

## Terraform Integration

If you use Terraform to provision your infrastructure, you can make KubeOne read its
output to learn about the nodes in your cluster. Take a look at the `terraform/aws`
and especially the `output.tf` file to learn more about what data KubeOne expects to
read from Terraform.

To use Terraform's output, use the `--tfjson` CLI flag:

    terraform apply
    terraform output -json > tf.json
    ./kubeone install --tfjson tf.json myconfig.yaml

This will overwrite the nodes in your `myconfig.yaml` (if any) before installing
Kubernetes.

## Debugging

To see exactly what's happening during installation, pass the `--verbose` flag to
KubeOne:

    ./kubeone --verbose install myconfig.yaml

This will stream the output of all shell commands to your shell.

## License

Apache License
