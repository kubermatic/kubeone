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

provider "azurerm" {
  features {}
  skip_provider_registration = true
}

locals {
  kubeapi_endpoint                   = var.disable_kubeapi_loadbalancer ? azurerm_network_interface.control_plane.0.private_ip_address : azurerm_public_ip.lbip.0.ip_address
  loadbalancer_count                 = var.disable_kubeapi_loadbalancer ? 0 : 1
  nic_address_pool_association_count = local.loadbalancer_count > 0 ? var.control_plane_vm_count : 0
  worker_os                          = var.worker_os == "" ? var.image_references[var.os].worker_os : var.worker_os
  ssh_username                       = var.ssh_username == "" ? var.image_references[var.os].ssh_username : var.ssh_username

  cluster_autoscaler_min_replicas = var.cluster_autoscaler_min_replicas > 0 ? var.cluster_autoscaler_min_replicas : var.initial_machinedeployment_replicas
  cluster_autoscaler_max_replicas = var.cluster_autoscaler_max_replicas > 0 ? var.cluster_autoscaler_max_replicas : var.initial_machinedeployment_replicas
}

provider "time" {
}

resource "azurerm_resource_group" "rg" {
  name     = "${var.cluster_name}-rg"
  location = var.location

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_availability_set" "avset" {
  name                         = "${var.cluster_name}-avset"
  location                     = var.location
  resource_group_name          = azurerm_resource_group.rg.name
  platform_fault_domain_count  = 2
  platform_update_domain_count = 2
  managed                      = true

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_route_table" "rt" {
  name                          = "${var.cluster_name}-rt"
  location                      = azurerm_resource_group.rg.location
  resource_group_name           = azurerm_resource_group.rg.name
  disable_bgp_route_propagation = false

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_virtual_network" "vpc" {
  name                = "${var.cluster_name}-vpc"
  address_space       = ["172.16.0.0/12"]
  location            = var.location
  resource_group_name = azurerm_resource_group.rg.name

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_subnet" "subnet" {
  name                 = "${var.cluster_name}-subnet"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vpc.name
  address_prefixes     = ["172.16.1.0/24"]
}

resource "azurerm_network_security_group" "sg" {
  name                = "${var.cluster_name}-sg"
  location            = var.location
  resource_group_name = azurerm_resource_group.rg.name

  security_rule {
    name                       = "SSH"
    description                = "Allow inbound SSH"
    priority                   = 1001
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "NodePorts"
    description                = "Allow inbound NodePorts"
    priority                   = 1010
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "30000-32767"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "kubeapi"
    description                = "Allow inbound kubeAPI"
    priority                   = 1020
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "6443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_subnet_network_security_group_association" "sg_subnet" {
  subnet_id                 = azurerm_subnet.subnet.id
  network_security_group_id = azurerm_network_security_group.sg.id
}

resource "azurerm_public_ip" "lbip" {
  count = local.loadbalancer_count

  name                = "${var.cluster_name}-lbip"
  location            = var.location
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Static"
  sku                 = "Standard"

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_public_ip" "control_plane" {
  count = var.control_plane_vm_count

  name                = "${var.cluster_name}-cp-${count.index}"
  location            = var.location
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Static"
  sku                 = "Standard"

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_lb" "lb" {
  count = local.loadbalancer_count

  resource_group_name = azurerm_resource_group.rg.name
  name                = "kubernetes"
  location            = var.location
  sku                 = "Standard"

  frontend_ip_configuration {
    name                 = "KubeApi"
    public_ip_address_id = azurerm_public_ip.lbip.0.id
  }

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_lb_backend_address_pool" "backend_pool" {
  count = local.loadbalancer_count

  loadbalancer_id = azurerm_lb.lb.0.id
  name            = "ApiServers"
}

resource "azurerm_lb_rule" "lb_rule" {
  count = local.loadbalancer_count

  loadbalancer_id                = azurerm_lb.lb.0.id
  name                           = "LBRule"
  protocol                       = "Tcp"
  frontend_port                  = 6443
  backend_port                   = 6443
  frontend_ip_configuration_name = "KubeApi"
  enable_floating_ip             = false
  backend_address_pool_ids       = [azurerm_lb_backend_address_pool.backend_pool.0.id]
  idle_timeout_in_minutes        = 5
  probe_id                       = azurerm_lb_probe.lb_probe.0.id
  depends_on                     = [azurerm_lb_probe.lb_probe.0]
}

resource "azurerm_lb_probe" "lb_probe" {
  count = local.loadbalancer_count

  loadbalancer_id     = azurerm_lb.lb.0.id
  name                = "tcpProbe"
  protocol            = "Tcp"
  port                = 6443
  interval_in_seconds = 5
  number_of_probes    = 2
}

resource "azurerm_network_interface" "control_plane" {
  count = var.control_plane_vm_count

  name                = "${var.cluster_name}-cp-${count.index}"
  location            = var.location
  resource_group_name = azurerm_resource_group.rg.name

  ip_configuration {
    name                          = "${var.cluster_name}-cp-${count.index}"
    subnet_id                     = azurerm_subnet.subnet.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = element(azurerm_public_ip.control_plane.*.id, count.index)
  }
}

resource "azurerm_network_interface_backend_address_pool_association" "control_plane" {
  count = local.nic_address_pool_association_count

  ip_configuration_name   = "${var.cluster_name}-cp-${count.index}"
  network_interface_id    = element(azurerm_network_interface.control_plane.*.id, count.index)
  backend_address_pool_id = azurerm_lb_backend_address_pool.backend_pool.0.id
}

resource "azurerm_virtual_machine" "control_plane" {
  count = var.control_plane_vm_count

  name                             = "${var.cluster_name}-cp-${count.index}"
  location                         = var.location
  resource_group_name              = azurerm_resource_group.rg.name
  availability_set_id              = azurerm_availability_set.avset.id
  vm_size                          = var.control_plane_vm_size
  network_interface_ids            = [element(azurerm_network_interface.control_plane.*.id, count.index)]
  delete_os_disk_on_termination    = true
  delete_data_disks_on_termination = true

  dynamic "plan" {
    for_each = var.image_references[var.os].plan

    content {
      name      = plan.value["name"]
      publisher = plan.value["publisher"]
      product   = plan.value["product"]
    }
  }

  storage_image_reference {
    publisher = var.image_references[var.os].image.publisher
    offer     = var.image_references[var.os].image.offer
    sku       = var.image_references[var.os].image.sku
    version   = var.image_references[var.os].image.version
  }

  storage_os_disk {
    name              = "${var.cluster_name}-cp-${count.index}"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "${var.cluster_name}-cp-${count.index}"
    admin_username = local.ssh_username
    custom_data = templatefile("./cloud-config.tftpl", {
      os            = var.os
      rhsm_username = var.rhsm_username
      rhsm_password = var.rhsm_password
    })
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      key_data = file(var.ssh_public_key_file)
      path     = "/home/${local.ssh_username}/.ssh/authorized_keys"
    }
  }

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

# Hack to ensure we get access to public ip in first attempt
resource "time_sleep" "wait_30_seconds" {
  depends_on      = [azurerm_virtual_machine.control_plane]
  create_duration = "30s"
}

data "azurerm_public_ip" "control_plane" {
  depends_on = [
    time_sleep.wait_30_seconds
  ]
  count               = var.control_plane_vm_count
  name                = "${var.cluster_name}-cp-${count.index}"
  resource_group_name = azurerm_resource_group.rg.name
}
