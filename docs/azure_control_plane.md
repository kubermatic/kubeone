# Azure Managed Control Plane

KubeOne can provision control-plane nodes directly on Azure using machine-controller, eliminating the need for
pre-provisioned servers or external tooling (e.g. Terraform) to manage control-plane VMs.

When `cloudProvider.azure.controlPlane` is configured in your `kubeone.yaml`, KubeOne will:

1. Ensure an Azure Public IP and Load Balancer exist (creates them if missing) and set the `apiEndpoint` from the Public
   IP address.
2. Provision control-plane VMs via machine-controller's Azure driver, driven by the `controlPlane.nodeSets` spec.
3. Attach each provisioned control-plane VM's primary NIC to the load balancer backend pool so traffic is routed to
   kube-apiserver.

If `cloudProvider.azure.controlPlane` is omitted, existing behaviour is preserved - you must supply `apiEndpoint` and
control-plane host IPs yourself (e.g. via Terraform's `kubeone_hosts` and `kubeone_api` outputs).

## Prerequisites

- Azure credentials with permissions to manage virtual machines, network interfaces, public IP addresses, and load
  balancers (see [Credentials](#credentials) below).
- Existing Azure networking for control-plane nodes (resource group, VNet, subnet, network security group, route table,
  and availability set).

The Terraform example in `examples/terraform/azure-managed-control-plane` provisions this base infrastructure.

## Configuration

```yaml
apiVersion: kubeone.k8c.io/v1beta2
kind: KubeOneCluster
name: my-cluster

versions:
  kubernetes: 1.36.2

cloudProvider:
  azure:
    controlPlane:
      loadBalancer:
        # Name of the load balancer to create. Default: "<CLUSTER_NAME>-kubeapi"
        name: my-cluster-kubeapi

        # Resource group where the load balancer is created.
        # Default: resourceGroup of the first control plane NodeSet.
        resourceGroup: my-cluster-rg

        # Azure region where the load balancer is created.
        # Default: location of the first control plane NodeSet.
        location: westeurope

        # SKU of the load balancer. Possible values: "Standard" (default), "Basic".
        sku: Standard

        # Name of the public IP attached to the load balancer.
        # Default: "<LOAD_BALANCER_NAME>-pubip".
        publicIPName: my-cluster-kubeapi-pubip

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
        # Values from the `networking` Terraform output.
        location: westeurope
        resourceGroup: my-cluster-rg
        vnetName: my-cluster-vpc
        subnetName: my-cluster-subnet
        securityGroupName: my-cluster-sg
        routeTableName: my-cluster-rt
        availabilitySet: my-cluster-avset
        loadBalancerSku: Standard

        # Values from the `vm` Terraform output.
        imageReference:
          publisher: Canonical
          offer: ubuntu-24_04-lts
          sku: server-gen1
          version: latest

        vmSize: Standard_F2s_v2
        assignPublicIP: true
        publicIPSKU: Standard
```

## Load Balancer Configuration

The `controlPlane.loadBalancer` section supports:

| Field | Default | Description |
|-------|---------|-------------|
| `name` | `<CLUSTER_NAME>-kubeapi` | Name of the Azure load balancer |
| `resourceGroup` | first control-plane NodeSet's `resourceGroup` | Resource group where the load balancer and public IP are created |
| `location` | first control-plane NodeSet's `location` | Azure region where the load balancer and public IP are created |
| `sku` | `Standard` | Load balancer SKU (`Standard` or `Basic`) |
| `publicIPName` | `<LOAD_BALANCER_NAME>-pubip` | Name of the public IP resource attached to the load balancer |

### How the load balancer is provisioned

During `kubeone apply`, KubeOne uses the Azure Network API to:

1. Resolve load balancer settings from `controlPlane.loadBalancer`. If `resourceGroup` and/or `location` are not set
   there, KubeOne falls back to the first control-plane NodeSet's `cloudProviderSpec.resourceGroup` and
   `cloudProviderSpec.location`.

2. Ensure a static IPv4 Public IP exists for the kube-apiserver endpoint.

3. Ensure the load balancer exists with:
   - One frontend IP configuration bound to that Public IP
   - One backend pool for control-plane nodes
   - A TCP health probe on port 6443
   - A TCP load-balancing rule for port 6443

4. As each control-plane VM is provisioned, attach its primary NIC to the backend pool.

When the API endpoint is already known (for example from a previous `kubeone apply`), the load balancer creation step is
skipped.

## Without Managed Control Plane

If you prefer to manage control-plane servers with Terraform (or another tool), omit the `controlPlane` section on
`azure` and use static hosts:

```yaml
cloudProvider:
  azure: {}

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

In this mode, `apiEndpoint.host` is required.

## Credentials

KubeOne reads Azure credentials from the credentials file (or environment) using the following keys:

```ini
ARM_CLIENT_ID=<your-client-id>
ARM_CLIENT_SECRET=<your-client-secret>
ARM_TENANT_ID=<your-tenant-id>
ARM_SUBSCRIPTION_ID=<your-subscription-id>
```

For machine-controller integration, KubeOne maps these credentials to the corresponding `AZURE_*` environment variables
automatically.

Pass the credentials file to KubeOne via the `--credentials` flag:

```bash
kubeone apply --manifest kubeone.yaml --credentials credentials.yaml
```
