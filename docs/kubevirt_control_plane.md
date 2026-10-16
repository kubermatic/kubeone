# KubeVirt Managed Control Plane

KubeOne can provision control-plane nodes directly on a KubeVirt infrastructure
cluster, eliminating the need for external tooling (e.g. Terraform) to manage
control-plane VMs.

When `cloudProvider.kubevirt.controlPlane` is configured in your `kubeone.yaml`,
KubeOne will:

1. Create a Kubernetes `Service` (type `LoadBalancer` or `NodePort`) in the
   infra cluster to serve as the kube-apiserver endpoint. The `apiEndpoint` is
   populated automatically from this service.
2. Provision control-plane `VirtualMachine` resources in the infra cluster via
   the machine-controller provider, driven by the `controlPlane.nodeSets` spec.
3. Resolve SSH connectivity using the VM's in-cluster IP (KubeVirt VMs only
   receive an internal address).

If `cloudProvider.kubevirt.controlPlane` is omitted, existing behaviour is
preserved — you must supply `apiEndpoint` and control-plane host IPs yourself.

## Prerequisites

- A running KubeVirt infrastructure cluster.
- The infra cluster kubeconfig, provided via the `KUBEVIRT_KUBECONFIG`
  environment variable or base64-encoded in the credentials file.
- An `infraNamespace` where KubeOne will create resources (VMs, services, etc.).

### Infra kubeconfig permissions

KubeOne creates least-privilege identities in `infraNamespace` for the
components that talk to the infra cluster (machine-controller, CCM, CSI
controller); see the
[proposal](proposals/20261015-kubevirt-infra-side-components.md).

The infra kubeconfig must therefore be allowed to create, update, list and
delete the following objects in `infraNamespace`:

- `ServiceAccounts`, `Secrets` and `ConfigMaps`
- `Roles` and `RoleBindings`
- `Deployments`
- `Services`

Kubernetes RBAC only lets a user grant permissions they hold themselves, so the
infra kubeconfig must also hold every permission granted to the components'
Roles (KubeVirt `VirtualMachines`, `VirtualMachineInstances` and their
subresources, CDI `DataVolumes`, `PersistentVolumeClaims`, `VolumeSnapshots`),
or be granted the `escalate` and `bind` verbs on `roles` in `infraNamespace`.

Several KubeOne clusters can share one `infraNamespace`. Every object KubeOne
creates there is prefixed with the cluster name and labelled
`kubeone_cluster_name=<cluster name>`.

## Configuration

```yaml
apiVersion: kubeone.k8c.io/v1beta3
kind: KubeOneCluster
name: my-kubevirt-cluster

versions:
  kubernetes: 1.36.2

cloudProvider:
  kubevirt:
    infraNamespace: mycluster
    controlPlane:
      loadBalancer:
        # Name of the Service to create. Default: "<CLUSTER_NAME>-kubeapi"
        name: my-cluster-kubeapi

        # Service type: "LoadBalancer" (default) or "NodePort"
        serviceType: LoadBalancer

        # Optional annotations on the Service
        annotations:
          metallb.universe.tf/address-pool: public

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
        flavor: "4"
        sourceURL: "https://some-storage/path/ubuntu.qcow2"
```

## Load Balancer Configuration

The `controlPlane.loadBalancer` section supports:

| Field | Default | Description |
|-------|---------|-------------|
| `name` | `<CLUSTER_NAME>-kubeapi` | Name of the Kubernetes `Service` to create |
| `serviceType` | `LoadBalancer` | Either `LoadBalancer` or `NodePort` |
| `annotations` | empty | Annotations applied to the Service |

### LoadBalancer Mode

When `serviceType: LoadBalancer`, KubeOne waits for the load balancer ingress
address to be assigned and uses it as the `apiEndpoint`. This mode requires a
load balancer provisioner in the infra cluster (e.g., MetalLB).

### NodePort Mode

When `serviceType: NodePort`, KubeOne uses an infra cluster node's external IP
(or internal IP as fallback) combined with the allocated `NodePort` as the
`apiEndpoint`.

## Without Managed Control Plane

If you prefer to manage control-plane VMs externally (e.g., via Terraform),
omit the `controlPlane` section:

```yaml
cloudProvider:
  kubevirt:
    infraNamespace: kubeone-clusters

apiEndpoint:
  host: 192.0.2.10
  port: 6443

controlPlane:
  hosts:
    - publicAddress: 192.0.2.10
      privateAddress: 10.0.0.10
      sshUsername: ubuntu
      sshPrivateKeyFile: /path/to/ssh-key
```

When `controlPlane` is omitted, `apiEndpoint.host` is **required** (validated
by KubeOne).

## Credentials

KubeOne reads the infra cluster kubeconfig from the credentials file using the
`KUBEVIRT_KUBECONFIG` key. The value can be a plain or base64-encoded
kubeconfig YAML.

The `KUBEVIRT_KUBECONFIG` environment variable is also supported for
convenience during development.
