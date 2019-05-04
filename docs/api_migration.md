# Migrating to the KubeOneCluster API

We implemented a new KubeOneCluster API in v0.6.0 that replaced the old configuration API.
As the new API introduces many breaking changes, you can **not** continue using old manifests
as of v0.6.0. To continue using KubeOne, you have to migrate old manifests to
new KubeOneCluster manifests.

The new API brings many possibilities and follows the Kubernetes API conventions. It allows us
to improve the user experience and ensure we can provide the
[backwards compatibility policy](./backwards_compatibility_policy.md).

To make migration to the new API easier, we implemented the `config migrate` command that migrates
old configuration manifests to KubeOneCluster manifests.

This document shows how to use the `config migrate` command and lists changes made in the new API.

:warning: **Note: Before you can use the `config migrate` command you need to make sure that your manifests
are updated for KubeOne versions v0.4.0 or newer. If not, please see
[the changelog](https://github.com/kubermatic/kubeone/blob/master/CHANGELOG.md) to see what's new and what
are required actions to migrate.** :warning:

## Using the `config migrate` command

The `config migrate` can automatically migrate your old manifests to new KubeOneCluster manifests.
The command takes path to the old configuration manifest and prints the converted manifest.
Once the manifest is migrated, it is validated against the new API to ensure that all fields are
correct and contain correct values.

```bash
kubeone config migrate config.yaml
```

For a `config.yaml` manifest that looks like the following one:

```yaml
name: demo
provider:
  name: aws
hosts:
  - public_address: 1.1.1.1
    private_address: 1.1.1.2
proxy:
  http_proxy: 1.1.2.1
  https_proxy: 1.1.2.2
  no_proxy: 1.1.2.3
features:
  pod_security_policy:
    enable: true
```

The `config migrate` command will print manifest such as:

```yaml
name: demo
hosts:
- publicAddress: 1.1.1.1
  privateAddress: 1.1.1.2
proxy:
  http: 1.1.2.1
  https: 1.1.2.2
  noProxy: 1.1.2.3
features:
  podSecurityPolicy:
    enable: true
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
cloudProvider:
  name: aws
```

**Note: The `apiVersion` and `kind` fields are not placed at the top automatically.
If you prefer, you can move them to the top manually.**

## The API Changelog

The new API introduces many changes to ensure the API follows the Kubernetes API conventions
and that we can provide as best as possible user experience.

The most important changes include:

* The API is now versioned,
* All fields are renamed to use the camel case notation,
* Some fields are renamed or replaced with new ones, so it's easier to understand what each
field is doing.

### `apiVersion` and `kind` fields

The `apiVersion` and `kind` fields must be added to the manifest. The `apiVersion` tells KubeOne
what API version you are using. When we change the API in a backwards incompatible way, we introduce
a new API version along with an automatic migration path. That ensures you can continue using the
old API as long as it's supported. To learn more for how long the old API versions are supported,
see the [backwards compatibility policy](./backwards_compatibility_policy.md).

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
```

### Changes at the root level

The following changes are introduced at the root level:

* `apiserver` is replaced with the **`apiEndpoint`** structure,
* `provider` is renamed to **`cloudProvider`**,
* `network` is renamed to **`clusterNetwork`**,
* `machine_controller` is renamed to **`machineController`**.

### `apiEndpoint` structure

The `apiserver` structure is replaced with the **`apiEndpoint`** structure which contains the following fields:

* **`host`** - address or hostname of the API endpoint (by default load balancer IP address or DNS),
* **`port`** - port used to access the Kubernetes API (by default `6443`).

### `hosts` structure

All fields in the `hosts` structure are renamed to use the camel case notation:

* `public_address` -> **`publicAddress`**
* `private_address` -> **`privateAddress`**
* `ssh_port` -> **`sshPort`**
* `ssh_username` -> **`sshUsername`**
* `ssh_private_key_file` -> **`sshPrivateKeyFile`**
* `ssh_agent_socket` -> **`sshAgentSocket`**

### `cloudProvider` structure

The `provider` structure is renamed to **`cloudProvider`** and the `cloud_config` field is
renamed to `cloudConfig` (camel case notation).

### `clusterNetwork` structure

The `network` structure is renamed to `clusterNetwork`. All fields are renamed to use the camel case
notation and a new field is added:

* `pod_subnet` -> **`podSubnet`**
* `service_subnet` -> **`serviceSubnet`**
* `node_port_range` -> **`nodePortRange`**
* [NEW] **`serviceDomainName`** (by default `cluster.local`)

### `proxy` structure

Fields in the `proxy` structure are renamed:

* `http_proxy` -> **`http`**
* `https_proxy` -> **`https`**
* `no_proxy` -> **`noProxy`**

### `features` structure

All fields in the `features` structure are renamed to use the camel case notation:

* `pod_security_policy` -> **`podSecurityPolicy`**
* `dynamic_audit_log` -> **`dynamicAuditLog`**
* `metrics_server` -> **`metricsServer`**
* `openid_connect` -> **`openidConnect`**

### `workers` structure

The `config` field is renamed to **`providerSpec`**.

### `credentials` structure

We're now storing credentials for the `machine-controller` and the external CCM at the root level,
in the **`credentials`** structure, instead of in `machineController.Credentials`.

## The library changelog

The following changes can affect using KubeOne as a Go library:

* Structures are renamed to use the camel case notation and some structures are changed or removed
(see above points for more details),
* `WorkerConfig.ProviderSpec.CloudProviderSpec` and `WorkerConfig.ProviderSpec.OperatingSystemSpec`
are taking `json.RawMessage` instead of `map[string]interface{}`,
* All fields in the `Features` structure are now pointers. All fields for enabling the feature are
called `Enable` and are type of `bool` (previous pointer on `bool`),
* `MachineController` field is now a pointer of `MachineControllerConfig`
(previous `non-pointer MachineControllerConfig`),
* `MachineController.Deploy` is now `bool` instead of pointer on `bool`.
