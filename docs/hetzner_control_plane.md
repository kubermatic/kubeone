# Hetzner Managed Control Plane

KubeOne can provision control-plane nodes directly on Hetzner Cloud using
machine-controller, eliminating the need for pre-provisioned servers or
external tooling (e.g. Terraform) to manage control-plane VMs.

When `cloudProvider.hetzner.controlPlane` is configured in your
`kubeone.yaml`, KubeOne will:

1. Ensure a Hetzner Cloud load balancer exists (creates one if missing) and
   set the `apiEndpoint` from its public IPv4 address.
2. Provision control-plane servers via machine-controller's Hetzner driver,
   driven by the `controlPlane.nodeSets` spec.
3. Automatically attach the load balancer to the configured private network
   and use label-based target selection to route traffic to control-plane
   nodes.

If `cloudProvider.hetzner.controlPlane` is omitted, existing behaviour is
preserved — you must supply `apiEndpoint` and control-plane host IPs yourself
(e.g. via Terraform's `kubeone_hosts` and `kubeone_api` outputs).

## Prerequisites

- A Hetzner Cloud API token with read/write access to servers, networks,
  and load balancers.
- An existing private network in Hetzner Cloud. The network name (or ID)
  must be provided via `cloudProvider.hetzner.networkID`.
- Control-plane nodes must be reachable via the private network for
  kube-apiserver traffic.

## Configuration

```yaml
apiVersion: kubeone.k8c.io/v1beta3
kind: KubeOneCluster
name: my-cluster

versions:
  kubernetes: 1.36.2

cloudProvider:
  hetzner:
    # Private network name or ID. Required when controlPlane is set.
    networkID: my-private-network

    controlPlane:
      loadBalancer:
        # Name of the load balancer to create. Default: "<CLUSTER_NAME>-kubeapi"
        name: my-cluster-kubeapi

        # Load balancer type. Default: "lb11"
        type: lb11

        # Data center location. Default: "nbg1"
        location: nbg1

        # Assign a public IP to the load balancer. Default: true
        publicIP: true

        # Optional labels applied to the load balancer
        labels:
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
        serverType: cx22
        location: nbg1
        image: ubuntu-24.04
        networks:
          - my-private-network
```

## Load Balancer Configuration

The `controlPlane.loadBalancer` section supports:

| Field | Default | Description |
|-------|---------|-------------|
| `name` | `<CLUSTER_NAME>-kubeapi` | Name of the Hetzner load balancer |
| `type` | `lb11` | Hetzner load balancer type (size) |
| `location` | `nbg1` | Data center location for the load balancer |
| `publicIP` | `true` | Whether to assign a public IPv4 address |
| `labels` | empty | Optional labels applied to the load balancer resource |

### How the load balancer is provisioned

During `kubeone apply`, KubeOne uses the Hetzner Cloud API to:

1. Look up the named load balancer. If it already exists, it is reused as-is
   and its public IPv4 address becomes `apiEndpoint.host`.

2. If the load balancer does not exist, KubeOne creates one with:
   - A TCP listener on port 6443
   - Label-based target selection using `kubeone_cluster_name` and
     `kubeone_role` tags (automatically applied to both the LB and VMs)
   - One service per target (port 6443, private-network IP)
   - Attached to the private network specified in `networkID`
   - A health check on port 6443

3. KubeOne provisions control-plane servers with matching labels so the load
   balancer automatically picks them up as targets.

When the load balancer IP is already known (e.g. from a previous
`kubeone apply`), the LB creation step is skipped entirely.

### Label-based target selection

The load balancer uses Hetzner's label-selector feature to target
control-plane servers. The following labels are automatically applied to
both the load balancer and each provisioned server:

| Label | Value | Purpose |
|-------|-------|---------|
| `kubeone_cluster_name` | Cluster name | Identifies the cluster |
| `kubeone_role` | `control-plane` / `api` | Identifies node role |
| `kubeone_own_since_timestamp` | Unix timestamp | Tracks when the resource was created |

This means newly provisioned servers are automatically added as load
balancer targets without KubeOne needing to call the member registration API.

## Without Managed Control Plane

If you prefer to manage control-plane servers with Terraform (or another
tool), omit the `controlPlane` section on `hetzner` and use static hosts:

```yaml
cloudProvider:
  hetzner:
    networkID: my-private-network

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

In this mode, `networkID` can be omitted and `apiEndpoint.host` is required.

## Credentials

KubeOne reads the Hetzner Cloud API token from the credentials file using
the `HCLOUD_TOKEN` key. The token must have read/write access to servers,
networks, and load balancers.

```ini
HCLOUD_TOKEN=<your-api-token>
```

The `HCLOUD_TOKEN` environment variable is also supported for convenience
during development.

Pass the credentials file to KubeOne via the `--credentials` flag:

```bash
kubeone apply --manifest kubeone.yaml --credentials credentials.yaml
```
