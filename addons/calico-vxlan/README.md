# Calico addon (with VXLAN)

## Setup

```shell
mkdir addons
curl https://raw.githubusercontent.com/kubermatic/kubeone/master/addons/calico-vxlan/calico-vxlan.yaml > addons/calico-vxlan.yaml
```

## MTU

Please edit the `addons/calico-vxlan.yaml` and change `veth_mtu: "1450"` to appropriate MTU size. Please see more
documentation how to find MTU size for your cluster https://docs.projectcalico.org/networking/mtu#determine-mtu-size.

Example AWS kubeone config to use Calico addon.

```yaml
apiVersion: kubeone.io/v1beta1
kind: KubeOneCluster

versions:
  kubernetes: 1.18.5

cloudProvider:
  aws: {}

clusterNetwork:
  cni:
    external: {}

addons:
  enable: true
  path: "./addons"
```
