# Azure Managed Control Plane Terraform configs

The Azure Managed Control Plane Terraform configs provision the supporting
infrastructure for a Kubernetes HA cluster whose control plane is created and
managed directly by KubeOne (rather than by Terraform). They create the resource
group, virtual network, subnet, route table, network security group, and
availability set needed by the control-plane and worker nodes, but intentionally
do **not** create any control-plane instances or a load balancer — KubeOne
provisions those itself.

Check out the [Creating Infrastructure guide][docs-infrastructure] to learn more
about how to use the configs and how to provision a Kubernetes cluster using
KubeOne.

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/main/guides/using-terraform-configs/

## Managed Control Plane

Instead of provisioning control-plane instances and an API load balancer with
Terraform (as the regular `azure` example does), this config only lays the
groundwork and lets KubeOne provision and manage the control plane directly via
`cloudProvider.azure.controlPlane` and `controlPlane.nodeSets` in `kubeone.yaml`.

When using the managed control plane:

- Set `cloudProvider.azure.controlPlane.loadBalancer` in `kubeone.yaml`. KubeOne
  creates its own load balancer and public IP, so there is no
  `kubeone_api`/`kubeone_hosts` Terraform output to consume.
- The `networking` output exposes the resource group, location, VNet, subnet,
  network security group, route table, availability set, and load balancer SKU
  that should be referenced in each `nodeSets[].cloudProviderSpec`.
- The `vm` output exposes the `imageReference` and `imagePlan` to use for the
  control-plane machines.

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

controlPlane:
  nodeSets:
    - name: cp
      replicas: 3
      operatingSystem: ubuntu
      ssh:
        publicKeys:
          - ssh-ed25519 AAAA...
        username: ubuntu
      cloudProviderSpec:
        # Values from the `networking` output.
        location: westeurope
        resourceGroup: my-cluster-rg
        vnetName: my-cluster-vpc
        subnetName: my-cluster-subnet
        securityGroupName: my-cluster-sg
        routeTableName: my-cluster-rt
        availabilitySet: my-cluster-avset
        loadBalancerSku: Standard

        # Values from the `vm` output.
        imageReference:
          publisher: Canonical
          offer: ubuntu-24_04-lts
          sku: server-gen1
          version: latest

        vmSize: Standard_F2s_v2
        assignPublicIP: true
        publicIPSKU: Standard
```
