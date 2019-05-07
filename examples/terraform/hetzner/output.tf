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
    endpoint = "${hcloud_server.lb.ipv4_address}"
  }
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name   = "${var.cluster_name}"
      public_address = "${hcloud_server.control_plane.*.ipv4_address}"
    }
  }
}

output "kubeone_workers" {
  description = "Workers definitions, that will be transformed into MachineDeployment object"

  value = {
    pool1 = {
      serverType      = "${var.worker_type}"
      location        = "${var.datacenter}"
      replicas        = 3
      sshPublicKeys   = ["${file("${var.ssh_public_key_file}")}"]
      operatingSystem = "ubuntu"

      operatingSystemSpec = {
        distUpgradeOnBoot = true
      }
    }
  }
}
