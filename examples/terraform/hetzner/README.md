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

