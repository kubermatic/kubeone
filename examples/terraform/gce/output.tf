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

output "kubeone" {
  value = yamlencode({
    apiVersion : "kubeone.io/v1beta1"
    kind : "KubeOneCluster"
    name : "k1"
    versions : {
      kubernetes : "1.21.0"
    }
    controlPlane : {
      hosts : [for host in google_compute_instance.control_plane :
        {
          publicAddress : host.network_interface.0.access_config.0.nat_ip
          privateAddress : host.network_interface.0.network_ip
          sshUsername : "root"
        }
      ]
    }
    staticWorkers : {
      hosts : [for host in google_compute_instance.static_workers :
        {
          publicAddress : host.network_interface.0.access_config.0.nat_ip
          privateAddress : host.network_interface.0.network_ip
          sshUsername : "root"
        }
      ]
    }
    cloudProvider : {
      gce : {}
    }

  })
}

output "master_machine_deployment" {
  value = yamlencode({
    apiVersion : "cluster.k8s.io/v1alpha1"
    kind : "MachineDeployment"
    metadata : {
      name : "kkp-master-pool-az-a"
      namespace : "kube-system"
      annotations : {
        "machinedeployment.clusters.k8s.io/revision" : "1"
        "cluster.k8s.io/cluster-api-autoscaler-node-group-min-size" : "1"
        "cluster.k8s.io/cluster-api-autoscaler-node-group-max-size" : "5"
      }
    }
    spec : {
      minReadySeconds : 0
      progressDeadlineSeconds : 600
      replicas : 1
      revisionHistoryLimit : 1
      selector : {
        matchLabels : {
          workerset : "kkp-master-pool-az-a"
        }
      }
      strategy : {
        rollingUpdate : {
          maxSurge : 3
          maxUnavailable : 0
        }
        type : "RollingUpdate"
      }
      template : {
        metadata : {
          creationTimestamp : null
          labels : {
            workerset : "kkp-master-pool-az-a"
          }
          namespace : "kube-system"
        }
        spec : {
          metadata : {
            creationTimestamp : null
            labels : {
              workerset : "kkp-master-pool-az-a"
            }
          }
          providerSpec : {
            value : {
              cloudProvider : "gce"
              cloudProviderSpec : {
                assignPublicIPAddress : true
                diskSize : 20
                diskType : "pd-ssd"
                labels : {
                  kkp-master-workers : "pool-az-a"
                }
                machineType : "n1-standard-2"
                multizone : true
                network : google_compute_network.network.self_link
                preemptible : false
                regional : false
                subnetwork : google_compute_subnetwork.subnet.self_link
                tags : [
                  "firewall"
                  , "targets"
                  , "kkp-master-pool-az-a"
                ]
                zone : data.google_compute_zones.available.names[0]
              }
              operatingSystem : "ubuntu"
              operatingSystemSpec : {
                distUpgradeOnBoot : false
              }
            }
          }
          versions : {
            kubelet : "1.21.0"
          }
        }
      }
    }
  })
}
