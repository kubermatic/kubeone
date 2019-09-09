# AWS with Private VPC Quickstart Terraform scripts

:warning: **Experimental: The following setup is experimental and can't be used with KubeOne out-of-box.
Follow the [issue #337](https://github.com/kubermatic/kubeone/issues/337) for more details about progress.** :warning:

## Assumptions

* VPC (default or custom) with attached internet gateway exists

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| ami | AMI ID, use it to fixate control-plane AMI in order to avoid force-recreation it at later times | string | `""` | no |
| aws\_region | AWS region to speak to | string | `"eu-west-3"` | no |
| cluster\_name | Name of the cluster | string | n/a | yes |
| control\_plane\_type | AWS instance type | string | `"t3.medium"` | no |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port to be used to provision instances | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file used to access instances | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"ubuntu"` | no |
| bastion\_port | Bastion SSH port | string | `"22"` | no |
| bastion\_user | Bastion SSH username | string | `"ubuntu"` | no |
| subnet\_netmask\_bits | default 8 bits in /16 CIDR, makes it /24 subnetworks | string | `"8"` | no |
| subnet\_offset | subnet offset (from main VPC cidr_block) number to be cut | string | `"0"` | no |
| vpc\_id | VPC to use ('default' for default VPC) | string | `"default"` | no |
| worker\_os | OS to run on worker machines | string | `"ubuntu"` | no |
| worker\_type | instance type for workers | string | `"t3.medium"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_api | kube-apiserver LB endpoint |
| kubeone\_bastion |  |
| kubeone\_hosts | Control plane endpoints to SSH to |
| kubeone\_workers | Workers definitions, that will be transformed into MachineDeployment object |
