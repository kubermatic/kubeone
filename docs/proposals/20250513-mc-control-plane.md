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
  aws:
    controlPlane:
      loadBalancer:
        region: eu-center1
        # a EC2 loadBalancer, should exist before the first run
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
          - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIE1gO0bUOvixm2IOcqIlk4/zR0pCHBHDk4HKfCLtqOis artioms
        username: ubuntu
        bastion: 1.1.1.1
        bastionUser: ubuntu
      cloudProviderSpec:
        region: eu-center1
        availabilityZones:
          a: subnet-123
          b: subnet-456
          c: subnet-789
        ami: ami-123
        instanceProfile: my-control-plane-profile
        securityGroupIDs:
          - sg-111-common
          - sg-222-control-plane
        vpcId: vpc-123
        instanceType: t3a.medium
        assignPublicIP: true
        diskSize: 50
        ebsVolumeEncrypted: false
        # service discovery tags will be automatically added to the instances
        tags:
          # following tags will be automatically added to the instance, for later service discovery
          kubeone-created-on: "<TIMESTAMP>"
          kubeone-role: control-plane
          kubeone-cluster: "${cluster-name}"
```

## Operations

### Cluster Bootstrap

### Normal Operations

### Abnormal Operations

## Machine Controller and OSM changes
