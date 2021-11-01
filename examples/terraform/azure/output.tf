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
    endpoint = azurerm_public_ip.lbip.ip_address
    apiserver_alternative_names = var.apiserver_alternative_names
  }
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name         = var.cluster_name
      cloud_provider       = "azure"
      private_address      = azurerm_network_interface.control_plane.*.private_ip_address
      public_address       = data.azurerm_public_ip.control_plane.*.ip_address
      hostnames            = formatlist("${var.cluster_name}-cp-%d", [0, 1, 2])
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
      replicas = var.initial_machinedeployment_replicas
      providerSpec = {
        sshPublicKeys   = [file(var.ssh_public_key_file)]
        operatingSystem = var.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot = false
        }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at: 
          # https://github.com/kubermatic/machine-controller/blob/master/examples/azure-machinedeployment.yaml
          assignPublicIP    = true
          availabilitySet   = azurerm_availability_set.avset_workers.name
          location          = var.location
          resourceGroup     = azurerm_resource_group.rg.name
          routeTableName    = azurerm_route_table.rt.name
          securityGroupName = azurerm_network_security_group.sg.name
          subnetName        = azurerm_subnet.subnet.name
          vmSize            = var.worker_vm_size
          vnetName          = azurerm_virtual_network.vpc.name
          # Custom Image ID (optional)
          # imageID = ""
          # Size of the operating system disk (optional)
          # osDiskSize = 100
          # Size of the data disk (optional)
          # dataDiskSize = 100
          # Zones (optional)
          # Represents Availability Zones is a high-availability offering
          # that protects your applications and data from datacenter failures.
          # zones = {
          #   "1"
          # }
          tags = {
            "${var.cluster_name}-workers" = "pool1"
          }
        }
      }
    }
  }
}

