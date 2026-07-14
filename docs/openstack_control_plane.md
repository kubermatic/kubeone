# OpenStack Managed Control Plane

KubeOne can provision control-plane nodes directly on OpenStack using
machine-controller, eliminating the need to manually provision control-plane
VMs with Terraform. An Octavia load balancer handles kube-apiserver traffic.

When `cloudProvider.openstack.controlPlane` is configured in your
`kubeone.yaml`, KubeOne will:

1. Provision control-plane VMs via machine-controller's OpenStack driver,
   driven by the `controlPlane.nodeSets` spec.
2. Discover the pre-existing Octavia load balancer by name and set the
   `apiEndpoint` from its VIP address.
3. Register control-plane node IPs as members of the Octavia pool so
   kube-apiserver traffic reaches the nodes.

If `cloudProvider.openstack.controlPlane` is omitted, existing behaviour is
preserved — you must supply `apiEndpoint` and control-plane host IPs yourself
(e.g. via Terraform's `kubeone_hosts` and `kubeone_api` outputs).

**Important:** Unlike Hetzner, KubeOne does **not** create the Octavia load
balancer. The LB, its TCP listener (port 6443), pool, and health monitor must
exist before running `kubeone apply`. Create them via Terraform
(`examples/terraform/openstack/`) or any other tool.

## Prerequisites

- An existing Octavia load balancer with:
  - A TCP listener on port 6443
  - At least one pool (receiver) — members do not need to be pre-populated
  - A health monitor for the pool
- OpenStack credentials available via standard `OS_*` environment variables or
  a credentials file (see [Credentials](#credentials)).
- The Octavia load balancer name must match the configured
  `controlPlane.loadBalancer.name` (default: `<clusterName>-kube-apiserver`).

## Configuration

```yaml
apiVersion: kubeone.k8c.io/v1beta3
kind: KubeOneCluster
name: my-cluster

versions:
  kubernetes: 1.36.2

cloudProvider:
  openstack:
    controlPlane:
      loadBalancer:
        # Name of the pre-existing Octavia load balancer.
        # Default: "<CLUSTER_NAME>-kube-apiserver"
        name: my-cluster-kube-apiserver

        # Octavia pool ID. If empty, KubeOne discovers the pool
        # automatically from the load balancer name.
        # poolID: ""

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
        image: ubuntu-24.04
        flavor: m1.small
        securityGroups:
          - my-cluster-securitygroup
        network: my-cluster-network
        subnet: my-cluster-subnet
        configDrive: false
```

## Load Balancer Configuration

The `controlPlane.loadBalancer` section supports:

| Field | Default | Description |
|-------|---------|-------------|
| `name` | `<CLUSTER_NAME>-kube-apiserver` | Name of the pre-existing Octavia `loadbalancer_v2` resource |
| `poolID` | empty (auto-discovered) | Optional Octavia pool ID. When empty, KubeOne lists pools on the named LB and uses the first match |

### How the LB is discovered

During `kubeone apply`, KubeOne uses the `openstackLBClient()` function to
authenticate against the OpenStack API and:

1. Resolves the load balancer by name to obtain its `vip_address`, which
   becomes `apiEndpoint.host`.

2. Discovers the associated pool (unless `poolID` is explicitly set).

3. Registers each control-plane node's private (or public, as fallback) IP
   address as a member of the pool on port 6443. Only members not already
   registered are added — the operation is idempotent.

When the load balancer VIP is already known (e.g. from a previous
`kubeone apply`), the LB lookup step is skipped.

## Terraform Setup

The standard OpenStack Terraform configs (`examples/terraform/openstack/`)
create all required resources: network, subnet, router, security groups,
Octavia LB (listener, pool, health monitor), and optionally bastion.

When using managed control plane, the Terraform still creates the LB and
networking, but control-plane VMs are provisioned by KubeOne instead of
Terraform. The `kubeone_hosts` output from Terraform is **not** used in this
mode — KubeOne populates `controlPlane.hosts` automatically from the
provisioned VMs.

The LB name in Terraform must match the configured
`controlPlane.loadBalancer.name`. The default in both is
`<clusterName>-kube-apiserver`.

## Without Managed Control Plane

If you prefer to manage control-plane VMs with Terraform (or another tool),
omit the `controlPlane` section and use static hosts:

```yaml
cloudProvider:
  openstack: {}

apiEndpoint:
  host: 203.0.113.10
  port: 6443

controlPlane:
  hosts:
    - publicAddress: 10.0.0.1
      privateAddress: 10.0.0.1
      sshUsername: ubuntu
      sshPrivateKeyFile: /path/to/ssh-key
```

In this mode, no `controlPlane` field is needed on `cloudProvider.openstack`,
and `apiEndpoint.host` is required.

## Credentials

KubeOne reads OpenStack credentials from the standard `OS_*` environment
variables. Both password-based authentication and application credentials
are supported.

**Password authentication:**

| Variable | Required | Description |
|----------|----------|-------------|
| `OS_AUTH_URL` | Yes | Identity service endpoint (e.g. `https://identity.example.com/v3`) |
| `OS_USERNAME` | Yes | OpenStack username |
| `OS_PASSWORD` | Yes | OpenStack password |
| `OS_DOMAIN_NAME` | Yes | Domain name |
| `OS_TENANT_NAME` | No | Project/tenant name |
| `OS_TENANT_ID` | No | Project/tenant ID |
| `OS_REGION_NAME` | Yes | Region for the LB and compute services |

**Application credentials:**

| Variable | Required | Description |
|----------|----------|-------------|
| `OS_AUTH_URL` | Yes | Identity service endpoint |
| `OS_APPLICATION_CREDENTIAL_ID` | Yes | Application credential ID |
| `OS_APPLICATION_CREDENTIAL_SECRET` | Yes | Application credential secret |
| `OS_REGION_NAME` | Yes | Region for the LB and compute services |

Credentials can also be provided via a credentials file (`--credentials` flag).
The file must contain key-value pairs in the format `KEY=VALUE`, one per line.
