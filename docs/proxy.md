# Proxy support

Annotated example config with enabled PROXY support
```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo-cluster
...
clusterNetwork:
  # following config values are defaults, and used only for demonstration
  podSubnet: "10.244.0.0/16"
  serviceSubnet: "10.96.0.0/12"
  serviceDomainName: "cluster.local"

# Proxy settings will be applied to:
# * kubeone remote commands over SSH
# * package managers (YUM/APT) on control-plane VMs
# * docker
# * kubelet
# * machine-controller
# * worker nodes kubelets and docker
proxy:
  # main proxy server
  http: "http://proxy.example.com"
  https: "http://proxy.example.com"

  # Following resources are IN-CLUSTER, and are AUTO-POPULATED to the noProxy,
  # you DON'T NEED to specify them explicitly.
  # * 127.0.0.1/8     — loopback
  # * localhost       — loopback
  # * cluster.local   — value from clusterNetwork.serviceDomainName
  # * 10.244.0.0/16   — value from clusterNetwork.podSubnet
  # * 10.96.0.0/12    — value from clusterNetwork.serviceSubnet
  #
  # Setting anything additionally to noProxy will be added after the above
  # noProxy settings
  # 172.20.0.0/16 and 172.25.0.0/16 are examples of external resources to the
  # cluster, but internal to organization network, thus requests to them should
  # not hit proxy
  noProxy: "172.25.0.0/16,172.20.0.0/16"
```
