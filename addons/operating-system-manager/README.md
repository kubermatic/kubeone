# Operating-system-manager addon

Operating system manager can be used to create and manage worker node configurations in a kubernetes cluster. It is responsibile for managing user-data for worker machines in the cluster.

**Note:** Existing worker machines will not be migrated to use OSM automatically. User needs to update the `MachineDeployments` manually or **simply delete the machines** to ensure that the machines are consuming configurations created by OSM.

## Using legacy machine-controller for userdata

OSM is enabled by default starting from KubeOne `v1.5.0`. To fallback to, the now deprecated, implementation of userdata in machine-controller we can disable OSM as follows:

```yaml
apiVersion: kubeone.k8c.io/v1beta2
kind: KubeOneCluster
versions:
  kubernetes: 1.26.0
addons:
  enable: true
operatingSystemManager:
  deploy: false
```

For more details, please refer to the [Operating System Manager documentation](https://github.com/kubermatic/operating-system-manager#readme).
