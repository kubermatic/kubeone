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

locals {
  cluster_autoscaler_min_replicas = var.cluster_autoscaler_min_replicas > 0 ? var.cluster_autoscaler_min_replicas : var.initial_machinedeployment_replicas
  cluster_autoscaler_max_replicas = var.cluster_autoscaler_max_replicas > 0 ? var.cluster_autoscaler_max_replicas : var.initial_machinedeployment_replicas
}

output "kubeone_api" {
  description = "kube-apiserver LB endpoint"

  value = {
    endpoint                    = openstack_networking_floatingip_v2.kube_apiserver.address
    apiserver_alternative_names = var.apiserver_alternative_names
  }
}

output "ssh_commands" {
  value = formatlist("ssh -J ${var.bastion_user}@${openstack_networking_floatingip_v2.bastion.address} ${var.ssh_username}@%s", openstack_compute_instance_v2.control_plane[*].access_ip_v4)
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name         = var.cluster_name
      cloud_provider       = "openstack"
      private_address      = openstack_compute_instance_v2.control_plane[*].access_ip_v4
      hostnames            = openstack_compute_instance_v2.control_plane[*].name
      ssh_agent_socket     = var.ssh_agent_socket
      ssh_port             = var.ssh_port
      ssh_private_key_file = var.ssh_private_key_file
      ssh_user             = var.ssh_username
      bastion              = openstack_networking_floatingip_v2.bastion.address
      bastion_port         = var.bastion_port
      bastion_user         = var.bastion_user
      ssh_hosts_keys       = var.ssh_hosts_keys
      bastion_host_key     = var.bastion_host_key
    }
  }
}

output "kubeone_workers" {
  description = "Workers definitions, that will be transformed into MachineDeployment object"

  value = {
    # following outputs will be parsed by kubeone and automatically merged into
    # corresponding (by name) worker definition
    "${var.cluster_name}-pool1" = {
      replicas = var.initial_machinedeployment_replicas
      providerSpec = {
        annotations = {
          "k8c.io/operating-system-profile"                           = var.initial_machinedeployment_operating_system_profile
          "cluster.k8s.io/cluster-api-autoscaler-node-group-min-size" = tostring(local.cluster_autoscaler_min_replicas)
          "cluster.k8s.io/cluster-api-autoscaler-node-group-max-size" = tostring(local.cluster_autoscaler_max_replicas)
        }
        sshPublicKeys   = [file(var.ssh_public_key_file)]
        operatingSystem = var.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot = false
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
          # https://github.com/kubermatic/machine-controller/blob/main/examples/openstack-machinedeployment.yaml
          image          = data.openstack_images_image_v2.image.name
          flavor         = var.worker_flavor
          securityGroups = [openstack_networking_secgroup_v2.securitygroup.name]
          network        = openstack_networking_network_v2.network.name
          subnet         = openstack_networking_subnet_v2.subnet.name
          # floatingIpPool = var.external_network_name
          # Optional: If set, the rootDisk will be a volume.
          # Otherwise, the rootDisk will be on ephemeral storage and its size will
          # be derived from the flavor
          # rootDiskSizeGB = 50
          # Optional: limit how many volumes can be attached to a node
          # nodeVolumeAttachLimit = 25
          tags = {
            "${var.cluster_name}-workers" = "pool1"
          }
        }
      }
    }
  }
}
