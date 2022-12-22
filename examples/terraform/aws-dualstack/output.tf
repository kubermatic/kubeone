/*
Copyright 2019 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

output "kubeone_api" {
  description = "kube-apiserver LB endpoint"

  value = {
    endpoint                    = local.kubeapi_endpoint
    apiserver_alternative_names = var.apiserver_alternative_names
  }
}

output "ssh_commands" {
  value = formatlist("ssh -J ${local.bastion_user}@${aws_instance.bastion.public_ip} ${local.ssh_username}@%s", aws_instance.control_plane.*.private_ip)
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name         = var.cluster_name
      cloud_provider       = "aws"
      private_address      = aws_instance.control_plane.*.private_ip
      ipv6_addresses       = aws_instance.control_plane.*.ipv6_addresses
      hostnames            = aws_instance.control_plane.*.private_dns
      operating_system     = var.os
      ssh_agent_socket     = var.ssh_agent_socket
      ssh_port             = var.ssh_port
      ssh_private_key_file = var.ssh_private_key_file
      ssh_user             = local.ssh_username
      ssh_hosts_keys       = var.ssh_hosts_keys
      bastion              = aws_instance.bastion.public_ip
      bastion_port         = var.bastion_port
      bastion_user         = local.bastion_user
      bastion_host_key     = var.bastion_host_key
      labels               = var.control_plane_labels
      # uncomment to following to set those kubelet parameters. More into at:
      # https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/
      # kubelet            = {
      #   system_reserved = "cpu=200m,memory=200Mi"
      #   kube_reserved   = "cpu=200m,memory=300Mi"
      #   eviction_hard   = ""
      #   max_pods        = 110
      # }
    }
  }
}

output "kubeone_static_workers" {
  description = "Static worker config"

  value = {
    workers1 = {
      private_address      = aws_instance.static_workers1.*.private_ip
      hostnames            = aws_instance.static_workers1.*.private_dns
      operating_system     = var.os
      ssh_agent_socket     = var.ssh_agent_socket
      ssh_port             = var.ssh_port
      ssh_private_key_file = var.ssh_private_key_file
      ssh_user             = local.ssh_username
      bastion              = aws_instance.bastion.public_ip
      bastion_port         = var.bastion_port
      bastion_user         = local.bastion_user
    }
  }
}

output "kubeone_workers" {
  description = "Workers definitions, that will be transformed into MachineDeployment object"

  value = {
    # following outputs will be parsed by kubeone and automatically merged into
    # corresponding (by name) worker definition
    "${var.cluster_name}-${local.zoneA}" = {
      replicas = var.initial_machinedeployment_replicas
      providerSpec = {
        annotations = {
          "k8c.io/operating-system-profile"                           = var.initial_machinedeployment_operating_system_profile
          "cluster.k8s.io/cluster-api-autoscaler-node-group-min-size" = tostring(local.cluster_autoscaler_min_replicas)
          "cluster.k8s.io/cluster-api-autoscaler-node-group-max-size" = tostring(local.cluster_autoscaler_max_replicas)
        }
        sshPublicKeys   = local.worker_deploy_ssh_key
        operatingSystem = local.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot   = false
          provisioningUtility = var.provisioning_utility
        }
        labels = {
          isSpotInstance = format("%t", local.initial_machinedeployment_spotinstances)
        }
        # nodeAnnotations are applied on resulting Node objects
        # nodeAnnotations = {
        #   "key" = "value"
        # }
        # machineObjectAnnotations are applied on resulting Machine objects
        # uncomment to following to set those kubelet parameters. More into at:
        # https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/
        # machineObjectAnnotations = {
        #   "v1.kubelet-config.machine-controller.kubermatic.io/SystemReserved" = "cpu=200m,memory=200Mi"
        #   "v1.kubelet-config.machine-controller.kubermatic.io/KubeReserved"   = "cpu=200m,memory=300Mi"
        #   "v1.kubelet-config.machine-controller.kubermatic.io/EvictionHard"   = ""
        #   "v1.kubelet-config.machine-controller.kubermatic.io/MaxPods"        = "110"
        # }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at:
          # https://github.com/kubermatic/machine-controller/blob/main/examples/aws-machinedeployment.yaml
          region           = var.aws_region
          ami              = local.ami
          availabilityZone = local.zoneA
          instanceProfile  = aws_iam_instance_profile.profile.name
          securityGroupIDs = [aws_security_group.common.id]
          vpcId            = data.aws_vpc.selected.id
          subnetId         = local.subnets[local.zoneA]
          instanceType     = var.worker_type
          assignPublicIP   = true
          diskSize         = 50
          diskType         = "gp2"
          ## Only applicable if diskType = io1
          diskIops       = 500
          isSpotInstance = local.initial_machinedeployment_spotinstances
          ## Only applicable if isSpotInstance is true
          spotInstanceConfig = {
            maxPrice = format("%f", var.initial_machinedeployment_spotinstances_max_price)
          }
          ebsVolumeEncrypted = false
          tags = {
            "${var.cluster_name}-workers" = ""
          }
        }
        network = {
          ipFamily = var.ip_family
        }
      }
    }

    "${var.cluster_name}-${local.zoneB}" = {
      replicas = var.initial_machinedeployment_replicas
      providerSpec = {
        annotations = {
          "k8c.io/operating-system-profile"                           = var.initial_machinedeployment_operating_system_profile
          "cluster.k8s.io/cluster-api-autoscaler-node-group-min-size" = tostring(local.cluster_autoscaler_min_replicas)
          "cluster.k8s.io/cluster-api-autoscaler-node-group-max-size" = tostring(local.cluster_autoscaler_max_replicas)
        }
        sshPublicKeys   = local.worker_deploy_ssh_key
        operatingSystem = local.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot   = false
          provisioningUtility = var.provisioning_utility
        }
        labels = {
          isSpotInstance = format("%t", local.initial_machinedeployment_spotinstances)
        }
        # nodeAnnotations are applied on resulting Node objects
        # nodeAnnotations = {
        #   "key" = "value"
        # }
        # machineObjectAnnotations are applied on resulting Machine objects
        # uncomment to following to set those kubelet parameters. More into at:
        # https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/
        # machineObjectAnnotations = {
        #   "v1.kubelet-config.machine-controller.kubermatic.io/SystemReserved" = "cpu=200m,memory=200Mi"
        #   "v1.kubelet-config.machine-controller.kubermatic.io/KubeReserved"   = "cpu=200m,memory=300Mi"
        #   "v1.kubelet-config.machine-controller.kubermatic.io/EvictionHard"   = ""
        #   "v1.kubelet-config.machine-controller.kubermatic.io/MaxPods"        = "110"
        # }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at:
          # https://github.com/kubermatic/machine-controller/blob/main/examples/aws-machinedeployment.yaml
          region           = var.aws_region
          ami              = local.ami
          availabilityZone = local.zoneB
          instanceProfile  = aws_iam_instance_profile.profile.name
          securityGroupIDs = [aws_security_group.common.id]
          vpcId            = data.aws_vpc.selected.id
          subnetId         = local.subnets[local.zoneB]
          instanceType     = var.worker_type
          assignPublicIP   = true
          diskSize         = 50
          diskType         = "gp2"
          ## Only applicable if diskType = io1
          diskIops       = 500
          isSpotInstance = local.initial_machinedeployment_spotinstances
          ## Only applicable if isSpotInstance is true
          spotInstanceConfig = {
            maxPrice = format("%f", var.initial_machinedeployment_spotinstances_max_price)
          }
          ebsVolumeEncrypted = false
          tags = {
            "${var.cluster_name}-workers" = ""
          }
        }
        network = {
          ipFamily = var.ip_family
        }
      }
    }

    "${var.cluster_name}-${local.zoneC}" = {
      replicas = var.initial_machinedeployment_replicas
      providerSpec = {
        annotations = {
          "k8c.io/operating-system-profile"                           = var.initial_machinedeployment_operating_system_profile
          "cluster.k8s.io/cluster-api-autoscaler-node-group-min-size" = tostring(local.cluster_autoscaler_min_replicas)
          "cluster.k8s.io/cluster-api-autoscaler-node-group-max-size" = tostring(local.cluster_autoscaler_max_replicas)
        }
        sshPublicKeys   = local.worker_deploy_ssh_key
        operatingSystem = local.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot   = false
          provisioningUtility = var.provisioning_utility
        }
        labels = {
          isSpotInstance = format("%t", local.initial_machinedeployment_spotinstances)
        }
        # nodeAnnotations are applied on resulting Node objects
        # nodeAnnotations = {
        #   "key" = "value"
        # }
        # machineObjectAnnotations are applied on resulting Machine objects
        # uncomment to following to set those kubelet parameters. More into at:
        # https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/
        # machineObjectAnnotations = {
        #   "v1.kubelet-config.machine-controller.kubermatic.io/SystemReserved" = "cpu=200m,memory=200Mi"
        #   "v1.kubelet-config.machine-controller.kubermatic.io/KubeReserved"   = "cpu=200m,memory=300Mi"
        #   "v1.kubelet-config.machine-controller.kubermatic.io/EvictionHard"   = ""
        #   "v1.kubelet-config.machine-controller.kubermatic.io/MaxPods"        = "110"
        # }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at:
          # https://github.com/kubermatic/machine-controller/blob/main/examples/aws-machinedeployment.yaml
          region           = var.aws_region
          ami              = local.ami
          availabilityZone = local.zoneC
          instanceProfile  = aws_iam_instance_profile.profile.name
          securityGroupIDs = [aws_security_group.common.id]
          vpcId            = data.aws_vpc.selected.id
          subnetId         = local.subnets[local.zoneC]
          instanceType     = var.worker_type
          assignPublicIP   = true
          diskSize         = 50
          diskType         = "gp2"
          ## Only applicable if diskType = io1
          diskIops       = 500
          isSpotInstance = local.initial_machinedeployment_spotinstances
          ## Only applicable if isSpotInstance is true
          spotInstanceConfig = {
            maxPrice = format("%f", var.initial_machinedeployment_spotinstances_max_price)
          }
          ebsVolumeEncrypted = false
          tags = {
            "${var.cluster_name}-workers" = ""
          }
        }
        network = {
          ipFamily = var.ip_family
        }
      }
    }
  }
}
