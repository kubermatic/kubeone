# Managed Control Plane Implementation Notes

**Status:** **Draft**
**Created:** 2026-04-06
**Last updated:** 2026-04-06
**Author:** Artiom Diomin ([@kron4eg](https://github.com/kron4eg))

---

## Overview

This branch implements the foundation for **full node lifecycle management** of Kubernetes control-plane nodes on
Hetzner Cloud, removing the requirement for pre-provisioned hosts. Instead of supplying a static list of SSH addresses,
users declare a `controlPlane.nodeSets` spec and KubeOne provisions (or discovers) the underlying VMs and load balancer
automatically using the machine-controller cloud-provider drivers.

The work is tracked by the design proposal at
[docs/proposals/20250513-mc-control-plane.md](20250513-mc-control-plane.md).

---

## Changes by Area

### 1. New API types

#### `ControlPlaneConfig` gains `NodeSets`

```go
type ControlPlaneConfig struct {
    Hosts    []HostConfig `json:"hosts,omitempty"`
    NodeSets []NodeSet    `json:"nodeSets,omitempty"`
}
```

`NodeSets` is the new declarative way to describe managed control-plane nodes. It is an alternative to the static
`Hosts` list; both can coexist during migration.

#### `NodeSet`

Describes a group of identically configured control-plane nodes:

```go
type NodeSet struct {
    Name                string              `json:"name"`
    Replicas            int                 `json:"replicas"`
    Generation          int                 `json:"generation,omitempty"`
    NodeSettings        NodeSettingsSpec    `json:"nodeSettings,omitempty"`
    OperatingSystem     OperatingSystemName `json:"operatingSystem"`
    OperatingSystemSpec OperatingSystemSpec `json:"operatingSystemSpec,omitempty"`
    SSH                 SSHSpec             `json:"ssh"`
    CloudProviderSpec   json.RawMessage     `json:"cloudProviderSpec"`
}
```

`CloudProviderSpec` is a raw JSON blob passed verbatim to the machine-controller cloud-provider (Hetzner `RawConfig` in
this iteration).

#### `NodeSettingsSpec`

```go
type NodeSettingsSpec struct {
    Labels      map[string]string `json:"labels,omitempty"`
    Annotations map[string]string `json:"annotations,omitempty"`
    Taints      []corev1.Taint    `json:"taints,omitempty"`
}
```

`Annotations` was changed from `[]string` to `map[string]string` to match Kubernetes conventions.  `KubeletConfig` was
removed from this struct (it belongs on `HostConfig`).

#### `SSHSpec`

A new dedicated type replacing the repeated SSH fields found on `HostConfig`,
covering: `PublicKeys`, `Port`, `Username`, `PrivateKeyFile`, `CertFile`,
`HostPublicKey`, `AgentSocket`, `Bastion*` fields.

#### `OperatingSystemSpec`

```go
type OperatingSystemSpec struct {
    DistUpgradeOnBoot bool `json:"distUpgradeOnBoot,omitempty"`
}
```

#### `HetznerSpec` extended

```go
type HetznerSpec struct {
    NetworkID    string                `json:"networkID,omitempty"`
    ControlPlane *HetznerControlPlane  `json:"controlPlane,omitempty"`
}
```

#### `HetznerControlPlane` / `HetznerLoadBalancer` (new)

```go
type HetznerControlPlane struct {
    LoadBalancer HetznerLoadBalancer `json:"loadBalancer"`
}

type HetznerLoadBalancer struct {
    Name     string            `json:"name,omitempty"`     // default: "<cluster>-kubeapi"
    Type     string            `json:"type,omitempty"`     // default: "lb11"
    Location string            `json:"location,omitempty"` // default: "nbg1"
    PublicIP *bool             `json:"publicIP,omitempty"` // default: true
    Labels   map[string]string `json:"labels,omitempty"`
}
```

All fields are optional; sensible defaults are applied during defaulting.

---

### 2. Defaulting

* `SetDefaults_NodeSet` / `setDefaultsNodeSets` – populates SSH defaults (`Port: 22`, `Username: "root"`, `AgentSocket:
  "env:SSH_AUTH_SOCK"`) and adds the standard `node-role.kubernetes.io/control-plane:NoSchedule` taint when none is
  specified.
* For Hetzner clusters that declare `NodeSets`, a `HetznerControlPlane` struct is auto-created on `HetznerSpec` if
  absent (so the load-balancer name and type have valid defaults even without an explicit `controlPlane:` key).

---

### 3. Validation

`validateHetznerSpec` now enforces that `networkID` is provided whenever `hetzner.controlPlane` is set:

```
networkID is required when controlPlane is specified
```

A missing `networkID` would prevent the load-balancer from being attached to the private network, which is required for
the LB → node routing.

---

### 4. New `pkg/provisioner` package

A brand-new package that wraps machine-controller cloud-provider calls to provide VM lifecycle operations without a
running machine-controller.

#### `provisioner.go`

| Function | Purpose |
|----------|---------|
| `FindMachines` | Looks up existing cloud instances for a list of `Machine` objects. Returns error if any instance is missing. |
| `FindOrCreateMachines` | Idempotent: creates an instance if it does not exist, then polls until both a public and private IP are available (up to 5 retries × 5 s). |
| `getProvider` | Resolves the correct machine-controller `Provider` from a `Machine`'s `ProviderSpec`. |
| `getUserData` | Renders a minimal `#cloud-config` that injects SSH public keys into new instances. |

#### `output.go`

`GetMachineInfo` inspects `cloud.Instance.Addresses()` and extracts
`PublicAddress`, `PrivateAddress`, and `Hostname`. **IPv4 is preferred; IPv6 is
used as a fallback** (commit `3103f572`).

`publicAndPrivateIPExist` is a helper to detect when a freshly-created instance
has finished network assignment.

The `Machine` struct is the bridge between the provisioner and the rest of
KubeOne (`HostConfig`):

```go
type Machine struct {
    PublicAddress  string
    PrivateAddress string
    Hostname       string
}
```

---

### 5. New cloud resources provision tasks

This wires the provisioner into the KubeOne task pipeline:

#### `WithFindControlPlane()`

Used by all commands that need to discover existing infrastructure (e.g.
`status`, `kubeconfig`, `reset`, `etcd`, `proxy`, `migrate`, `certificates`).

Tasks added:

1. **Find Hetzner load balancer** – calls `lookupHetznerLoadBalancer` to resolve the LB by name and populate
   `s.Cluster.APIEndpoint.Host`.  Skipped if an endpoint is already known.
2. **Find Hetzner VMs** – calls `lookupHetznerVMs` to find existing instances matching the generated `Machine` manifests
   and append them to `s.Cluster.ControlPlane.Hosts`.
3. **Defaulting cluster hosts** – re-runs defaulting (via a round-trip through the v1beta3 scheme) so that all fields
   (e.g. `IsLeader`, host IDs) are populated after the dynamic hosts are injected.

Both the LB and VM tasks are guarded by the `isHetznerControlPlaneEnabled` predicate (`cloudProvider.hetzner != nil &&
len(controlPlane.nodeSets) > 0`).

#### `WithEnsureControlPlane()`

Used only during `kubeone apply` to create-or-reconcile infrastructure before cluster bootstrap:

1. **Ensure Hetzner load balancer** – calls `ensureHetznerLoadBalancer`. Creates a new LB (TCP/6443, label-selector
   target, private-network attached) if one doesn't exist; otherwise reuses the existing one.
2. **Ensure Hetzner control-plane VM** (one task *per* machine) – calls `ensureHetznerControlPlaneVM` →
   `provisioner.FindOrCreateMachines`.
3. **Defaulting cluster hosts** – same as above.

The split between `WithEnsureControlPlane` (creates) and `WithFindControlPlane` (looks up) was introduced so that
read-only commands do not accidentally create cloud resources.

#### `generateHetznerControlPlaneMachines`

Converts `[]NodeSet` into `[]clusterv1alpha1.Machine` objects (the CAPI Machine CRD format understood by
machine-controller):

* Labels `kubeone_cluster_name`, `kubeone_role=api/control-plane`, and a `kubeone_own_since_timestamp` (Unix epoch) are
  automatically merged into both the Machine metadata labels and the cloud-provider `hetznerConfig.Labels`.
* Machines are named `<cluster>-<nodeSet.Name>-<index>`.
* SSH public keys are injected via `cloud-config` user-data.

#### `hostConfigsFromMachines`

Converts `[]provisioner.Machine` back into `[]kubeoneapi.HostConfig`, mapping SSH settings, labels, annotations, taints,
and OS from the originating `NodeSet`.

---

### 6. Terraform state-file shortcut

`TFOutput` now has a fast path for the common case where a Terraform directory is given and a local `terraform.tfstate`
file is present:

```go
func readStateFileOutputs(dir string) ([]byte, error) {
    stateBytes, _ := os.ReadFile(filepath.Join(dir, "terraform.tfstate"))
    // parse state JSON, extract .outputs
    return state.Outputs, nil
}
```

The `outputs` section of a `tfstate` file has the same structure as `terraform output -json`, so this avoids spawning a
`terraform` subprocess when the state is local.  If the file is absent (remote backend), the code falls back to running
the Terraform CLI as before.

---

### 7. apply integration

`runApply` now calls `WithEnsureControlPlane` before probing the cluster:

```go
managedCP, err := tasks.WithEnsureControlPlane(nil,
    st.Cluster.Name,
    st.Cluster.ControlPlane.NodeSets,
    st.Cluster.Versions.Kubernetes,
)
...
if ops := managedCP.Descriptions(st); len(ops) > 0 {
    // print planned operations and ask for confirmation
    managedCP.Run(st)
}
```

This keeps the managed-CP provisioning step distinct from the main cluster
reconciliation, and it respects `--auto-approve`.

---

## Example `KubeOneCluster` manifest (Hetzner, managed CP)

```yaml
apiVersion: kubeone.k8c.io/v1beta2
kind: KubeOneCluster
name: my-cluster

versions:
  kubernetes: 1.32.2

cloudProvider:
  hetzner:
    networkID: my-private-network
    controlPlane:
      loadBalancer:
        name: my-cluster-kubeapi    # optional, this is the default
        type: lb11                  # optional, default
        location: nbg1              # optional, default
        publicIP: true              # optional, default
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
          - "ssh-ed25519 AAAA..."
        username: ubuntu
      cloudProviderSpec:
        serverType: cx22
        location: nbg1
        image: ubuntu-24.04
        networks:
          - my-private-network
```

---

## What is NOT yet implemented

* Support for cloud providers other than Hetzner.
* Cluster upgrade path for `NodeSets`-based control planes.
* Deletion / scale-down of control-plane nodes.
* End-to-end test coverage.
