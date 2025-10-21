# Full nodes lifecycle management.

**Status:** **Draft**
**Created:** 2025-05-13
**Last updated:** 2025-05-13
**Author:** Artiom Diomin ([@kron4eg](https://github.com/kron4eg))

## Abstract

## API changes

```yaml
apiVersion: kubeone.k8c.io/v1beta2
kind: KubeOneCluster
name: test1

versions:
  kubernetes: 1.32.2

apiEndpoint:
  host: my-custom.dns.com

cloudProvider:
  hetzner:
    controlPlane:
      loadBalancer:
        region: eu-center1
        name: "${cluster-name}-api-lb"

controlPlane:
  nodeSets:
    - name: cp
      replicas: 3
      nodeSettings:
        labels: {}
        annotations: []
        taints: []
        kubelet: {}
      operatingSystem: ubuntu
      operatingSystemSpec:
        distUpgradeOnBoot: false
      ssh:
        publicKeys:
          - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIE1gO0bUOvixm2IOcqIlk4/zR0pCHBHDk4HKfCLtqOis sysop
        username: ubuntu
      cloudProviderSpec:
        location:
        image:
        serverType:
        networks: []
        labels:
          # following tags will be automatically added to the instance, for later service discovery
          # kubeone-created-on: "<TIMESTAMP>"
          # kubeone-role: control-plane
          # kubeone-cluster: "${cluster-name}"
```

## Operations

### Cluster Bootstrap

### Normal Operations

### Abnormal Operations

## Machine Controller and OSM changes
