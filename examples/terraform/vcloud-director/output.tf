/*
Copyright 2022 The KubeOne Authors.

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
    endpoint                    = local.public_ip
    apiserver_alternative_names = var.apiserver_alternative_names
  }
}

output "ssh_commands" {
  value = formatlist("ssh -J ${var.bastion_user}@${local.public_ip} ${var.ssh_username}@%s", vcd_vapp_vm.control_plane.*.network.0.ip)
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name = var.cluster_name
      # TODO(@ahmedwaleedmalik): Cloud provider is not implemented yet
      cloud_provider       = "noop"
      private_address      = vcd_vapp_vm.control_plane.*.network.0.ip
      hostnames            = vcd_vapp_vm.control_plane.*.name
      ssh_agent_socket     = var.ssh_agent_socket
      ssh_port             = var.ssh_port
      ssh_private_key_file = var.ssh_private_key_file
      ssh_user             = var.ssh_username
      bastion              = local.public_ip
      bastion_port         = var.bastion_port
      bastion_user         = var.bastion_user
    }
  }
}