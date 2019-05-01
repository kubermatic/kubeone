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

provider "packet" {}

locals {
  kube_cluster_tag = "kubernetes-cluster:${var.cluster_name}"
}

resource "packet_ssh_key" "deployer" {
  name       = "terraform"
  public_key = "${file("${var.ssh_public_key_file}")}"
}

resource "packet_device" "control_plane" {
  count      = "${var.control_plane_count}"
  depends_on = ["packet_ssh_key.deployer"]

  hostname         = "${var.cluster_name}-control-plane-${count.index + 1}"
  plan             = "${var.device_type}"
  facilities       = ["${var.facility}"]
  operating_system = "${var.operating_system}"
  billing_cycle    = "hourly"
  project_id       = "${var.project_id}"
  tags             = ["${local.kube_cluster_tag}"]
}

resource "packet_device" "lb" {
  depends_on = ["packet_ssh_key.deployer"]

  hostname         = "${var.cluster_name}-lb"
  plan             = "t1.small.x86"
  facilities       = ["${var.facility}"]
  operating_system = "${var.operating_system}"
  billing_cycle    = "hourly"
  project_id       = "${var.project_id}"
  tags             = ["${local.kube_cluster_tag}"]

  connection {
    host = "${self.access_public_ipv4}"
  }

  provisioner "remote-exec" {
    script = "gobetween.sh"
  }
}

data "template_file" "lbconfig" {
  template = "${file("etc_gobetween.tpl")}"

  vars = {
    bind       = "${packet_device.lb.access_public_ipv4}"
    lb_target1 = "${packet_device.control_plane.0.access_private_ipv4}"
    lb_target2 = "${packet_device.control_plane.1.access_private_ipv4}"
    lb_target3 = "${packet_device.control_plane.2.access_private_ipv4}"
  }
}

resource "null_resource" "lb_config" {
  triggers = {
    cluster_instance_ids = "${join(",", packet_device.control_plane.*.id)}"
    config               = "${data.template_file.lbconfig.rendered}"
  }

  connection {
    host = "${packet_device.lb.access_public_ipv4}"
  }

  provisioner "file" {
    content     = "${data.template_file.lbconfig.rendered}"
    destination = "/etc/gobetween.toml"
  }

  provisioner "remote-exec" {
    inline = [
      "systemctl restart gobetween",
    ]
  }
}
