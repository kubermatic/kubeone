# AWS Quickstart Terraform scripts

The AWS Quickstart Terraform scripts can be used to create the needed infrastructure for a Kubernetes HA cluster.
Check out the following [AWS getting started walkthrough][aws-quickstart] to learn more about how to use the
scripts and how to provision a Kubernetes cluster using KubeOne.

[aws-quickstart]: https://github.com/kubermatic/kubeone/blob/master/docs/quickstart-aws.md

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| aws\_region | AWS region to speak to | string | `"eu-west-3"` | no |
| cluster\_name | prefix for cloud resources | string | n/a | yes |
| control\_plane\_count | Number of instances | string | `"3"` | no |
| control\_plane\_type | AWS instance type | string | `"t3.medium"` | no |
| control\_plane\_volume\_size | Size of the EBS volume, in Gb | string | `"100"` | no |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file, only specify in absence of SSH agent | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"ubuntu"` | no |
| vpc\_id | VPC to use ('default' for default VPC) | string | `"default"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_api |  |
| kubeone\_hosts |  |
| kubeone\_workers |  |
