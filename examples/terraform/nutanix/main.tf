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

provider "nutanix" {}

locals {
  cluster_autoscaler_min_replicas = var.cluster_autoscaler_min_replicas > 0 ? var.cluster_autoscaler_min_replicas : var.initial_machinedeployment_replicas
  cluster_autoscaler_max_replicas = var.cluster_autoscaler_max_replicas > 0 ? var.cluster_autoscaler_max_replicas : var.initial_machinedeployment_replicas

  rendered_lb_config = templatefile("./etc_gobetween.tpl", {
    lb_targets = nutanix_virtual_machine.control_plane.*.nic_list.0.ip_endpoint_list.0.ip,
  })
}

data "nutanix_cluster" "cluster" {
  name = var.nutanix_cluster_name
}

data "nutanix_project" "project" {
  project_name = var.project_name
}

data "nutanix_subnet" "subnet" {
  subnet_name = var.subnet_name
}

data "nutanix_image" "image" {
  image_name = var.image_name
}

resource "nutanix_category_key" "category_key" {
  name        = "KubeOneCluster"
  description = "KubeOne Cluster category key"
}

resource "nutanix_category_value" "category_value" {
  name        = nutanix_category_key.category_key.id
  description = "KubeOne Cluster category value"
  value       = var.cluster_name
}

resource "nutanix_virtual_machine" "control_plane" {
  count        = var.control_plane_vm_count
  name         = "${var.cluster_name}-cp-${count.index}"
  cluster_uuid = data.nutanix_cluster.cluster.metadata.uuid
  project_reference = {
    kind = "project"
    uuid = data.nutanix_project.project.metadata.uuid
    name = var.project_name
  }

  num_vcpus_per_socket = var.control_plane_vcpus
  num_sockets          = var.control_plane_sockets
  memory_size_mib      = var.control_plane_memory_size

  nic_list {
    subnet_uuid = data.nutanix_subnet.subnet.metadata.uuid
  }

  disk_list {
    disk_size_mib = var.control_plane_disk_size

    data_source_reference = {
      kind = "image"
      uuid = data.nutanix_image.image.metadata.uuid
    }
  }

  guest_customization_cloud_init_user_data = base64encode(templatefile("./cloud-config.tftpl", {
    machine_name = "${var.cluster_name}-cp-${count.index}"
    ssh_key      = file(var.ssh_public_key_file)
  }))

  categories {
    name  = nutanix_category_key.category_key.name
    value = nutanix_category_value.category_value.value
  }
}

resource "nutanix_virtual_machine" "lb" {
  name         = "${var.cluster_name}-lb"
  cluster_uuid = data.nutanix_cluster.cluster.metadata.uuid
  project_reference = {
    kind = "project"
    uuid = data.nutanix_project.project.metadata.uuid
    name = var.project_name
  }

  num_vcpus_per_socket = var.bastion_vcpus
  num_sockets          = var.bastion_sockets
  memory_size_mib      = var.bastion_memory_size

  nic_list {
    subnet_uuid = data.nutanix_subnet.subnet.metadata.uuid
  }

  disk_list {
    disk_size_mib = var.bastion_disk_size

    data_source_reference = {
      kind = "image"
      uuid = data.nutanix_image.image.metadata.uuid
    }
  }

  guest_customization_cloud_init_user_data = base64encode(templatefile("./cloud-config.tftpl", {
    machine_name = "${var.cluster_name}-lb"
    ssh_key      = file(var.ssh_public_key_file)
  }))

  categories {
    name  = nutanix_category_key.category_key.name
    value = nutanix_category_value.category_value.value
  }

  connection {
    type = "ssh"
    host = nutanix_virtual_machine.lb.nic_list.0.ip_endpoint_list.0.ip
    user = var.ssh_username
  }

  provisioner "remote-exec" {
    script = "gobetween.sh"
  }
}

resource "null_resource" "lb_config" {
  triggers = {
    cluster_instance_ids = join(",", nutanix_virtual_machine.control_plane.*.metadata.uuid)
    config               = local.rendered_lb_config
  }

  depends_on = [
    nutanix_virtual_machine.lb
  ]

  connection {
    host = nutanix_virtual_machine.lb.nic_list.0.ip_endpoint_list.0.ip
    user = var.ssh_username
  }

  provisioner "file" {
    content     = local.rendered_lb_config
    destination = "/tmp/gobetween.toml"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo mv /tmp/gobetween.toml /etc/gobetween.toml",
      "sudo systemctl restart gobetween",
    ]
  }
}
