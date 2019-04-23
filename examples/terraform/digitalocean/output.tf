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
    endpoint = "${digitalocean_loadbalancer.control_plane.ip}"
  }
}

output "kubeone_hosts" {
  value = {
    control_plane = {
      cluster_name         = "${var.cluster_name}"
      cloud_provider       = "digitalocean"
      private_address      = "${digitalocean_droplet.control_plane.*.ipv4_address_private}"
      public_address       = "${digitalocean_droplet.control_plane.*.ipv4_address}"
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
    fra1-1 = {
      replicas        = 3
      sshPublicKeys   = ["${digitalocean_ssh_key.deployer.public_key}"]
      operatingSystem = "ubuntu"

      operatingSystemSpec = {
        distUpgradeOnBoot = false
      }

      region             = "${var.region}"
      size               = "${var.droplet_size}"
      private_networking = true
      backups            = false
      ipv6               = false
      monitoring         = false
    }
  }
}
