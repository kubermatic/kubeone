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
    endpoint                    = var.kubeapi_hostname != "" ? var.kubeapi_hostname : vcd_vapp_vm.control_plane.0.network.0.ip
    apiserver_alternative_names = var.apiserver_alternative_names
  }
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name         = var.cluster_name
      cloud_provider       = "vmwareCloudDirector"
      private_address      = vcd_vapp_vm.control_plane.*.network.0.ip
      hostnames            = vcd_vapp_vm.control_plane.*.name
      ssh_agent_socket     = var.ssh_agent_socket
      ssh_port             = var.ssh_port
      ssh_private_key_file = var.ssh_private_key_file
      ssh_user             = var.ssh_username
      vapp_name            = vcd_vapp.cluster.name
      storage_profile      = var.worker_disk_storage_profile
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
        annotations = {
          "k8c.io/operating-system-profile" = var.initial_machinedeployment_operating_system_profile
        }
        sshPublicKeys   = [file(var.ssh_public_key_file)]
        operatingSystem = var.worker_os
        operatingSystemSpec = {
          distUpgradeOnBoot = false
        }
        # nodeAnnotations are applied on resulting Node objects
        # nodeAnnotations = {
        #   "key" = "value"
        # }
        # machineObjectAnnotations are applied on resulting Machine objects
        # uncomment to following to set those kubelet parameters. More into at:
        # https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/
        # machineObjectAnnotations = {
        #   "v1.kubelet-config.machine-controller.kubermatic.io/SystemReserved" = "cpu=200m,memory=200Mi"
        #   "v1.kubelet-config.machine-controller.kubermatic.io/KubeReserved"   = "cpu=200m,memory=300Mi"
        #   "v1.kubelet-config.machine-controller.kubermatic.io/EvictionHard"   = ""
        #   "v1.kubelet-config.machine-controller.kubermatic.io/MaxPods"        = "110"
        # }
        cloudProviderSpec = {
          # provider specific fields:
          # see example under `cloudProviderSpec` section at:
          # https://github.com/kubermatic/machine-controller/blob/master/examples/vmware-cloud-director-machinedeployment.yaml
          organization     = var.vcd_org_name
          vdc              = var.vcd_vdc_name
          vapp             = vcd_vapp.cluster.name
          catalog          = var.catalog_name
          template         = var.template_name
          network          = vcd_vapp_org_network.network.org_network_name
          cpus             = var.worker_cpus
          cpuCores         = var.worker_cpu_cores
          memoryMB         = var.worker_memory
          diskSizeGB       = var.worker_disk_size_gb
          storageProfile   = var.worker_disk_storage_profile
          ipAllocationMode = "DHCP"
          metadata = {
            "KubeOneCluster" = var.cluster_name
          }
        }
      }
    }
  }
}
