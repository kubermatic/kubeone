# Operating-system-manager addon

Operating system manager can be used to create and manage worker node configurations in a kubernetes cluster.

Once enabled, this addon will take over the responsibility for managing user-data for worker machines in the cluster.

**Note:** Existing worker machines will not be rotated if this addon is enabled after cluster creation.

## Enabling the addon

OSM addon can be enabled as follows:

```yaml
apiVersion: kubeone.k8c.io/v1beta2
kind: KubeOneCluster

versions:
  kubernetes: 1.22.9

addons:
  enable: true
  addons:
  - name: operating-system-manager
```


For more details, please refer to the [Operating System Manager documentation](https://github.com/kubermatic/operating-system-manager#readme).