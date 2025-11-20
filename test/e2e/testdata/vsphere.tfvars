allow_insecure       = true
dc_name              = "Hamburg"
datastore_name       = "Datastore0-truenas"
compute_cluster_name = "vSAN Cluster"
network_name         = "VM Network"
folder_name          = "KubeOne-E2E"
template_name        = "kubeone-e2e-ubuntu-24.04"
control_plane_memory = 4096
worker_memory        = 4096
disk_size            = 25
# We don't have permissions to create the required anti-affinity resource
# so we'll just disable this for now.
is_vsphere_enterprise_plus_license = false
