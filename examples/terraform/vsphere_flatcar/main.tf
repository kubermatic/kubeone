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
      "guestinfo.ignition.config.data.encoding" = "base64"
      "guestinfo.ignition.config.data" = base64encode(jsonencode({
        ignition = {
          version = "2.2.0"
        }
        systemd = {
          units = [
            {
              name    = "docker.socket"
              enabled = false
            },
            {
              name    = "docker.service"
              enabled = true
            }
          ]
        },
        storage = {
          files = [
            {
              filesystem = "root"
              path       = "/etc/hostname"
              mode       = 420
              contents = {
                source = "data:,${local.hostnames[count.index]}"
              }
            }
          ]
        },
        passwd = {
          users = [
            {
              name              = "core"
              sshAuthorizedKeys = [file(var.ssh_public_key_file)]
            }
          ]
        }
      }))
    }
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
