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

resource "hcloud_ssh_key" "kubeone" {
  name       = "kubeone-${var.cluster_name}"
  public_key = "${file("${var.ssh_public_key_file}")}"
}

resource "hcloud_server" "control_plane" {
  count       = "${var.control_plane_count}"
  name        = "${var.cluster_name}-control-plane-${count.index +1}"
  server_type = "cx31"
  image       = "centos-7"
  location    = "hel1"

  ssh_keys = [
    "${hcloud_ssh_key.kubeone.id}",
  ]

  labels = "${map("kubeone_cluster_name", "${var.cluster_name}")}"
}
