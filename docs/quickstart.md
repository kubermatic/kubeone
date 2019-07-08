# How To Install a Kubernetes Cluster Using KubeOne

In this quick start we're going to show how to get started with KubeOne. We'll cover how to create the needed infrastructure using our example Terraform scripts and then install Kubernetes. Finally, we're going to show how to destroy the cluster along with the infrastructure.

As a result, you'll get a High-Available (HA) Kubernetes cluster with three control and worker nodes.

## Preparation

#### Install Terraform
`terraform` v0.12.0 or later need to be installed. Older releases are not compatible. The binaries for `terraform` can be found on the [Terraform website](https://www.terraform.io/downloads.html)

#### Install KubeOne
`kubeone` v0.9.0 or later need to be installed, which can be done by following the `Installing KubeOne` section of [the README](https://github.com/kubermatic/kubeone/blob/master/README.md),

#### Set variables
To use KubeOne you first need to set variables appropriate for your provider. See [here](./environment_variables.md) how to do that.

#### Provide infrastructure
To provision your infrastructure you can use our Terraform plans. See [here](./quickstart_terraform.md) on how to manage that.

#### Create configuration
<details><summary>AWS</summary><p>

Before you start you'll need a configuration file that defines how Kubernetes will be installed, e.g. what version will be used and what features will be enabled. For the configuration file reference run `kubeone config print --full`.

To get started you can use the following configuration. It'll install Kubernetes 1.14.2 and create 3 worker nodes. KubeOne automatically populates all needed information about worker nodes from the [Terraform output](https://github.com/kubermatic/kubeone/blob/ec8bf305446ac22529e9683fd4ce3c9abf753d1e/examples/terraform/aws/output.tf#L38-L87). Alternatively, you can set those information manually. As KubeOne is using [Kubermatic `machine-controller`](https://github.com/kubermatic/machine-controller) for creating worker nodes, see [AWS example manifest](https://github.com/kubermatic/machine-controller/blob/master/examples/aws-machinedeployment.yaml) for available options.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo
versions:
  kubernetes: '1.14.2'
cloudProvider:
  name: 'aws'
```

</p></details><details><summary>Azure</summary><p>

Before you start you'll need a configuration file that defines how Kubernetes will be installed, e.g. what version will be used and what features will be enabled. For the configuration file reference run `kubeone config print --full`.

To get started you can use the following configuration. It'll install Kubernetes 1.14.2 and create 1 worker nodes. KubeOne automatically populates information about template, VM size and networking settings for worker nodes from the Terraform output. Alternatively, you can set those information manually. As KubeOne is using [Kubermatic `machine-controller`][7] for creating worker nodes, see [Azure example manifest][8] for available options.

For Azure you also need to provide a `cloud-config` file containing credentials, so Azure Cloud Controller Manager works as expected. Make sure to replace sample values with real values.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster

versions:
  kubernetes: '1.14.2'
cloudProvider:
  name: 'azure'
  cloudConfig: |
    {
      "tenantId": "<AZURE TENANT ID>",
      "subscriptionId": "<AZURE SUBSCIBTION ID>",
      "aadClientId": "<AZURE CLIENT ID>",
      "aadClientSecret": "<AZURE CLIENT SECRET>",
      "resourceGroup": "<SOME RESOURCE GROUP>",
      "location": "westeurope",
      "subnetName": "<SOME SUBNET NAME>",
      "routeTableName": "",
      "securityGroupName": "<SOME SECURITY GROUP>",
      "vnetName": "<SOME VIRTUAL NETWORK>",
      "primaryAvailabilitySetName": "<SOME AVAILABILITY SET NAME>",
      "useInstanceMetadata": true,
      "useManagedIdentityExtension": false,
      "userAssignedIdentityID": ""
    }
```

</p></details><details><summary>Digital Ocean</summary><p>

Before you start you'll need a configuration file that defines how Kubernetes will be installed, e.g. what version will be used and what features will be enabled. For the configuration file reference run `kubeone config print --full`.

To get started you can use the following configuration. It'll install Kubernetes 1.14.2, create 3 worker nodes and deploy the [external cloud controller manager](https://github.com/digitalocean/digitalocean-cloud-controller-manager). The external cloud controller manager takes care of providing correct information about nodes from the DigitalOcean API and allows you to use the `type: LoadBalancer` services. KubeOne automatically populates information about the worker nodes from the [Terraform output](https://github.com/kubermatic/kubeone/blob/ec8bf305446ac22529e9683fd4ce3c9abf753d1e/examples/terraform/digitalocean/output.tf#L38-L59). Alternatively, you can set those information manually. As KubeOne is using [Kubermatic `machine-controller`](https://github.com/kubermatic/machine-controller) for creating worker nodes, see [DigitalOcean example manifest](https://github.com/kubermatic/machine-controller/blob/master/examples/digitalocean-machinedeployment.yaml) for available options.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo
versions:
  kubernetes: '1.14.2'
cloudProvider:
  name: 'digitalocean'
  external: true
```

**Note:** It is recommended to enable private networking so cluster can function properly.

</p></details><details><summary>GKE</summary><p>

Before you start you'll need a configuration file that defines how Kubernetes will be installed, e.g. what version will be used and what features will be enabled. For the configuration file reference run `kubeone config print --full`.

To get started you can use the following configuration. It'll install Kubernetes 1.14.2 and create 3 worker nodes. KubeOne automatically populates information about worker nodes from the [Terraform output](https://github.com/kubermatic/kubeone/blob/ec8bf305446ac22529e9683fd4ce3c9abf753d1e/examples/terraform/gce/output.tf#L41-L81). Alternatively, you can set those information manually. As KubeOne is using [Kubermatic `machine-controller`](https://github.com/kubermatic/machine-controller) for creating worker nodes, see [GCE example manifest](https://github.com/kubermatic/machine-controller/blob/master/examples/gce-machinedeployment.yaml) for available options.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo
versions:
  kubernetes: '1.14.2'
cloudProvider:
  name: 'gce'
```

</p></details><details><summary>Hetzner</summary><p>

Before you start you'll need a configuration file that defines how Kubernetes will be installed, e.g. what version will be used and what features will be enabled. For the configuration file reference run `kubeone config print --full`.

To get started you can use the following configuration. It'll install Kubernetes 1.14.2, create 3 worker nodes and deploy the [external cloud controller manager](https://github.com/hetznercloud/hcloud-cloud-controller-manager). The external cloud controller manager takes care of providing correct information about the nodes. KubeOne automatically populates all needed information about worker nodes from the [Terraform output](https://github.com/kubermatic/kubeone/blob/a874fd5913ca2a86c3b8136982c2a00e835c2f62/examples/terraform/hetzner/output.tf#L26-L36). Alternatively, you can set those information manually. As KubeOne is using [Kubermatic `machine-controller`](https://github.com/kubermatic/machine-controller) for creating worker nodes see [Hetzner example manifest](https://github.com/kubermatic/machine-controller/blob/master/examples/hetzner-machinedeployment.yaml) for available options.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo
versions:
  kubernetes: '1.14.2'
cloudProvider:
  name: 'hetzner'
  external: true
```

</p></details><details><summary>OpenStack</summary><p>

Before you start you'll need a configuration file that defines how Kubernetes will be installed, e.g. what version will be used and what features will be enabled. For the configuration file reference run `kubeone config print --full`.

To get started you can use the following configuration. It'll install Kubernetes 1.14.2 and create 2 worker nodes. KubeOne automatically populates information about image, instance size and networking settings for worker nodes from the Terraform output. Alternatively, you can set those information manually. As KubeOne is using [Kubermatic `machine-controller`](https://github.com/kubermatic/machine-controller) for creating worker nodes, see [OpenStack example manifest](https://github.com/kubermatic/machine-controller/blob/master/examples/openstack-machinedeployment.yaml) for available options.

For OpenStack you also need to provide a `cloud-config` file containing credentials, so OpenStack Cloud Controller Manager works as expected. Make sure to replace sample values with real values.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo
versions:
  kubernetes: '1.14.2'
cloudProvider:
  name: 'openstack'
  cloudConfig: |
    [Global]
    username=OS_USERNAME
    password=OS_PASSWORD
    auth-url=https://OS_AUTH_URL/identity/v3
    tenant-id=OS_TENANT_ID
    domain-id=OS_DOMAIN_ID

    [LoadBalancer]
    subnet-id=SUBNET_ID
workers:
- name: nodes1
  replicas: 2
  providerSpec:
    labels:
      mylabel: 'nodes1'
    cloudProviderSpec:
      tags:
        tagKey: tagValue
    operatingSystem: 'ubuntu'
    operatingSystemSpec:
      distUpgradeOnBoot: true
```

</p></details><details><summary>Packet</summary><p>

Before you start you'll need a configuration file that defines how Kubernetes will be installed, e.g. what version will be used and what features will be enabled. For the configuration file reference run `kubeone config print --full`.

To get started you can use the following configuration. It'll install Kubernetes 1.14.2, create 1 worker node and deploy the
[external cloud controller manager][packet_ccm]. The external cloud controller manager takes care of providing correct information about nodes from the Packet API. KubeOne automatically populates information about the worker nodes from the [Terraform output][packet_tf_output].

Alternatively, you can set those information manually. As KubeOne is using Kubermatic [`machine-controller`][machine_controller] for creating worker nodes, see [Packet example manifest][packet_mc_example] for available options.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo

versions:
  kubernetes: "1.14.2"

cloudProvider:
  name: "packet"
  external: true

clusterNetwork:
  podSubnet: "192.168.0.0/16"
  serviceSubnet: "172.16.0.0/12"
```

**Note:** It's important to provide custom `clusterNetwork` settings in order to avoid colliding with private Packet network (which is `10.0.0.0/8`).

</p></details><details><summary>vSphere</summary><p>

Before you start you'll need a configuration file that defines how Kubernetes will be installed, e.g. what version will be used and what features will be enabled. For the configuration file reference run `kubeone config print --full`.

To get started you can use the following configuration. It'll install Kubernetes 1.14.2 and create 1 worker nodes. KubeOne automatically populates information about template, VM size and networking settings for worker nodes from the Terraform output. Alternatively, you can set those information manually. As KubeOne is using [Kubermatic `machine-controller`][7] for creating worker nodes, see [vSphere example manifest][8] for available options.

For vSphere you also need to provide a `cloud-config` file containing credentials, so vSphere Cloud Controller Manager works as expected. Make sure to replace sample values with real values.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo
versions:
  kubernetes: '1.14.2'
cloudProvider:
  name: 'vsphere'
  cloudConfig: |
    [Global]
    user = "<USERNAME>"
    password = "<PASSWORD>"
    port = "443"
    insecure-flag = "0"

    [VirtualCenter "1.1.1.1"]

    [Workspace]
    server = "1.1.1.1"
    datacenter = "dc-1"
    default-datastore="exsi-nas"
    resourcepool-path="kubeone"
    folder = "kubeone"

    [Disk]
    scsicontrollertype = pvscsi

    [Network]
    public-network = "NAT Network"  
```

</p></details>


## Cluster management

### Installing

Now that you have infrastructure you can proceed with installing Kubernetes using KubeOne.

Finally, we're going to install Kubernetes by using the `install` command and providing the configuration file and the Terraform output:

```bash
kubeone install config.yaml --tfjson tf.json
```

The installation process takes some time, usually 5-10 minutes. The output should look like the following one:

```
$ kubeone install config.yaml -t tf.json
INFO[13:15:31 EEST] Installing prerequisites…
INFO[13:15:32 EEST] Determine operating system…                   node=192.168.11.142
INFO[13:15:33 EEST] Determine operating system…                   node=192.168.11.139
INFO[13:15:34 EEST] Determine hostname…                           node=192.168.11.142
INFO[13:15:34 EEST] Creating environment file…                    node=192.168.11.142
INFO[13:15:34 EEST] Installing kubeadm…                           node=192.168.11.142 os=ubuntu
INFO[13:15:34 EEST] Determine operating system…                   node=192.168.11.140
INFO[13:15:36 EEST] Determine hostname…                           node=192.168.11.139
INFO[13:15:36 EEST] Creating environment file…                    node=192.168.11.139
INFO[13:15:36 EEST] Installing kubeadm…                           node=192.168.11.139 os=ubuntu
INFO[13:15:36 EEST] Determine hostname…                           node=192.168.11.140
INFO[13:15:36 EEST] Creating environment file…                    node=192.168.11.140
INFO[13:15:37 EEST] Installing kubeadm…                           node=192.168.11.140 os=ubuntu
INFO[13:16:45 EEST] Deploying configuration files…                node=192.168.11.139 os=ubuntu
INFO[13:16:45 EEST] Deploying configuration files…                node=192.168.11.140 os=ubuntu
INFO[13:17:03 EEST] Deploying configuration files…                node=192.168.11.142 os=ubuntu
INFO[13:17:04 EEST] Generating kubeadm config file…
INFO[13:17:06 EEST] Configuring certs and etcd on first controller…
INFO[13:17:06 EEST] Ensuring Certificates…                        node=192.168.11.139
INFO[13:17:14 EEST] Downloading PKI files…                        node=192.168.11.139
INFO[13:17:16 EEST] Creating local backup…                        node=192.168.11.139
INFO[13:17:16 EEST] Deploying PKI…
INFO[13:17:16 EEST] Uploading files…                              node=192.168.11.142
INFO[13:17:16 EEST] Uploading files…                              node=192.168.11.140
INFO[13:17:21 EEST] Configuring certs and etcd on consecutive controller…
INFO[13:17:21 EEST] Ensuring Certificates…                        node=192.168.11.142
INFO[13:17:21 EEST] Ensuring Certificates…                        node=192.168.11.140
INFO[13:17:27 EEST] Initializing Kubernetes on leader…
INFO[13:17:27 EEST] Running kubeadm…                              node=192.168.11.139
INFO[13:18:45 EEST] Joining controlplane node…
INFO[13:18:45 EEST] Waiting 30s to ensure main control plane components are up…  node=192.168.11.140
INFO[13:20:27 EEST] Waiting 30s to ensure main control plane components are up…  node=192.168.11.142
INFO[13:22:03 EEST] Copying Kubeconfig to home directory…         node=192.168.11.140
INFO[13:22:03 EEST] Copying Kubeconfig to home directory…         node=192.168.11.139
INFO[13:22:03 EEST] Copying Kubeconfig to home directory…         node=192.168.11.142
INFO[13:22:10 EEST] Building Kubernetes clientset…
INFO[13:22:16 EEST] Creating credentials secret…
INFO[13:22:16 EEST] Applying canal CNI plugin…
INFO[13:22:21 EEST] Installing machine-controller…
INFO[13:22:27 EEST] Installing machine-controller webhooks…
INFO[13:22:30 EEST] Waiting for machine-controller to come up…
INFO[13:23:15 EEST] Creating worker machines…
```

### Using

KubeOne automatically downloads the Kubeconfig file for the cluster. It's named as **\<cluster_name>-kubeconfig**, where **\<cluster_name>** is the name from your configuration. You can use it with kubectl such as

```bash
kubectl --kubeconfig=<cluster_name>-kubeconfig
```

or export the `KUBECONFIG` variable environment variable:
```bash
export KUBECONFIG=$PWD/<cluster_name>-kubeconfig
```

### Upgrade

### Removing

Before deleting a cluster you should clean up all MachineDeployments, so all worker nodes are deleted. You can do it with the `kubeone reset` command:

```bash
kubeone reset config.yaml --tfjson tf.json
```

This command will wait for all worker nodes to be gone.
