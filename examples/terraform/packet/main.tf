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
