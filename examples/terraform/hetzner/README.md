# Hetzner Quickstart Terraform scripts

The Hetzner Quickstart Terraform scripts can be used to create the needed infrastructure for a Kubernetes HA cluster.
Check out the following [Hetzner getting started walkthrough][hetzner-quickstart] to learn more about how to use the
scripts and how to provision a Kubernetes cluster using KubeOne.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_name | prefix for cloud resources | string | n/a | yes |
| control\_plane\_count | Number of instances | string | `"3"` | no |
| control\_plane\_type |  | string | `"cx21"` | no |
| datacenter |  | string | `"fsn1"` | no |
| image |  | string | `"ubuntu-18.04"` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| worker\_type |  | string | `"cx21"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_hosts |  |
| kubeone\_workers |  |

[hetzner-quickstart]: https://github.com/kubermatic/kubeone/blob/master/docs/quickstart-hetzner.md
