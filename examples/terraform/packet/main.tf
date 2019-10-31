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

provider "packet" {
}

locals {
  cluster_name     = random_pet.cluster_name.id
  kube_cluster_tag = "kubernetes-cluster:${local.cluster_name}"
}

resource "random_pet" "cluster_name" {
  prefix = var.cluster_name
  length = 1
}

resource "packet_ssh_key" "deployer" {
  name       = "terraform"
  public_key = file(var.ssh_public_key_file)
}

resource "packet_device" "control_plane" {
  count      = 3
  depends_on = [packet_ssh_key.deployer]

  hostname         = "${local.cluster_name}-cp-${count.index + 1}"
  plan             = var.device_type
  facilities       = [var.facility]
  operating_system = var.control_plane_operating_system
  billing_cycle    = "hourly"
  project_id       = var.project_id
  tags             = [local.kube_cluster_tag]
}

resource "packet_device" "lb" {
  depends_on = [packet_ssh_key.deployer]

  hostname         = "${local.cluster_name}-lb"
  plan             = "t1.small.x86"
  facilities       = [var.facility]
  operating_system = var.lb_operating_system
  billing_cycle    = "hourly"
  project_id       = var.project_id
  tags             = [local.kube_cluster_tag]

  connection {
    type = "ssh"
    host = self.access_public_ipv4
  }

  provisioner "remote-exec" {
    script = "gobetween.sh"
  }
}

locals {
  rendered_lb_config = templatefile("./etc_gobetween.tpl", {
    lb_targets = packet_device.control_plane.*.access_private_ipv4,
  })
}

resource "null_resource" "lb_config" {
  triggers = {
    cluster_instance_ids = join(",", packet_device.control_plane.*.id)
    config               = local.rendered_lb_config
  }

  connection {
    host = packet_device.lb.access_public_ipv4
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
