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

provider "vsphere" {
  /*
  See https://www.terraform.io/docs/providers/vsphere/index.html#argument-reference
  for config options reference
  */
}

locals {
  resource_pool_id = var.resource_pool_name == "" ? data.vsphere_compute_cluster.cluster.resource_pool_id : data.vsphere_resource_pool.pool[0].id
  hostnames        = formatlist("${var.cluster_name}-cp-%d", [1, 2, 3])

  cluster_autoscaler_min_replicas = var.cluster_autoscaler_min_replicas > 0 ? var.cluster_autoscaler_min_replicas : var.initial_machinedeployment_replicas
  cluster_autoscaler_max_replicas = var.cluster_autoscaler_max_replicas > 0 ? var.cluster_autoscaler_max_replicas : var.initial_machinedeployment_replicas
}

data "vsphere_datacenter" "dc" {
  name = var.dc_name
}

data "vsphere_datastore" "datastore" {
  name          = var.datastore_name
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_compute_cluster" "cluster" {
  name          = var.compute_cluster_name
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_resource_pool" "pool" {
  count         = var.resource_pool_name == "" ? 0 : 1
  name          = var.resource_pool_name
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_network" "network" {
  name          = var.network_name
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_virtual_machine" "template" {
  name          = var.template_name
  datacenter_id = data.vsphere_datacenter.dc.id
}

resource "vsphere_virtual_machine" "control_plane" {
  count = var.control_plane_vm_count

  name             = "${var.cluster_name}-cp-${count.index + 1}"
  resource_pool_id = local.resource_pool_id
  folder           = var.folder_name
  datastore_id     = data.vsphere_datastore.datastore.id
  num_cpus         = var.control_plane_num_cpus
  memory           = var.control_plane_memory
  guest_id         = data.vsphere_virtual_machine.template.guest_id
  scsi_type        = data.vsphere_virtual_machine.template.scsi_type
  firmware         = data.vsphere_virtual_machine.template.firmware

  network_interface {
    network_id   = data.vsphere_network.network.id
    adapter_type = data.vsphere_virtual_machine.template.network_interface_types[0]
  }

  disk {
    label            = "disk0"
    size             = var.disk_size
    thin_provisioned = data.vsphere_virtual_machine.template.disks[0].thin_provisioned
    eagerly_scrub    = data.vsphere_virtual_machine.template.disks[0].eagerly_scrub
  }

  cdrom {
    client_device = true
  }

  clone {
    template_uuid = data.vsphere_virtual_machine.template.id
  }

  extra_config = {
    "disk.enableUUID" = "TRUE"
  }

  vapp {
    properties = {
      hostname    = local.hostnames[count.index]
      public-keys = file(var.ssh_public_key_file)
    }
  }

  lifecycle {
    ignore_changes = [
      vapp[0].properties,
      tags,
    ]
  }

  connection {
    type         = "ssh"
    host         = self.default_ip_address
    user         = var.ssh_username
    bastion_host = var.bastion_host
    bastion_port = var.bastion_port
    bastion_user = var.bastion_username
  }

  provisioner "remote-exec" {
    script = "keepalived.sh"
  }
}

resource "random_string" "keepalived_auth_pass" {
  length  = 8
  special = false
}

resource "null_resource" "keepalived_config" {
  count = var.api_vip != "" ? 3 : 0

  triggers = {
    cluster_instance_ids = join(",", vsphere_virtual_machine.control_plane.*.id)
  }

  connection {
    type         = "ssh"
    user         = var.ssh_username
    host         = vsphere_virtual_machine.control_plane[count.index].default_ip_address
    bastion_host = var.bastion_host
    bastion_port = var.bastion_port
    bastion_user = var.bastion_username
  }

  provisioner "file" {
    content = templatefile("./etc_keepalived_keepalived_conf.tpl", {
      STATE         = count.index == 0 ? "MASTER" : "BACKUP",
      APISERVER_VIP = var.api_vip,
      INTERFACE     = var.vrrp_interface,
      ROUTER_ID     = var.vrrp_router_id,
      PRIORITY      = count.index == 0 ? "101" : "100",
      AUTH_PASS     = random_string.keepalived_auth_pass.result
    })
    destination = "/tmp/keepalived.conf"
  }

  provisioner "file" {
    content = templatefile("./etc_keepalived_check_apiserver_sh.tpl", {
      APISERVER_VIP = var.api_vip
    })
    destination = "/tmp/check_apiserver.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo mkdir -p /etc/keepalived",
      "sudo mv /tmp/keepalived.conf /etc/keepalived/keepalived.conf",
      "sudo mv /tmp/check_apiserver.sh /etc/keepalived/check_apiserver.sh",
      "sudo chmod +x /etc/keepalived/check_apiserver.sh",
      "sudo systemctl restart keepalived",
    ]
  }
}

/*
vSphere DRS requires a vSphere Enterprise Plus license. Toggle variable value off if you don't have it.
An anti-affinity rule places a control_plane machines across different hosts within a cluster, and is useful for preventing single points of failure.
*/

resource "vsphere_compute_cluster_vm_anti_affinity_rule" "vm_anti_affinity_rule" {
  count               = var.is_vsphere_enterprise_plus_license ? 1 : 0
  name                = "${var.cluster_name}-cp-vm-anti-affinity"
  compute_cluster_id  = data.vsphere_compute_cluster.cluster.id
  virtual_machine_ids = vsphere_virtual_machine.control_plane.*.id
}
