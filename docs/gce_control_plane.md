# GCP Managed Control Plane

KubeOne can provision control-plane nodes directly on Google Cloud Platform (GCP) as Compute Engine VMs using
machine-controller, eliminating the need for pre-provisioned VMs or external tooling (e.g. Terraform) to manage
control-plane VMs.

When `controlPlane.nodeSets` is configured together with `cloudProvider.gce` in your `kubeone.yaml`, KubeOne will:

1. Ensure the firewall rules and a backend service-based regional external passthrough Network Load Balancer exist for
   the kube-apiserver endpoint (creates them if missing) and set the `apiEndpoint` from its static IP address.
2. Provision control-plane VMs via machine-controller's GCE driver, driven by the `controlPlane.nodeSets` spec.
3. Add each control-plane VM to the load balancer as an endpoint of the network endpoint group of its zone, once it
   serves kube-apiserver, see [Load balancer membership](#load-balancer-membership).

If `controlPlane.nodeSets` is omitted, existing behaviour is preserved — you must supply `apiEndpoint` and
control-plane host IPs yourself (e.g. via Terraform's `kubeone_hosts` and `kubeone_api` outputs).

## Prerequisites

- A GCP project with the Compute Engine API enabled.
- A service account JSON key with the roles listed under [Credentials](#credentials) below. A gcloud or OAuth access
  token is not enough: the same key is used by machine-controller, the GCP cloud controller manager and the CSI driver
  in the cluster.
- A VPC network that allows SSH (22) from where KubeOne runs and all traffic between the nodes. The `default` network
  does this with its `default-allow-ssh` and `default-allow-internal` rules; for other networks create equivalent
  firewall rules. KubeOne only creates the firewall rules of the kube-apiserver load balancer.
- Control-plane VMs need a public IP address (`assignPublicIPAddress: true`, the default), or a bastion
  (`ssh.bastion`), so that KubeOne can reach them over SSH.
- The cluster name, the nodeSet names and the load balancer name are used in GCP resource names and labels, which must
  be 1-63 lowercase letters, digits or `-` and start with a letter. KubeOne validates all derived names, e.g.
  `<CLUSTER_NAME>-<NODESET>-<INDEX>`, `<CLUSTER_NAME>-control-plane` and `<LOAD_BALANCER_NAME>-<ZONE>`.

## Configuration

```yaml
apiVersion: kubeone.k8c.io/v1beta2
kind: KubeOneCluster
name: my-cluster

versions:
  kubernetes: 1.36.2

cloudProvider:
  external: true
  gce:
    controlPlane:
      loadBalancer:
        # Project of the load balancer resources, the control-plane VMs must be in the same project.
        # Default: project_id of the service account
        projectID: my-project

        # Region of the load balancer, must contain the zones of all control-plane nodeSets.
        # Default: region of the first nodeSet zone
        region: europe-west3

        # Network of the control-plane VMs, where the firewall rules are created.
        # Default: the network of the nodeSets, resolved when the load balancer is ensured
        # network: global/networks/default

        # Name of the load balancer resources. Default: "<CLUSTER_NAME>-kubeapi"
        name: my-cluster-kubeapi

        # IPv4 CIDRs allowed to reach kube-apiserver through the load balancer. Default: ["0.0.0.0/0"]
        allowedSourceRanges:
          - 0.0.0.0/0
  cloudConfig: |
    [global]
    regional = true
    multizone = true

controlPlane:
  # One nodeSet per zone spreads the control plane over the zones of the region
  nodeSets:
    - name: cp-a
      replicas: 1
      operatingSystem: ubuntu
      ssh:
        publicKeys:
          - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA...
        username: ubuntu
      cloudProviderSpec:
        zone: europe-west3-a
        machineType: e2-standard-2
        diskSize: 50
        diskType: pd-ssd
        assignPublicIPAddress: true
        multizone: true
        # network: global/networks/my-vpc
        # subnetwork: regions/europe-west3/subnetworks/my-subnet
    - name: cp-b
      replicas: 1
      operatingSystem: ubuntu
      ssh:
        publicKeys:
          - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA...
        username: ubuntu
      cloudProviderSpec:
        zone: europe-west3-b
        machineType: e2-standard-2
        diskSize: 50
        diskType: pd-ssd
        assignPublicIPAddress: true
        multizone: true
    # cp-c in europe-west3-c likewise
```

`cloudProviderSpec` is the machine-controller GCE provider spec and is decoded strictly: unknown fields are rejected.
The `zone` field is required and must be in the load balancer region. The `serviceAccount` field must not be set,
KubeOne passes the service account from the credentials. The VMs run as that service account unless
`disableMachineServiceAccount: true` is set. All nodeSets must be in the same network.

KubeOne adds the network tag `<CLUSTER_NAME>-control-plane` and the labels `kubeone_cluster_name` and `kubeone_role` to
the control-plane VMs. The control-plane nodes are named after their instances, e.g. `my-cluster-cp-a-0`.

The GCP cloud controller manager targets the firewall rules of `LoadBalancer` Services at a network tag of each node
that is a prefix of its instance name. Give the workers such a tag, e.g. the MachineDeployment name, worker instances
are named `<MACHINEDEPLOYMENT>-<HASH>-<SUFFIX>`.

A complete example is in `examples/terraform/gce/kubeone-managed-cp.yaml`.

## Load Balancer Configuration

The `controlPlane.loadBalancer` section supports:

| Field | Default | Description |
|-------|---------|-------------|
| `projectID` | `project_id` of the service account | Project of the load balancer resources. The control-plane VMs must be in the same project, set `cloudProviderSpec.projectID` of the nodeSets if it's not the project of the service account |
| `region` | region of the first nodeSet zone | Region of the load balancer, must contain the zones of all control-plane nodeSets |
| `network` | network of the nodeSets | Network of the control-plane VMs, where the firewall rules are created. Resolved from `cloudProviderSpec.network`, or the network of `cloudProviderSpec.subnetwork`, else `default`; if set, it must be that network |
| `name` | `<CLUSTER_NAME>-kubeapi` | Name of the load balancer resources, and prefix of the network endpoint groups and firewall rules |
| `allowedSourceRanges` | `["0.0.0.0/0"]` | IPv4 CIDRs allowed to reach kube-apiserver through the load balancer |

The load balancer preserves the client source IP, so `allowedSourceRanges` is the only access control. The nodes
reach kube-apiserver through the load balancer from their public (or Cloud NAT) IPs: if you restrict
`allowedSourceRanges`, include them, otherwise control-plane nodes can't join and workers can't register. KubeOne logs
a warning if `0.0.0.0/0` is not included. The load balancer is IPv4 only and external only.

If `apiEndpoint.host` is set, KubeOne doesn't create or look up the load balancer and uses the given endpoint as-is,
but still creates the firewall rules. On reset, a load balancer created by KubeOne is only deleted if
`apiEndpoint.host` is its IP address.

### How the load balancer is provisioned

During `kubeone apply`, KubeOne uses the Compute Engine API to:

1. Create the firewall rules, in the network of the control-plane VMs, targeting the `<CLUSTER_NAME>-control-plane`
   network tag:
   - `<name>-6443`: TCP 6443 from `allowedSourceRanges`
   - `<name>-6443-hc`: TCP 6443 from the health check probe ranges `35.191.0.0/16`, `209.85.152.0/22` and
     `209.85.204.0/22`

   The firewall rules of a [Shared VPC](https://docs.cloud.google.com/vpc/docs/shared-vpc) network are created in its
   host project.

2. Create the [backend service-based regional external passthrough Network Load
   Balancer](https://docs.cloud.google.com/load-balancing/docs/network/networklb-backend-service), the type Google
   recommends over target pools, all named `<name>`:
   - A regional static external IP address, which becomes `apiEndpoint.host`
   - A regional TCP health check on port 6443
   - A regional backend service (`EXTERNAL`, TCP) using the health check
   - A regional forwarding rule (`EXTERNAL`, TCP 6443) using the address and the backend service

3. Provision the control-plane VMs, and add each VM as an endpoint to the zonal network endpoint group
   (`GCE_VM_IP`) `<name>-<zone>`, which is created and added to the backend service with the first VM of its zone.
   Network endpoint groups are used instead of instance groups because a VM can only be in one load-balanced instance
   group, and the GCP cloud controller manager uses instance groups for `LoadBalancer` Services.

Every resource gets the description `kubeone:<CLUSTER_NAME>`. KubeOne never adopts, changes or deletes a resource
without it: if a resource with the same name but another description exists, `kubeone apply` fails. Existing firewall
rules are updated if `allowedSourceRanges` changed. Every operation is waited for until it's done.

### Load balancer membership

The GCE guest environment adds a local route for the load balancer address on every VM behind a passthrough load
balancer, so a VM can accept the load balanced traffic. Traffic a VM sends to the load balancer address is therefore
[always answered by the VM itself](https://docs.cloud.google.com/load-balancing/docs/internal/setting-up-internal),
regardless of its health. A control-plane VM that joins the cluster through the load balancer must not be behind it
yet, otherwise it reaches itself, finds no kube-apiserver and `kubeadm join` hangs. KubeOne therefore:

1. Adds only the first VM of the first nodeSet, which initializes a new cluster, if no other member of the load
   balancer is healthy.
2. Before the followers join, adds the elected leader and removes the control-plane VMs that are about to join, e.g.
   VMs of a cluster that was reset without destroying them, or a recreated VM.
3. Adds all control-plane VMs after they joined, on every `kubeone apply`.

Each control-plane node's kubelet and other components reach kube-apiserver through the load balancer address, which
is always their own kube-apiserver. If a node's kube-apiserver stops, that node turns `NotReady` instead of failing
over to another control-plane node. Workers are not affected.

Membership is not cleaned up in every case:

- Scaling down the control plane or renaming a nodeSet doesn't remove the former VMs from the load balancer; remove
  their endpoints with `gcloud compute network-endpoint-groups update <name>-<zone> --remove-endpoint=instance=<vm>`.
- If a run fails after the followers joined but before they were added back to the load balancer, and the cluster is
  not healthy, the next `kubeone apply` doesn't add them; they are added by the first `kubeone apply` on a healthy
  cluster.
- The health check is TCP: a stale kube-apiserver of a VM whose `kubeadm reset` failed still counts as healthy, and
  the leader of a new cluster isn't added to the load balancer then. Reset or delete that VM first.

### Reset

`kubeone reset --destroy-control-plane --destroy-control-plane-load-balancer` deletes the control-plane VMs and then
the load balancer resources and firewall rules, in reverse order of their dependencies. Add `--force` if
kube-apiserver is not reachable, e.g. after a failed `kubeone apply`; missing VMs and load balancer resources are
skipped.

- The machine-controller GCE driver deletes VMs by name only, so before deleting any VM KubeOne checks that every
  existing VM with a control-plane name has the labels set by KubeOne and machine-controller.
- Only resources with the description `kubeone:<CLUSTER_NAME>` are deleted, others are logged and kept.
- If `apiEndpoint.host` is set to something else than the load balancer address, including a DNS name pointing at it,
  the load balancer is kept and only its firewall rules are deleted.
- `kubeone reset` without `--destroy-control-plane` keeps the VMs and their load balancer membership. The next
  `kubeone apply` removes the VMs that join again from the load balancer until they joined.
- If the network of the control plane can't be resolved during reset, e.g. its subnetwork was deleted already, the
  firewall rules are looked for in the load balancer project; delete them manually if they are in a Shared VPC host
  project.

The cluster name, the nodeSet names and the project must not change between `kubeone apply` and `kubeone reset`, see
[Credentials](#credentials). gcloud commands to inspect and delete the resources are in
`examples/terraform/gce/managed-cp-gcloud.md`.

## Without Managed Control Plane

If you prefer to manage control-plane VMs with Terraform (or another tool), e.g. with `examples/terraform/gce`, omit
`controlPlane.nodeSets` and use static hosts:

```yaml
cloudProvider:
  gce: {}

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

In this mode, `cloudProvider.gce.controlPlane` is not used and `apiEndpoint.host` is required.

## Credentials

KubeOne reads the service account JSON key from the credentials file using the `GOOGLE_CREDENTIALS` key. The service
account needs these roles in the project (and `roles/compute.securityAdmin` in the host project of a Shared VPC
network):

| Role | Used for |
|------|----------|
| `roles/compute.instanceAdmin.v1` | Control-plane VMs |
| `roles/compute.loadBalancerAdmin` | Health check, backend service, forwarding rule, network endpoint groups |
| `roles/compute.securityAdmin` | Firewall rules |
| `roles/compute.publicIpAdmin` | Static IP address |
| `roles/iam.serviceAccountUser` | Running the VMs as the service account |

```yaml
GOOGLE_CREDENTIALS: |
  { "type": "service_account", "project_id": "my-project", ... }
```

The `GOOGLE_CREDENTIALS` environment variable is also supported for convenience during development. For
machine-controller integration, KubeOne maps the key to the base64 encoded `GOOGLE_SERVICE_ACCOUNT` environment
variable automatically.

Unless `loadBalancer.projectID` or `cloudProviderSpec.projectID` is set, the VMs and the load balancer are created in
the key's `project_id`. KubeOne doesn't record the project of a cluster, so use a key of the same project for every
`kubeone apply` and `kubeone reset`, or set `loadBalancer.projectID` and `cloudProviderSpec.projectID` of every
nodeSet. With a key of another project, `kubeone apply` creates a second cluster in that project, and `kubeone reset`
finds nothing to delete and succeeds, leaving the VMs and the load balancer in the original project. Another service
account of the same project works if it has the roles above.

Pass the credentials file to KubeOne via the `--credentials` flag:

```bash
kubeone apply --manifest kubeone.yaml --credentials credentials.yaml
```
