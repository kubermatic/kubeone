# Calico addon (with VXLAN)

## Setup

Since this addon is now shipped together with KubeOne, it's possible to simply
configure it in kubeone.yaml.

Example kubeone config:

```yaml
apiVersion: kubeone.io/v1beta1
kind: KubeOneCluster

versions:
  kubernetes: 1.20.4

clusterNetwork:
  cni:
    external: {}

addons:
  enable: true
  addons:
  - name: calico-vxlan
```

## Custom MTU

MTU is set to 0 by default and is autodetected by the calico itself, but in case
when you'd like to set own custom MTU it's possible to use [addon params mechanism][addon_params]:

```yaml
apiVersion: kubeone.io/v1beta1
kind: KubeOneCluster

versions:
  kubernetes: 1.20.4

clusterNetwork:
  cni:
    external: {}

addons:
  enable: true
  addons:
  - name: calico-vxlan
    params:
      MTU: 1400 # custom MTU
```

[addon_params]: https://docs.kubermatic.com/kubeone/v1.3/guides/addons/#parameters
