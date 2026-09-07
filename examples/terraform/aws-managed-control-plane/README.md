# AWS Managed Control Plane Terraform configs

The AWS Managed Control Plane Terraform configs provision the supporting
infrastructure for a Kubernetes HA cluster whose control plane is created and
managed directly by KubeOne (rather than by Terraform). They create the VPC,
public subnet, security groups, and IAM instance profile/role needed by the
control-plane and worker nodes, but intentionally do **not** create any
control-plane instances or a load balancer — KubeOne provisions those itself.

Check out the [Creating Infrastructure guide][docs-infrastructure] to learn more
about how to use the configs and how to provision a Kubernetes cluster using
KubeOne.

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/main/guides/using-terraform-configs/

## Managed Control Plane

Instead of provisioning control-plane instances and an API load balancer with
Terraform (as the regular `aws` example does), this config only lays the
groundwork and lets KubeOne provision and manage the control plane directly via
`cloudProvider.aws.controlPlane` and `controlPlane.nodeSets` in `kubeone.yaml`.
See [docs/aws_control_plane.md](../../../docs/aws_control_plane.md) for details.

When using the managed control plane:

- Set `cloudProvider.aws.region` and `cloudProvider.aws.controlPlane.loadBalancer`
  in `kubeone.yaml`. KubeOne creates its own Network Load Balancer, so there is
  no `kubeone_api`/`kubeone_hosts` Terraform output to consume.
- The VPC, subnet, and security group resources created here
  (`data.aws_vpc.selected`, `aws_subnet.public`, `aws_security_group.common`)
  are referenced in each `nodeSets[].cloudProviderSpec` (`vpcId`, `subnetId`,
  `securityGroupIDs`).
- The `vm`, `networking`, and `loadbalancer` outputs expose the AMI ID, IAM
  instance profile, region, VPC/subnet IDs, and security group IDs that should
  be used when writing the `controlPlane.nodeSets` in `kubeone.yaml`.
