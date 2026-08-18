# AWS Managed Control Plane

KubeOne can provision control-plane nodes directly on AWS using machine-controller, eliminating the need for
pre-provisioned servers or external tooling (e.g. Terraform) to manage control-plane VMs.

When `cloudProvider.aws.controlPlane` is configured in your `kubeone.yaml`, KubeOne will:

1. Ensure an AWS Network Load Balancer (NLB), Target Group, and TCP listener exist for the kube-apiserver endpoint
   (creates them if missing) and set the `apiEndpoint` from the NLB's DNS name.
2. Provision control-plane EC2 instances via machine-controller's AWS driver, driven by the `controlPlane.nodeSets`
   spec.
3. Register each control-plane instance's private IP address as a target in the Target Group so the NLB routes traffic
   to it.

If `cloudProvider.aws.controlPlane` is omitted, existing behaviour is preserved — you must supply `apiEndpoint` and
control-plane host IPs yourself (e.g. via Terraform's `kubeone_hosts` and `kubeone_api` outputs).

## Prerequisites

- AWS credentials with permissions to manage EC2 instances and Elastic Load Balancing v2 resources (see
  [Credentials](#credentials) and [IAM permissions](#iam-permissions) below).
- An existing VPC with at least one subnet reachable by the control-plane nodes. All control-plane `nodeSets` must
  reference the same VPC.

## Configuration

```yaml
apiVersion: kubeone.k8c.io/v1beta3
kind: KubeOneCluster
name: my-cluster

versions:
  kubernetes: 1.36.2

cloudProvider:
  aws:
    # Region for the AWS resources (NLB, Target Group, EC2 instances). Required.
    region: eu-central-1

    controlPlane:
      loadBalancer:
        # Name of the load balancer to create. Default: "<CLUSTER_NAME>-kubeapi"
        name: my-cluster-kubeapi

        # Whether the load balancer should be internal (no public IP). Default: false
        internal: false

        # Optional security groups to attach to the load balancer
        securityGroupIDs:
          - sg-0123456789abcdef0

        # Optional tags applied to the load balancer and target group
        tags:
          env: production

controlPlane:
  nodeSets:
    - name: cp
      replicas: 3
      operatingSystem: ubuntu
      operatingSystemSpec:
        distUpgradeOnBoot: false
      ssh:
        publicKeys:
          - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA...
        username: ubuntu
      cloudProviderSpec:
        region: eu-central-1
        availabilityZone: eu-central-1a
        vpcId: vpc-0123456789abcdef0
        subnetId: subnet-0123456789abcdef0
        instanceType: t3.medium
        ami: ami-0123456789abcdef0
        diskSize: 50
        diskType: gp3
```

## Load Balancer Configuration

The `controlPlane.loadBalancer` section supports:

| Field | Default | Description |
|-------|---------|-------------|
| `name` | `<CLUSTER_NAME>-kubeapi` | Name of the NLB and derived Target Group. Limited to 32 characters by AWS. |
| `internal` | `false` | Whether the NLB should be internal (private) instead of internet-facing |
| `securityGroupIDs` | empty | Optional security groups to attach to the NLB |
| `tags` | empty | Optional tags applied to both the NLB and the Target Group |

### How the load balancer is provisioned

During `kubeone apply`, KubeOne uses the AWS ELBv2 API to:

1. Look up the named NLB. If it already exists, it is reused as-is and its DNS name becomes `apiEndpoint.host`.

2. If the NLB does not exist, KubeOne creates one with:
   - A TCP Target Group on port 6443, using IP-based targets (`target-type: ip`)
   - A TCP listener on port 6443 forwarding to the Target Group
   - Subnets derived from the VPC/subnet referenced by the control-plane `nodeSets`' `cloudProviderSpec`

3. As each control-plane instance is provisioned, KubeOne registers its private IP address as a Target Group target.

When the load balancer DNS name is already known (e.g. from a previous `kubeone apply`), the NLB creation step is
skipped entirely.

## Without Managed Control Plane

If you prefer to manage control-plane servers with Terraform (or another tool), omit the `controlPlane` section on `aws`
and use static hosts:

```yaml
cloudProvider:
  aws: {}

apiEndpoint:
  host: 203.0.113.10
  port: 6443

controlPlane:
  hosts:
    - publicAddress: 203.0.113.10
      privateAddress: 10.0.0.1
      sshUsername: ubuntu
      sshPrivateKeyFile: /path/to/ssh-key
```

In this mode, `region` can be omitted and `apiEndpoint.host` is required.

## Credentials

KubeOne reads AWS credentials from the credentials file (or environment) using the `AWS_ACCESS_KEY_ID` and
`AWS_SECRET_ACCESS_KEY` keys. If those are not set, KubeOne falls back to the shared AWS config/credentials files using
the profile named by the `AWS_PROFILE` environment variable.

```ini
AWS_ACCESS_KEY_ID=<your-access-key-id>
AWS_SECRET_ACCESS_KEY=<your-secret-access-key>
```

Pass the credentials file to KubeOne via the `--credentials` flag:

```bash
kubeone apply --manifest kubeone.yaml --credentials credentials.yaml
```

### IAM permissions

The credentials used by KubeOne need permissions to manage EC2 instances and
Elastic Load Balancing v2 resources, for example:

- `ec2:RunInstances`, `ec2:DescribeInstances`, `ec2:TerminateInstances`, `ec2:CreateTags`, `ec2:DescribeSubnets`,
  `ec2:DescribeSecurityGroups`
- `elasticloadbalancing:CreateLoadBalancer`, `DescribeLoadBalancers`, `CreateTargetGroup`, `DescribeTargetGroups`,
  `CreateListener`, `RegisterTargets`, `DescribeTargetHealth`
