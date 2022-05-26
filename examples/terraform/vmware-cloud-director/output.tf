/*
Copyright 2022 The KubeOne Authors.

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
    endpoint                    = local.public_ip
    apiserver_alternative_names = var.apiserver_alternative_names
  }
}

output "ssh_commands" {
  value = formatlist("ssh -J ${var.bastion_user}@${local.public_ip} ${var.ssh_username}@%s", vcd_vapp_vm.control_plane.*.network.0.ip)
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name         = var.cluster_name
      cloud_provider       = "vmware-cloud-director"
      private_address      = vcd_vapp_vm.control_plane.*.network.0.ip
      hostnames            = vcd_vapp_vm.control_plane.*.name
      ssh_agent_socket     = var.ssh_agent_socket
      ssh_port             = var.ssh_port
      ssh_private_key_file = var.ssh_private_key_file
      ssh_user             = var.ssh_username
      bastion              = local.public_ip
      bastion_port         = var.bastion_port
      bastion_user         = var.bastion_user
    }
  }
}

output "kubeone_workers" {
  description = "Workers definitions, that will be transformed into MachineDeployment object"

  value = {
    # following outputs will be parsed by kubeone and automatically merged into
    # corresponding (by name) worker definition
    "${var.cluster_name}-pool1" = {
      replicas = var.initial_machinedeployment_replicas
      providerSpec = {
        sshPublicKeys   = [file(var.ssh_public_key_file)]
        operatingSystem = var.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot = false
        }
        # uncomment to following to set those kubelet parameters. More into at:
        # https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/
        # machineAnnotations = {
        #  "v1.kubelet-config.machine-controller.kubermatic.io/SystemReserved" = "cpu=200m,memory=200Mi"
        #  "v1.kubelet-config.machine-controller.kubermatic.io/KubeReserved"   = "cpu=200m,memory=300Mi"
        #  "v1.kubelet-config.machine-controller.kubermatic.io/EvictionHard"   = ""
        # }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at:
          # https://github.com/kubermatic/machine-controller/blob/master/examples/nutanix-machinedeployment.yaml
          organization = var.vcd_org_name
          vdc          = var.vcd_vdc_name
          vapp         = vcd_vapp.cluster.name
          catalog      = var.catalog_name
          template     = var.template_name
          network      = vcd_vapp_org_network.network.org_network_name
          cpus         = var.worker_cpus
          cpuCores     = var.worker_cpu_cores
          memoryMB     = var.worker_memory
          diskSizeGB   = var.worker_disk_size
          storageProfile = var.worker_disk_storage_profile
          metadata = {
            "KubeOneCluster" = var.cluster_name
          }
        }
      }
    }
  }
}