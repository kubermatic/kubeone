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
  address_prefix       = "172.16.1.0/24"
}

resource "azurerm_network_security_group" "sg" {
  name                = "${var.cluster_name}-sg"
  location            = var.location
  resource_group_name = azurerm_resource_group.rg.name

  security_rule {
    name                       = "SSH"
    priority                   = 1001
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_public_ip" "lbip" {
  name                = "${var.cluster_name}-lbip"
  location            = var.location
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Dynamic"

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_public_ip" "control_plane" {
  count = 3

  name                = "${var.cluster_name}-cp-${count.index}"
  location            = var.location
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Dynamic"

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_lb" "lb" {
  resource_group_name = azurerm_resource_group.rg.name
  name                = "${var.cluster_name}-lb"
  location            = var.location

  frontend_ip_configuration {
    name                 = "KubeApi"
    public_ip_address_id = azurerm_public_ip.lbip.id
  }

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

resource "azurerm_lb_backend_address_pool" "backend_pool" {
  resource_group_name = azurerm_resource_group.rg.name
  loadbalancer_id     = azurerm_lb.lb.id
  name                = "ApiServers"
}

resource "azurerm_lb_rule" "lb_rule" {
  resource_group_name            = azurerm_resource_group.rg.name
  loadbalancer_id                = azurerm_lb.lb.id
  name                           = "LBRule"
  protocol                       = "tcp"
  frontend_port                  = 6443
  backend_port                   = 6443
  frontend_ip_configuration_name = "KubeApi"
  enable_floating_ip             = false
  backend_address_pool_id        = azurerm_lb_backend_address_pool.backend_pool.id
  idle_timeout_in_minutes        = 5
  probe_id                       = azurerm_lb_probe.lb_probe.id
  depends_on                     = [azurerm_lb_probe.lb_probe]
}

resource "azurerm_lb_probe" "lb_probe" {
  resource_group_name = azurerm_resource_group.rg.name
  loadbalancer_id     = azurerm_lb.lb.id
  name                = "tcpProbe"
  protocol            = "tcp"
  port                = 6443
  interval_in_seconds = 5
  number_of_probes    = 2
}

resource "azurerm_network_interface" "control_plane" {
  count = 3

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
  count = 3

  ip_configuration_name   = "${var.cluster_name}-cp-${count.index}"
  network_interface_id    = element(azurerm_network_interface.control_plane.*.id, count.index)
  backend_address_pool_id = azurerm_lb_backend_address_pool.backend_pool.id
}

resource "azurerm_virtual_machine" "control_plane" {
  count = 3

  name                             = "${var.cluster_name}-cp-${count.index}"
  location                         = var.location
  resource_group_name              = azurerm_resource_group.rg.name
  availability_set_id              = azurerm_availability_set.avset.id
  vm_size                          = var.control_plane_vm_size
  network_interface_ids            = [element(azurerm_network_interface.control_plane.*.id, count.index)]
  delete_os_disk_on_termination    = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "${var.cluster_name}-cp-${count.index}"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "${var.cluster_name}-cp-${count.index}"
    admin_username = var.ssh_username
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      key_data = file(var.ssh_public_key_file)
      path     = "/home/${var.ssh_username}/.ssh/authorized_keys"
    }
  }

  tags = {
    environment = "kubeone"
    cluster     = var.cluster_name
  }
}

