# DigitalOcean Managed Control Plane

KubeOne can provision control-plane nodes directly on DigitalOcean using machine-controller, eliminating the need for
pre-provisioned droplets or external tooling (e.g. Terraform) to manage control-plane VMs.

When `controlPlane.nodeSets` is configured together with `cloudProvider.digitalocean` in your `kubeone.yaml`, KubeOne
will:

1. Ensure a DigitalOcean load balancer exists for the kube-apiserver endpoint (creates one if missing, or uses an
   existing one referenced by ID) and set the `apiEndpoint` from its IP address.
2. Provision control-plane droplets via machine-controller's DigitalOcean driver, driven by the `controlPlane.nodeSets`
   spec.
3. Tag each control-plane droplet with `kubeone-cluster:<CLUSTER_NAME>`. The load balancer selects its backends by this
   tag, so droplets are added as targets without KubeOne calling the member registration API.

If `controlPlane.nodeSets` is omitted, existing behaviour is preserved — you must supply `apiEndpoint` and
control-plane host IPs yourself (e.g. via Terraform's `kubeone_hosts` and `kubeone_api` outputs).

## Prerequisites

- A DigitalOcean API token with full access, or with the custom scopes listed under [Credentials](#credentials) below.
- No other infrastructure is required. Droplets and the load balancer are created in the default VPC of the region,
  with private networking enabled.
- Control-plane droplets are named `<CLUSTER_NAME>-<NODESET>-<INDEX>`, which becomes their hostname, so it must be at
  most 63 characters (e.g. a cluster name of up to 58 characters with a nodeSet `cp` of up to 10 replicas).

## Configuration

```yaml
apiVersion: kubeone.k8c.io/v1beta2
kind: KubeOneCluster
name: my-cluster

versions:
  kubernetes: 1.36.2

cloudProvider:
  external: true
  digitalocean:
    controlPlane:
      loadBalancer:
        # Name of the load balancer to create. Default: "<CLUSTER_NAME>-kubeapi"
        name: my-cluster-kubeapi

        # Region of the load balancer, must match the nodeSets region.
        # Default: region of the first nodeSet
        region: fra1

        # Number of load balancer nodes (1-100). Only used when the load
        # balancer is created. Default: 1
        sizeUnit: 1

        # Use a pre-existing load balancer instead of creating one. It's never
        # deleted by KubeOne, and must health check TCP on port 6443. When set, name and sizeUnit aren't used, but are
        # still validated (name at most 255 characters, sizeUnit at most 100).
        # id: 4de7ac8b-495b-4884-9a69-1050c6793cd6

controlPlane:
  nodeSets:
    - name: cp
      replicas: 3
      # ubuntu or rockylinux, the operating systems supported by the
      # machine-controller DigitalOcean driver
      operatingSystem: ubuntu
      ssh:
        publicKeys:
          - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA...
        username: root
      cloudProviderSpec:
        region: fra1
        size: s-2vcpu-4gb
        # Default: true, must not be false
        private_networking: true
        tags:
          - env-production
```

`cloudProviderSpec` is the machine-controller DigitalOcean provider spec and is decoded strictly: unknown fields are
rejected. The `token` field must not be set, KubeOne passes the token from the credentials. All nodeSets must use the
same region.

The region can't be changed after the droplets are created, DigitalOcean can't move droplets between regions. Before
looking up or provisioning the control plane, KubeOne lists the droplets tagged `kubeone-owned:<CLUSTER_NAME>` and fails
if any of them is in another region than the nodeSets. To move a cluster to another region, reset it and apply it again.

The cluster name identifies the control-plane droplets and the load balancer, so it must be unique per DigitalOcean
account, also across regions.

## Load Balancer Configuration

The `controlPlane.loadBalancer` section supports:

| Field | Default | Description |
|-------|---------|-------------|
| `id` | empty | ID of a pre-existing load balancer to use. KubeOne never creates or deletes it |
| `name` | `<CLUSTER_NAME>-kubeapi` | Name of the load balancer to create |
| `region` | region of the first nodeSet | Region of the load balancer, must match the nodeSets region |
| `sizeUnit` | `1` | Number of load balancer nodes (1-100). Only applied when the load balancer is created; resize an existing one in the DigitalOcean console |

If `apiEndpoint.host` is set, KubeOne doesn't create or look up any load balancer and uses the given endpoint as-is
(e.g. a load balancer or DNS name managed outside of KubeOne). On reset, a load balancer created by KubeOne is only
deleted if `apiEndpoint.host` is its IP address.

### How the load balancer is provisioned

During `kubeone apply`, KubeOne uses the DigitalOcean API to:

1. Look up the named load balancer in the nodeSets region. DigitalOcean doesn't enforce unique names; if several load
   balancers match, the one targeting `kubeone-cluster:<CLUSTER_NAME>` is used.

2. If the load balancer exists, it's reused only if it targets droplets tagged `kubeone-cluster:<CLUSTER_NAME>`.
   Otherwise `kubeone apply` fails, because the load balancer wasn't created by KubeOne for this cluster or was
   modified (e.g. switched to a list of droplets). Choose a different `name`, or restore the load balancer to target
   the cluster tag. A missing kube-apiserver forwarding rule (TCP 6443) and a changed health check are restored; if port
   6443 was changed to forward elsewhere, `kubeone apply` fails, as DigitalOcean allows only one rule per entry port.

3. If the load balancer does not exist, KubeOne creates one with:
   - A TCP forwarding rule on port 6443
   - A TCP health check on port 6443
   - Backends selected by the `kubeone-cluster:<CLUSTER_NAME>` droplet tag
   - The default VPC of the region, the same VPC machine-controller creates the droplets in

4. Wait for the load balancer to become active. DigitalOcean assigns the IP address only then, which usually takes a few
   minutes. The IP address becomes `apiEndpoint.host`.

5. Provision the control-plane droplets with the cluster tag, so the load balancer picks them up as targets.

Don't share this load balancer with other VMs: all forwarding rules of a DigitalOcean load balancer use the same
backends, and `kubeone reset --destroy-control-plane-load-balancer` deletes it. Use a separate load balancer, or a
`Service` of type `LoadBalancer` for workloads running in the cluster.

### Using an existing load balancer

When `loadBalancer.id` is set, KubeOne gets the load balancer by ID and checks that it's:

- A regional load balancer. Global load balancers only forward HTTP/HTTPS, kube-apiserver needs TCP on port 6443.
- In the nodeSets region and in that region's default VPC, where the droplets are created. A load balancer only reaches
  droplets in its own VPC.
- Not targeting other droplets or tags. A load balancer without backends is configured to target droplets tagged
  `kubeone-cluster:<CLUSTER_NAME>`.
- Not forwarding port 6443 to another port or protocol.
- Health checking kube-apiserver: TCP on port 6443, or HTTPS on port 6443 with path `/readyz`, `/livez` or `/healthz`.
  KubeOne never changes the health check, so configure it before running `kubeone apply`. The DigitalOcean default
  checks HTTP on port 80, which control-plane droplets don't serve, and is rejected.

KubeOne then adds the TCP 6443 forwarding rule if it's missing.

The load balancer is never deleted by KubeOne. Once configured, it targets the `kubeone-cluster:<CLUSTER_NAME>` tag.
`kubeone reset --destroy-control-plane --destroy-control-plane-load-balancer` removes that backend selector again,
leaving the load balancer without backends; the forwarding rule and the health check are kept. Don't replace `id` with a
`name` pointing to this load balancer before resetting the cluster: while it targets the cluster tag, KubeOne treats it as
its own and deletes it on reset.

### Tag-based target selection

DigitalOcean doesn't persist tags on load balancers, so a load balancer is considered created by KubeOne when its
backends are selected by the `kubeone-cluster:<CLUSTER_NAME>` tag. The following tags are automatically applied to each
provisioned control-plane droplet, in addition to the tags from `cloudProviderSpec.tags`:

| Tag | Purpose |
|-----|---------|
| `kubeone-cluster:<CLUSTER_NAME>` | Selects the droplet as load balancer target |
| `kubeone-owned:<CLUSTER_NAME>` | Marks the droplet as created by KubeOne |
| `kubeone_cluster_name:<CLUSTER_NAME>` | Identifies the cluster (same key other providers use) |
| `kubeone_role:control-plane` | Identifies the node role (same key other providers use) |

Characters that DigitalOcean doesn't allow in tags (e.g. `.`) are replaced with `_`.

During `kubeone reset --destroy-control-plane --destroy-control-plane-load-balancer`, the load balancer is deleted only
if it targets the cluster tag, isn't referenced by `id`, and `apiEndpoint.host` is unset or is the load balancer IP. A
load balancer referenced by `id` only stops targeting the cluster tag, see
[Using an existing load balancer](#using-an-existing-load-balancer).
Control-plane droplets are found by name and machine UID tag (both `<CLUSTER_NAME>-<NODESET>-<INDEX>`, the tag with
invalid characters replaced) in all regions, so the cluster `name` and the nodeSet names must not change between `apply`
and `reset`. A changed region is detected and refused, see [Configuration](#configuration).

KubeOne doesn't create a DigitalOcean cloud firewall. If you add one, the control-plane droplets must allow SSH (22)
from the machine running KubeOne, kube-apiserver (6443) from the load balancer and nodes, and the ports required by
Kubernetes and the CNI between the nodes.

## Without Managed Control Plane

If you prefer to manage control-plane droplets with Terraform (or another tool), omit `controlPlane.nodeSets` and use
static hosts:

```yaml
cloudProvider:
  digitalocean: {}

apiEndpoint:
  host: 203.0.113.10
  port: 6443

controlPlane:
  hosts:
    - publicAddress: 203.0.113.10
      privateAddress: 10.0.0.1
      sshUsername: root
      sshPrivateKeyFile: /path/to/ssh-key
```

In this mode, `cloudProvider.digitalocean.controlPlane` is not used and `apiEndpoint.host` is required.

## Credentials

KubeOne reads the DigitalOcean API token from the credentials file using the `DIGITALOCEAN_TOKEN` key. Use a token
with full access, or a custom-scoped token with:

| Resource | Scopes | Used for |
|----------|--------|----------|
| `droplet` | create, read, update, delete | Control-plane droplets |
| `load_balancer` | create, read, update, delete | kube-apiserver load balancer |
| `tag` | create, read, delete | Droplet tags, load balancer backend selection |
| `ssh_key` | create, read, delete | machine-controller uploads a temporary SSH key for each droplet it creates |
| `vpc` | read | Default VPC check for a load balancer referenced by `id` |
| `regions`, `sizes`, `image` | read | Droplet creation |

```yaml
DIGITALOCEAN_TOKEN: <your-api-token>
```

The `DIGITALOCEAN_TOKEN` environment variable is also supported for convenience during development. For
machine-controller integration, KubeOne maps the token to the `DO_TOKEN` environment variable automatically.

Pass the credentials file to KubeOne via the `--credentials` flag:

```bash
kubeone apply --manifest kubeone.yaml --credentials credentials.yaml
```
