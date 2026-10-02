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

output "networking" {
  description = "Networking resources to reference in controlPlane.nodeSets[].cloudProviderSpec"

  value = {
    location          = azurerm_resource_group.rg.location
    resourceGroup     = azurerm_resource_group.rg.name
    vnetName          = azurerm_virtual_network.vpc.name
    subnetName        = azurerm_subnet.subnet.name
    securityGroupName = azurerm_network_security_group.sg.name
    routeTableName    = azurerm_route_table.rt.name
    availabilitySet   = azurerm_availability_set.avset.name
    loadBalancerSku   = "Standard"
  }
}

output "vm" {
  description = "Image reference to use for control plane machines in controlPlane.nodeSets[].cloudProviderSpec"

  value = {
    imageReference = var.os != "rhel" ? var.image_references[var.os].image : null
    imagePlan      = length(var.image_references[var.os].plan) > 0 && var.os != "rhel" ? var.image_references[var.os].plan[0] : null
  }
}
