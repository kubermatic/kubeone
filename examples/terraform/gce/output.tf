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
  description = "kubernetes API loadbalancer"

  value = {
    endpoint = "${google_compute_address.lb_ip.address}"
  }
}

output "kubeone_hosts" {
  description = "control plain nodes"

  value = {
    control_plane = {
      cluster_name     = "${var.cluster_name}"
      cloud_provider   = "gce"
      private_address  = "${google_compute_instance.control_plane.*.network_interface.0.network_ip}"
      public_address   = "${google_compute_instance.control_plane.*.network_interface.0.access_config.0.nat_ip}"
      ssh_agent_socket = "${var.ssh_agent_socket}"
      ssh_port         = "${var.ssh_port}"
      ssh_user         = "${var.ssh_username}"
    }
  }
}

output "kubeone_workers" {
  description = "workers definitions translated into MachineDeployment ClusterAPI objects"

  value = {
    # following outputs will be parsed by kubeone and automatically merged into
    # corresponding (by name) worker definition
    workers1 = {
      replicas        = 1
      operatingSystem = "ubuntu"
      sshPublicKeys   = ["${file("${var.ssh_public_key_file}")}"]
      diskSize        = 50
      diskType        = "pd-ssd"
      machineType     = "${var.workers_type}"
      network         = "${google_compute_network.network.self_link}"
      subnetwork      = "${google_compute_subnetwork.subnet.self_link}"
      zone            = "${var.region}a"                                // hardcoded
    }
  }
}
