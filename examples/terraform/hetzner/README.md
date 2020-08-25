# Hetzner Quickstart Terraform configs

The Hetzner Quickstart Terraform configs can be used to create the needed
infrastructure for a Kubernetes HA cluster. Check out the following
[Creating Infrastructure guide][docs-infrastructure] to learn more about how to
use the configs and how to provision a Kubernetes cluster using KubeOne.

## Kubernetes API Server Load Balancing

See the [Terraform loadbalancers in examples document][docs-tf-loadbalancer].

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/master/infrastructure/terraform_configs/
[docs-tf-loadbalancer]: https://docs.kubermatic.com/kubeone/master/advanced/example_loadbalancer/

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_name | prefix for cloud resources | string | n/a | yes |
| control\_plane\_type |  | string | `"cx21"` | no |
| datacenter |  | string | `"fsn1"` | no |
| image |  | string | `"ubuntu-18.04"` | no |
| lb\_type |  | string | `"lb11"` | no |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port to be used to provision instances | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file used to access instances | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"root"` | no |
| worker\_os | OS to run on worker machines | string | `"ubuntu"` | no |
| worker\_type |  | string | `"cx21"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_api | kube-apiserver LB endpoint |
| kubeone\_hosts | Control plane endpoints to SSH to |
| kubeone\_workers | Workers definitions, that will be transformed into MachineDeployment object |
