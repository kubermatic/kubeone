# OpenStack

## Environment variables

So Terraform can communicate with the OpenStack API it is necessary that the required environment variables are set correctly.
Example
```bash
export OS_AUTH_URL="https://some-keystone-endpoint:5000/v3"
export OS_IDENTITY_API_VERSION=3
export OS_USERNAME="some-username"
export OS_PASSWORD="some-password"
export OS_REGION_NAME="region1"
export OS_INTERFACE=public
export OS_ENDPOINT_TYPE=public
export OS_USER_DOMAIN_NAME="Default"
export OS_PROJECT_ID="some-project-id"
```

## Terraform

```
$ terraform init
$ terraform plan
$ terraform apply
$ terraform output -json > tf.json
```

## KubeOne

Create a config.yaml and make sure you specify `.provider.cloud_config`.

```bash
kubeone install --tfjson tf.json config.yaml
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_name | prefix for cloud resources | string | n/a | yes |
| control\_plane\_count | Number of instances | string | `"3"` | no |
| control\_plane\_flavor | OpenStack instance flavor for the control plane nodes | string | `"m1.small"` | no |
| external\_network\_name | OpenStack external network name | string | n/a | yes |
| image | OpenStack image for the control plane nodes | string | `"Ubuntu 18.04 LTS"` | no |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file, only specify in absence of SSH agent | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"ubuntu"` | no |
| subnet\_cidr | OpenStack subnet cidr | string | `"192.168.1.0/24"` | no |
| subnet\_dns\_servers |  | list | `<list>` | no |
| worker\_flavor | OpenStack instance flavor for the worker nodes | string | `"m1.small"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_api |  |
| kubeone\_hosts |  |
| kubeone\_workers |  |

