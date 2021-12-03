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

provider "metal" {
}

locals {
  kube_cluster_tag = "kubernetes-cluster:${var.cluster_name}"
}

resource "metal_ssh_key" "deployer" {
  name       = "terraform"
  public_key = file(var.ssh_public_key_file)
}

resource "metal_device" "control_plane" {
  count      = 3
  depends_on = [metal_ssh_key.deployer]

  hostname         = "${var.cluster_name}-control-plane-${count.index + 1}"
  plan             = var.device_type
  facilities       = [var.facility]
  operating_system = var.control_plane_operating_system
  billing_cycle    = "hourly"
  project_id       = var.project_id
  tags             = [local.kube_cluster_tag]
}

resource "metal_device" "lb" {
  depends_on = [metal_ssh_key.deployer]

  hostname         = "${var.cluster_name}-lb"
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
    lb_targets = metal_device.control_plane.*.access_private_ipv4,
  })
}

resource "null_resource" "lb_config" {
  triggers = {
    cluster_instance_ids = join(",", metal_device.control_plane.*.id)
    config               = local.rendered_lb_config
  }

  connection {
    host = metal_device.lb.access_public_ipv4
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
