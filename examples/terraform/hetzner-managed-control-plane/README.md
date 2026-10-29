# Hetzner Managed Control Plane Terraform configs

The Hetzner Managed Control Plane Terraform configs provision the supporting
infrastructure for a Kubernetes HA cluster whose control plane is created and
managed directly by KubeOne (rather than by Terraform). They create the private
network, its subnet and a firewall for the control-plane and worker nodes, but
intentionally do **not** create any control-plane servers or a load balancer —
KubeOne provisions those itself.

Check out the [Creating Infrastructure guide][docs-infrastructure] to learn more
about how to use the configs and how to provision a Kubernetes cluster using
KubeOne.

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/main/guides/using-terraform-configs/

## Managed Control Plane

Instead of provisioning control-plane servers and an API load balancer with
Terraform (as the regular `hetzner` example does), this config only lays the
groundwork and lets KubeOne provision and manage the control plane directly via
`cloudProvider.hetzner.controlPlane` and `controlPlane.nodeSets` in
`kubeone.yaml`.

When using the managed control plane:

- Set `cloudProvider.hetzner.networkID` to the `networking.networkName` output.
  KubeOne looks the network up by name and attaches its load balancer to it.
  The load balancer location (`nbg1` by default) must be in the network zone of
  the subnet (`eu-central` by default).
- Reference the same network in each `nodeSets[].cloudProviderSpec.networks`
  and in the `networks` of the worker pools, the load balancer reaches the
  control plane servers over their private addresses.
- The firewall applies to all the servers labeled
  `kubeone_cluster_name=<name>`. KubeOne labels the control plane servers with
  it, add it to the `labels` of the worker pools as well.

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0.0 |
| hcloud | ~> 1.52 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| name | common name for the resources, should be the name of the cluster | `string` | n/a | yes |
| base\_network\_cidr | base cidr, resulting cidr is randomly generated depending on provided subnet\_mask | `string` | `"10.100.0.0/16"` | no |
| ip\_range | ip range to use for private network | `string` | `""` | no |
| network\_zone | network zone to use for private network, the servers and the load balancer must be in a location of this zone | `string` | `"eu-central"` | no |
| ssh\_port | SSH port to be used to provision instances | `number` | `22` | no |
| subnet\_mask | subnet mask to use for generating cidr for a private network | `number` | `24` | no |

## Outputs

| Name | Description |
|------|-------------|
| networking | n/a |
