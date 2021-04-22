# Encryption Providers for encrypted secrets at rest

**Auther**: Mohamed Elsayed (@moelsayed)
**Status**: Implemented


## Abstract

By default, all Kubernetes secret objects are stored on disk in plain text inside etcd. The [Encryption Providers](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/
) feature was added to Kubernets starting with version 1.13. 

At rest data encryption is a requirement for security compliance and adds an additional layer of security for secret data, especially when etcd nodes are separated from the control plan and in off-node backups. 

KubeOne needs to support this feature natively. Meaning the user should be able to enable, disable the feature and rotate keys when needed without having to apply any actions manually. 

## Goals

* Provide a safe path to enable/disable Encryption Providers.
* Support atomic(?) rotation for existing keys.
* Support custom configuration files and external KMS.
* Rewriting all secret resources (no just secrets) after enable/disable/rotate operations.

## Non-Goals

* Deploy External KMS.
* Safely rotate configuration when a custom configuration file is used. 

## Challenges

The feature has a lot of moving parts; as it requires performing a specific sequence of actions, including changing the KubeAPI configuration, restarting KubeAPI and rewriting all secret resources to apply the encryption. This requires the implementation to be as idempotent as possible with ability to rollback on failure, with out breaking the cluster. 

## Implementation

Unfortunately, it's not possible to simply update the KubeAPI configuration and expect the configuration to reconcile. KubeOne will have to _read_ the _current_ configuration on the cluster, _mutate_ it based on the _required_ state and then apply it. Additionally, KubeOne will have to be able to revert changes on any errors and recover safely if the process is interrupted at any point.

The configuration for this will be added under `features` in the KubeOneCluster spec:

```yaml
apiVersion: kubeone.io/v1beta1
kind: KubeOneCluster
features:
  encryptionProviders:
    enabled: true
    customEncryptionConfiguration: |
      apiVersion: apiserver.config.k8s.io/v1
      kind: EncryptionConfiguration
      resources:
      - resources:
        - secrets
        providers:
        - identity: {}
        - aescbc:
            keys:
            - name: key1
            secret: <BASE 64 ENCODED SECRET>
```

To allow users to rotate the keys, a new flag will be added to the `apply` command:

```bash
--rotate-encryption-key      automatically rotate encryption provider key
```

### pre-flight checks

 * Cluster is healthy.
 * Current Encryption Providers state/configuration is valid and identical on all control plane nodes.

### Enable Encryption Providers for new cluster

* Generate a valid configuration file with the `identity` provider set last.
* Sync the configuration file to all Control Plane nodes. 
* Set the required KubeAPI configuration and deploy KubeAPI.

### Enable Encryption Providers for existing cluster

* Ensure there is no Encryption Provider Config (manually added by the user, broken previous enable process, etc..) present.
* Generate a valid configuration file with the `identity` provider set last.
* Sync the configuration file to all Control Plane nodes. 
* Update and restart KubeAPI on all nodes.
* Rewrite secrets to ensure they are encrypted successfully.

### Disable Encryption Providers for existing cluster 

* Read the current active Encryption Provider configuration from control plane nodes.
* Mutate the configuration to add `identity` provider first and the active provider last.
* Sync the configuration file to all Control Plane nodes. 
* Restart KubeAPI on all control plane nodes.
* Rewrite secrets to ensure they are decrypted successfully.
* Update KubeAPI configuration to remove the Encryption Provider configuration and restart KubeAPI on all control plane nodes. 
* Remove the old configuration file from all control plane nodes.

### Rotate Encryption Provider Key for existing cluster

* Read the current active Encryption Provider configuration from control plane nodes.
* Generate a new encryption key.
* Mutate the configuration file to include the new key first, current key second and `identity` last.
* Sync the updated configuration file to all control plane nodes and restart KubeAPI.
* Rewrite all secrets to ensure they are encrypted with the new key.
* Mutate the configuration file again to remove the old key.
* Sync the updated configuration file to all control plane nodes and restart KubeAPI.

### Apply Custom Encryption Provider file
This use case is useful for users who would like to utilize an external KMS provider or specify resources other than secrets for encryption. In this case, KubeOne will not manage the content of the file, it will only validate it to make sure it's syntactically valid. Additionally, KubeOne will not rewrite the resources in this case. 

* Ensure the configuration file is valid. 
* Sync the configuration file to all control plane nodes.
* Restart KubeAPI on all nodes. 

Additionally, if an external KMS is used, KubeOne will detect that and add a specific file mount for the KMS unix socket file in the KubeAPI static pod spec. This is necessary to allow KubeAPI to communicate with the external KMS service.
## Tasks & effort

* Implement the needed pre-flight checks.
* Implement validation for Encryption Provider configuration files. 
* Implement the workflow for each use case. 
* Add e2e tests for each workflow.
* Add documentation for the feature. 