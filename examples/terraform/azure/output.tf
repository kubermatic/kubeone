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
      ssh_user             = local.ssh_username
      ssh_hosts_keys       = var.ssh_hosts_keys
      bastion_host_key     = var.bastion_host_key
    }
  }
}

output "kubeone_workers" {
  description = "Workers definitions, that will be transformed into MachineDeployment object"
  sensitive   = true

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
        operatingSystem = local.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot               = false
          disableAutoUpdate               = var.disable_auto_update
          rhelSubscriptionManagerUser     = var.rhsm_username
          rhelSubscriptionManagerPassword = var.rhsm_password
          rhsmOfflineToken                = var.rhsm_offline_token
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
          # https://github.com/kubermatic/machine-controller/blob/main/examples/azure-machinedeployment.yaml
          location      = var.location
          resourceGroup = azurerm_resource_group.rg.name
          # vnetResourceGroup     = ""
          vmSize          = var.worker_vm_size
          vnetName        = azurerm_virtual_network.vpc.name
          subnetName      = azurerm_subnet.subnet.name
          loadBalancerSku = "Standard"
          routeTableName  = azurerm_route_table.rt.name
          availabilitySet = azurerm_availability_set.avset.name
          # assignAvailabilitySet = true/false
          securityGroupName = azurerm_network_security_group.sg.name
          assignPublicIP    = true
          publicIPSKU       = "Standard"
          imageReference    = var.os != "rhel" ? var.image_references[var.os].image : null
          imagePlan         = length(var.image_references[var.os].plan) > 0 && var.os != "rhel" ? var.image_references[var.os].plan[0] : null
          # Zones (optional)
          # Represents Availability Zones is a high-availability offering
          # that protects your applications and data from datacenter failures.
          # zones = {
          #   "1"
          # }
          # Custom Image ID (optional)
          # imageID = ""
          # Size of the operating system disk (optional)
          # osDiskSize = 100
          # osDiskSKU  = "Standard_LRS"
          # Size of the data disk (optional)
          # dataDiskSize = 100
          # dataDiskSKU  = "Standard_LRS"
          tags = {
            "${var.cluster_name}-workers" = "pool1"
          }
        }
      }
    }
  }
}
