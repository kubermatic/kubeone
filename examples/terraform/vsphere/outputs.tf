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
    endpoint                    = var.api_vip != "" ? var.api_vip : vsphere_virtual_machine.control_plane[0].default_ip_address
    apiserver_alternative_names = var.apiserver_alternative_names
  }
}

output "kubeone_hosts" {
  description = "Control plane endpoints to SSH to"

  value = {
    control_plane = {
      cluster_name    = var.cluster_name
      cloud_provider  = "vsphere"
      private_address = []
      hostnames       = local.hostnames
      public_address  = vsphere_virtual_machine.control_plane.*.guest_ip_addresses.0
      # KubeOne expects an array of array for IPv6 addresses since a single host/node can have multiple IPv6 addresses.
      ipv6_addresses       = var.ip_family == "IPv4+IPv6" ? [for ip in vsphere_virtual_machine.control_plane.*.guest_ip_addresses.1 : [ip]] : null
      ssh_agent_socket     = var.ssh_agent_socket
      ssh_port             = var.ssh_port
      ssh_private_key_file = var.ssh_private_key_file
      ssh_user             = var.ssh_username
      bastion              = var.bastion_host
      bastion_port         = var.bastion_port
      bastion_user         = var.bastion_username
      ssh_hosts_keys       = var.ssh_hosts_keys
      bastion_host_key     = var.bastion_host_key
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
          "k8c.io/operating-system-profile"                           = var.initial_machinedeployment_operating_system_profile
          "cluster.k8s.io/cluster-api-autoscaler-node-group-min-size" = tostring(local.cluster_autoscaler_min_replicas)
          "cluster.k8s.io/cluster-api-autoscaler-node-group-max-size" = tostring(local.cluster_autoscaler_max_replicas)
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
          # https://github.com/kubermatic/machine-controller/blob/main/examples/vsphere-machinedeployment.yaml
          allowInsecure = var.allow_insecure
          cluster       = var.compute_cluster_name
          cpus          = var.worker_num_cpus
          datacenter    = var.dc_name
          # Either Datastore or DatastoreCluster have to be provided.
          datastore        = var.datastore_name
          datastoreCluster = var.datastore_cluster_name
          # Optional: Resize the root disk to this size. Must be bigger than the existing size
          # Default is to leave the disk at the same size as the template
          diskSizeGB     = var.worker_disk
          memoryMB       = var.worker_memory
          templateVMName = var.template_name
          vmNetName      = var.network_name
          resourcePool   = var.resource_pool_name
          folder         = var.folder_name
        }
        network = {
          ipFamily = var.ip_family
        }
      }
    }
  }
}
