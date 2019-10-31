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

output "ami" {
  description = "AMI that was found initially"
  value       = local.ami
}

output "kubeone_api" {
  description = "kube-apiserver LB endpoint"

  value = {
    endpoint = aws_lb.control_plane.dns_name
  }
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name         = local.cluster_name
      cloud_provider       = "aws"
      private_address      = aws_instance.control_plane.*.private_ip
      hostnames            = aws_instance.control_plane.*.private_dns
      ssh_agent_socket     = var.ssh_agent_socket
      ssh_port             = var.ssh_port
      ssh_private_key_file = var.ssh_private_key_file
      ssh_user             = var.ssh_username
      bastion              = aws_instance.bastion.public_ip
      bastion_port         = var.bastion_port
      bastion_user         = var.bastion_user
    }
  }
}

output "proxy" {
  description = "Proxy settings"
  value = {
    http    = ""
    https   = ""
    noProxy = "${aws_lb.control_plane.dns_name}"
  }
}

output "kubeone_workers" {
  description = "Workers definitions, that will be transformed into MachineDeployment object"

  value = {
    # following outputs will be parsed by kubeone and automatically merged into
    # corresponding (by name) worker definition
    "${local.cluster_name}-${local.zoneA}" = {
      replicas = 1
      providerSpec = {
        sshPublicKeys   = [aws_key_pair.deployer.public_key]
        operatingSystem = var.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot = var.dist_upgrade_on_boot
        }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at:
          # https://github.com/kubermatic/machine-controller/blob/master/examples/aws-machinedeployment.yaml
          region           = var.aws_region
          ami              = local.ami
          availabilityZone = local.zoneA
          instanceProfile  = aws_iam_instance_profile.workers.name
          securityGroupIDs = [aws_security_group.common.id]
          vpcId            = data.aws_vpc.selected.id
          subnetId         = local.subnets["private"][local.zoneA]
          instanceType     = var.worker_type
          diskSize         = 100
          diskType         = "gp2"
          assignPublicIP   = false
          ## Only application if diskType = io1
          diskIops = 500
          tags = {
            "${local.cluster_name}-workers" = "${local.zoneA}"
          }
        }
      }
    }
    "${local.cluster_name}-${local.zoneB}" = {
      replicas = 1
      providerSpec = {
        sshPublicKeys   = [aws_key_pair.deployer.public_key]
        operatingSystem = var.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot = var.dist_upgrade_on_boot
        }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at:
          # https://github.com/kubermatic/machine-controller/blob/master/examples/aws-machinedeployment.yaml
          region           = var.aws_region
          ami              = local.ami
          availabilityZone = local.zoneB
          instanceProfile  = aws_iam_instance_profile.workers.name
          securityGroupIDs = [aws_security_group.common.id]
          vpcId            = data.aws_vpc.selected.id
          subnetId         = local.subnets["private"][local.zoneB]
          instanceType     = var.worker_type
          diskSize         = 100
          diskType         = "gp2"
          assignPublicIP   = false
          ## Only application if diskType = io1
          diskIops = 500
          tags = {
            "${local.cluster_name}-workers" = "${local.zoneB}"
          }
        }
      }
    }
    "${local.cluster_name}-${local.zoneC}" = {
      replicas = 1
      providerSpec = {
        sshPublicKeys   = [aws_key_pair.deployer.public_key]
        operatingSystem = var.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot = var.dist_upgrade_on_boot
        }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at:
          # https://github.com/kubermatic/machine-controller/blob/master/examples/aws-machinedeployment.yaml
          region           = var.aws_region
          ami              = local.ami
          availabilityZone = local.zoneC
          instanceProfile  = aws_iam_instance_profile.workers.name
          securityGroupIDs = [aws_security_group.common.id]
          vpcId            = data.aws_vpc.selected.id
          subnetId         = local.subnets["private"][local.zoneC]
          instanceType     = var.worker_type
          diskSize         = 100
          diskType         = "gp2"
          assignPublicIP   = false
          ## Only application if diskType = io1
          diskIops = 500
          tags = {
            "${local.cluster_name}-workers" = "${local.zoneC}"
          }
        }
      }
    }
  }
}
