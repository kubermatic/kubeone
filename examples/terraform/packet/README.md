# Packet Quickstart Terraform scripts

The Packet Quickstart Terraform scripts can be used to create the needed infrastructure for a Kubernetes HA cluster.
Check out the following [Packet getting started walkthrough][packet-quickstart] to learn more about how to use the
scripts and how to provision a Kubernetes cluster using KubeOne.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_name | Name of the cluster | string | n/a | yes |
| control\_plane\_count | Number of master instances | string | `"3"` | no |
| device\_type | type (size) of the device | string | `"t1.small.x86"` | no |
| facility | Facility (datacenter) | string | `"ams1"` | no |
| operating\_system | Image to use for provisioning control plane | string | `"ubuntu_18_04"` | no |
| project\_id | project ID | string | n/a | yes |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port to be used to provision instances | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file used to access instances | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"root"` | no |
| workers\_operating\_system | Image to use for provisioning workers | string | `"ubuntu"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_api |  |
| kubeone\_hosts |  |
| kubeone\_workers |  |

[packet-quickstart]: https://github.com/kubermatic/kubeone/blob/master/docs/quickstart-packet.md
