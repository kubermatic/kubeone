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

output "kubeone_api" {
  value = {
    endpoint = "${packet_device.lb.access_public_ipv4}"
  }
}

output "kubeone_hosts" {
  value = {
    control_plane = {
      cluster_name         = "${var.cluster_name}"
      cloud_provider       = "packet"
      private_address      = "${packet_device.control_plane.*.access_private_ipv4}"
      public_address       = "${packet_device.control_plane.*.access_public_ipv4}"
      ssh_agent_socket     = "${var.ssh_agent_socket}"
      ssh_port             = "${var.ssh_port}"
      ssh_private_key_file = "${var.ssh_private_key_file}"
      ssh_user             = "${var.ssh_username}"
    }
  }
}

output "kubeone_workers" {
  value = {
    # following outputs will be parsed by kubeone and automatically merged into
    # corresponding (by name) worker definition
    pool1 = {
      replicas        = 1
      sshPublicKeys   = ["${file("${var.ssh_public_key_file}")}"]
      operatingSystem = "${var.workers_operating_system}"

      operatingSystemSpec = {
        distUpgradeOnBoot = true
      }

      projectID    = "${var.project_id}"
      facilities   = ["${var.facility}"]
      instanceType = "${var.device_type}"
    }
  }
}
