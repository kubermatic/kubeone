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
  description = "kube-apiserver LB endpoint"

  value = {
    endpoint = google_compute_address.lb_ip.address
  }
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name         = var.cluster_name
      cloud_provider       = "gce"
      private_address      = google_compute_instance.control_plane.*.network_interface.0.network_ip
      public_address       = google_compute_instance.control_plane.*.network_interface.0.access_config.0.nat_ip
      hostnames            = google_compute_instance.control_plane.*.name
      ssh_agent_socket     = var.ssh_agent_socket
      ssh_port             = var.ssh_port
      ssh_private_key_file = var.ssh_private_key_file
      ssh_user             = var.ssh_username
    }
  }
}

output "kubeone_workers" {
  description = "Workers definitions, that will be transformed into MachineDeployment object"

  value = {
    # following outputs will be parsed by kubeone and automatically merged into
    # corresponding (by name) worker definition
    "${var.cluster_name}-pool1" = {
      replicas = 1
      providerSpec = {
        sshPublicKeys   = [file(var.ssh_public_key_file)]
        operatingSystem = var.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot = false
        }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at:
          # https://github.com/kubermatic/machine-controller/blob/master/examples/gce-machinedeployment.yaml
          diskSize              = 50
          diskType              = "pd-ssd"
          machineType           = var.workers_type
          network               = google_compute_network.network.self_link
          subnetwork            = google_compute_subnetwork.subnet.self_link
          zone                  = "${var.region}-a"
          preemptible           = false
          assignPublicIPAddress = true
          # Enable support for multizone clusters
          multizone = true
          labels = {
            "${var.cluster_name}-workers" = "pool1"
          }
          tags     = ["firewall", "targets", "${var.cluster_name}-pool1"]
          regional = false
          # Use custom image (optional)
          # customImage = ""
          # Red Hat subscription manager offline token (only to be used for RHEL)
          # rhsmOfflineToken = ""
        }
      }
    }
  }
}

