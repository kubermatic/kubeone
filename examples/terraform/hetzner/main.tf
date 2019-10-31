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

locals {
  cluster_name = random_pet.cluster_name.id
}

resource "random_pet" "cluster_name" {
  prefix = var.cluster_name
  length = 1
}

resource "hcloud_ssh_key" "kubeone" {
  name       = "kubeone-${local.cluster_name}"
  public_key = file(var.ssh_public_key_file)
}

resource "hcloud_network" "net" {
  name     = local.cluster_name
  ip_range = var.ip_range
}

resource "hcloud_network_subnet" "kubeone" {
  network_id   = hcloud_network.net.id
  type         = "server"
  network_zone = var.network_zone
  ip_range     = var.ip_range
}

resource "hcloud_server_network" "control_plane" {
  count      = 3
  server_id  = element(hcloud_server.control_plane.*.id, count.index)
  network_id = hcloud_network.net.id
}

resource "hcloud_server" "control_plane" {
  count       = 3
  name        = "${local.cluster_name}-cp-${count.index + 1}"
  server_type = var.control_plane_type
  image       = var.image
  location    = var.datacenter

  ssh_keys = [
    hcloud_ssh_key.kubeone.id,
  ]

  labels = {
    "kubeone_cluster_name" = local.cluster_name
    "role"                 = "api"
  }
}

resource "hcloud_server_network" "lb" {
  server_id  = hcloud_server.lb.id
  network_id = hcloud_network.net.id
}

resource "hcloud_server" "lb" {
  name        = "${local.cluster_name}-lb"
  server_type = var.lb_type
  image       = var.image
  location    = var.datacenter

  ssh_keys = [
    hcloud_ssh_key.kubeone.id,
  ]

  labels = {
    "kubeone_cluster_name" = local.cluster_name
    "role"                 = "lb"
  }

  connection {
    type = "ssh"
    host = self.ipv4_address
  }

  provisioner "remote-exec" {
    script = "gobetween.sh"
  }
}

locals {
  rendered_lb_config = templatefile("./etc_gobetween.tpl", {
    lb_targets = hcloud_server_network.control_plane.*.ip,
  })
}

resource "null_resource" "lb_config" {
  triggers = {
    cluster_instance_ids = join(",", hcloud_server_network.control_plane.*.ip)
    config               = local.rendered_lb_config
  }

  connection {
    host = hcloud_server.lb.ipv4_address
  }

  provisioner "file" {
    content     = local.rendered_lb_config
    destination = "/etc/gobetween.toml"
  }

  provisioner "remote-exec" {
    inline = [
      "systemctl restart gobetween",
    ]
  }
}

