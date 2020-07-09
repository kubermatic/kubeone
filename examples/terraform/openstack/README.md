# OpenStack Quickstart Terraform scripts

The OpenStack Quickstart Terraform scripts can be used to create the needed infrastructure for a Kubernetes HA cluster.
Check out the following [OpenStack getting started walkthrough](https://docs.kubermatic.com/kubeone/master/getting_started/openstack) to learn more about how to use the
scripts and how to provision a Kubernetes cluster using KubeOne.

## Kubernetes APIserver LoadBalancing

See [example-loadbalancer](https://docs.kubermatic.com/kubeone/master/using_kubeone/example_loadbalancer)

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_name | Name of the cluster | string | n/a | yes |
| control\_plane\_flavor | OpenStack instance flavor for the control plane nodes | string | `"m1.small"` | no |
| external\_network\_name | OpenStack external network name | string | n/a | yes |
| image | image name to use | string | `"Ubuntu 18.04"` | no |
| lb\_flavor | OpenStack instance flavor for the LoadBalancer node | string | `"m1.micro"` | no |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port to be used to provision instances | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file used to access instances | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"root"` | no |
| subnet\_cidr | OpenStack subnet cidr | string | `"192.168.1.0/24"` | no |
| subnet\_dns\_servers |  | list | `<list>` | no |
| worker\_flavor | OpenStack instance flavor for the worker nodes | string | `"m1.small"` | no |
| worker\_os | OS to run on worker machines | string | `"ubuntu"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_api | kube-apiserver LB endpoint |
| kubeone\_hosts | Control plane endpoints to SSH to |
| kubeone\_workers | Workers definitions, that will be transformed into MachineDeployment object |
