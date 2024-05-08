allow_insecure       = true
dc_name              = "Hamburg"
datastore_name       = "esxi-3"
compute_cluster_name = "Kubermatic"
network_name         = "Default Network"
folder_name          = "Kubermatic-ci"
control_plane_memory = 4096
worker_memory        = 4096
disk_size            = 25
# We don't have permissions to create the required anti-affinity resource
# so we'll just disable this for now.
is_vsphere_enterprise_plus_license = false
