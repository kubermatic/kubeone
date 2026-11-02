# OpenStack Managed Control Plane Terraform configs

The OpenStack Managed Control Plane Terraform configs provision the supporting
infrastructure for a Kubernetes HA cluster whose control plane is created and
managed directly by KubeOne (rather than by Terraform). They create the
network, subnet, router, security group, a bastion host, and the Octavia load
balancer (listener, pool, and health monitor) for kube-apiserver, but
intentionally do **not** create any control-plane instances. KubeOne provisions
those itself and adds them to the load balancer pool.

Check out the [Creating Infrastructure guide][docs-infrastructure] to learn more
about how to use the configs and how to provision a Kubernetes cluster using
KubeOne.

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/main/guides/using-terraform-configs/

## Managed Control Plane

Instead of provisioning control-plane instances with Terraform (as the regular
`openstack` example does), this config only lays the groundwork and lets
KubeOne provision and manage the control plane directly via
`cloudProvider.openstack.controlPlane` and `controlPlane.nodeSets` in
`kubeone.yaml`. See
[docs/openstack_control_plane.md](../../../docs/openstack_control_plane.md) for
details.

When using the managed control plane:

- Set `name` to the cluster name, so the load balancer is named
  `<name>-kube-apiserver`, which is the default
  `cloudProvider.openstack.controlPlane.loadBalancer.name` of KubeOne. Otherwise
  set that field to the `loadbalancer.name` output.
- KubeOne uses the private VIP of the load balancer as the API endpoint. It
  reaches kube-apiserver through an SSH tunnel, use `kubeone proxy` to reach it
  from outside of the network.
- The control-plane VMs have no floating IPs, set the `bastion` output as
  `nodeSets[].ssh.bastion` (and `bastionPort`, `bastionUser`).
- The `vm` and `networking` outputs expose the image name, the network and
  subnet IDs, and the security group name that should be used in each
  `nodeSets[].cloudProviderSpec` (`image`, `network`, `subnet`,
  `securityGroups`). The subnet has to be set by ID, KubeOne adds the
  control-plane VMs on it to the load balancer pool.
