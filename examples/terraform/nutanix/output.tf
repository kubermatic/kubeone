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
    endpoint = nutanix_virtual_machine.lb.nic_list.0.ip_endpoint_list.0.ip
    apiserver_alternative_names = var.apiserver_alternative_names
  }
}

output "ssh_commands" {
  value = formatlist("ssh -J ${var.bastion_user}@${nutanix_virtual_machine.lb.nic_list.0.ip_endpoint_list.0.ip} ${var.ssh_username}@%s", nutanix_virtual_machine.control_plane.*.nic_list.0.ip_endpoint_list.0.ip)
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name         = var.cluster_name
      cloud_provider       = "none"
      private_address      = nutanix_virtual_machine.control_plane.*.nic_list.0.ip_endpoint_list.0.ip
      ssh_agent_socket     = var.ssh_agent_socket
      ssh_port             = var.ssh_port
      ssh_private_key_file = var.ssh_private_key_file
      ssh_user             = var.ssh_username
      bastion              = nutanix_virtual_machine.lb.nic_list.0.ip_endpoint_list.0.ip
      bastion_port         = var.bastion_port
      bastion_user         = var.bastion_user
    }
  }
}

output "kubeone_workers" {
  description = "Workers definitions, that will be transformed into MachineDeployment object"

  value = {}
}

