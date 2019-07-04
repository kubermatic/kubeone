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
    endpoint = openstack_networking_floatingip_v2.lb.address
  }
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name         = var.cluster_name
      cloud_provider       = "openstack"
      private_address      = openstack_compute_instance_v2.control_plane.*.access_ip_v4
      public_address       = openstack_networking_floatingip_v2.control_plane.*.address
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
    "${var.cluster_name}-pool1" = {
      replicas = 1
      providerSpec = {
        sshPublicKeys   = [file(var.ssh_public_key_file)]
        operatingSystem = var.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot = false
        }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at:
          # https://github.com/kubermatic/machine-controller/blob/master/examples/openstack-machinedeployment.yaml
          image          = var.image
          flavor         = var.worker_flavor
          securityGroups = [openstack_networking_secgroup_v2.securitygroup.name]
          floatingIPPool = var.external_network_name
          network        = openstack_networking_network_v2.network.name
          subnet         = openstack_networking_subnet_v2.subnet.name
          # Optional: If set, the rootDisk will be a volume. 
          # Otherwise, the rootDisk will be on ephemeral storage and its size will
          # be derived from the flavor
          rootDiskSizeGB = 10
          tags = {
            "${var.cluster_name}-workers" = "pool1"
          }
        }
      }
    }
  }
}

