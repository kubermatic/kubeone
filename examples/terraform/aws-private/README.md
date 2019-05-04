# AWS with Private VPC Quickstart Terraform scripts

:warning: **Experimental: The following setup is experimental and can't be used with KubeOne out-of-box.
Follow the [issue #337](https://github.com/kubermatic/kubeone/issues/337) for more details about progress.** :warning:

## Assumptions

* VPC (default or custom) with attached internet gateway exists

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| aws\_region | AWS region to speak to | string | `"eu-west-3"` | no |
| cluster\_name | prefix for cloud resources | string | n/a | yes |
| control\_plane\_type | AWS instance type | string | `"t3.medium"` | no |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file, only specify in absence of SSH agent | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"ubuntu"` | no |
| subnet\_netmask\_bits | default 8 bits in /16 CIDR, makes it /24 subnetworks | string | `"8"` | no |
| subnet\_offset | subnet offset (from main VPC cidr_block) number to be cut | string | `"0"` | no |
| vpc\_id | VPC to use ('default' for default VPC) | string | `"default"` | no |
| worker\_type | instance type for workers | string | `"t3.medium"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_api |  |
| kubeone\_bastion |  |
| kubeone\_hosts |  |
| kubeone\_workers |  |

