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
    endpoint = hcloud_load_balancer.load_balancer.ipv4
    additional_names = var.additional_names
  }
}

output "ssh_commands" {
  value = formatlist("ssh ${var.ssh_username}@%s", hcloud_server.control_plane.*.ipv4_address)
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      hostnames            = hcloud_server.control_plane.*.name
      cluster_name         = var.cluster_name
      cloud_provider       = "hetzner"
      private_address      = hcloud_server_network.control_plane.*.ip
      public_address       = hcloud_server.control_plane.*.ipv4_address
      network_id           = hcloud_network.net.id
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
      replicas = var.workers_replicas
      providerSpec = {
        sshPublicKeys   = [file(var.ssh_public_key_file)]
        operatingSystem = var.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot = false
        }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at:
          # https://github.com/kubermatic/machine-controller/blob/master/examples/hetzner-machinedeployment.yaml
          serverType = var.worker_type
          location   = var.datacenter
          image      = var.image
          networks = [
            hcloud_network.net.id
          ]
          # Datacenter (optional)
          # datacenter = ""
          labels = {
            "kubeone_cluster_name"        = var.cluster_name
            "${var.cluster_name}-workers" = "pool1"
          }
        }
      }
    }
  }
}
