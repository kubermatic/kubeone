allow_insecure         = true
dc_name                = "Hamburg"
datastore_name         = "vsan"
compute_cluster_name   = "Kubermatic"
network_name           = "Default Network"
folder_name            = "Kubermatic-ci"
control_plane_memory   = 8192
worker_memory          = 8192
disk_size              = 25
control_plane_num_cpus = 4
worker_num_cpus        = 4
# We don't have permissions to create the required anti-affinity resource
# so we'll just disable this for now.
is_vsphere_enterprise_plus_license = false
