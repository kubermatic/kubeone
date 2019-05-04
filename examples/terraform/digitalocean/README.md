# DigitalOcean Quickstart Terraform scripts

The DigitalOcean Quickstart Terraform scripts can be used to create the needed infrastructure for a Kubernetes HA cluster.
Check out the following [DigitalOcean getting started walkthrough][do-quickstart] to learn more about how to use the
scripts and how to provision a Kubernetes cluster using KubeOne.

[do-quickstart]: https://github.com/kubermatic/kubeone/blob/master/docs/quickstart-digitalocean.md

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_name | Name of the cluster | string | n/a | yes |
| control\_plane\_count | Number of master instances | string | `"3"` | no |
| droplet\_image | Image to use for provisioning droplet | string | `"ubuntu-18-04-x64"` | no |
| droplet\_ipv6 | Enable IPv6 | string | `"false"` | no |
| droplet\_monitoring | Enable advance Droplet metrics | string | `"false"` | no |
| droplet\_private\_networking | Enable Private Networking on Droplets (recommended) | string | `"true"` | no |
| droplet\_size | Size of Droplets | string | `"s-2vcpu-4gb"` | no |
| region | Region to speak to | string | `"fra1"` | no |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port to be used to provision instances | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file used to access instances | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"root"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_api |  |
| kubeone\_hosts |  |
| kubeone\_workers |  |
