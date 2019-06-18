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
    endpoint = aws_lb.control_plane.dns_name
  }
}

output "kubeone_bastion" {
  value = aws_instance.bastion.public_ip
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name         = var.cluster_name
      cloud_provider       = "aws"
      private_address      = aws_instance.control_plane.*.private_ip
      ssh_agent_socket     = var.ssh_agent_socket
      ssh_port             = var.ssh_port
      ssh_private_key_file = var.ssh_private_key_file
      ssh_user             = var.ssh_username
    }
  }
}

output "kubeone_workers" {
  description = "Workers definitions, that will be transformed into MachineDeployment object"

  value = {
    # following outputs will be parsed by kubeone and automatically merged into
    # corresponding (by name) worker definition
    pool1 = {
      replicas        = 1
      sshPublicKeys   = [aws_key_pair.deployer.public_key]
      operatingSystem = var.worker_os
      operatingSystemSpec = {
        distUpgradeOnBoot = false
      }
      region           = var.aws_region
      ami              = local.ami
      availabilityZone = data.aws_availability_zones.available.names[0]
      instanceProfile  = aws_iam_instance_profile.workers.name
      securityGroupIDs = [aws_security_group.common.id]
      vpcId            = data.aws_vpc.selected.id
      subnetId         = aws_subnet.private[0].id
      instanceType     = var.worker_type
      diskSize         = 100
      diskType         = "gp2"
      ## Only application if diskType = io1
      diskIops         = 500
      tags = {
        "kubeone" = "pool1"
      }
    }
  }
  # provider specific fields:
  # see example under `cloudProviderSpec` section at: 
  # https://github.com/kubermatic/machine-controller/blob/master/examples/aws-machinedeployment.yaml
}

