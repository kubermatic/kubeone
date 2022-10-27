disable_kubeapi_loadbalancer = true
subnets_cidr                 = 27

# Use smaller instances in Ireland for E2E tests
aws_region                = "eu-west-1"
control_plane_type        = "t3a.medium"
control_plane_volume_size = 25
worker_type               = "t3a.medium"
bastion_type              = "t3a.nano"
